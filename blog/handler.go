package blog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
	blogStructs "github.com/yekta/banano-price-service/blog/structs"
)

func BlogHandler(c *fiber.Ctx, MEDIUM_SECRET string, MEDIUM_USER_ID string, GHOST_TO_MEDIUM_SECRET string, TYPESENSE_ADMIN_API_KEY string, GHOST_API_KEY string) error {
	key := c.Query("key")
	if key != GHOST_TO_MEDIUM_SECRET {
		log.Println("BlogHandler: Not authorized")
		return c.Status(http.StatusUnauthorized).SendString("Not authorized")
	}
	log.Println("BlogHandler triggered...")

	var payload blogStructs.SGhostPostWebhook
	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	post := payload.Post.Current

	mediumPostEndpoint := "https://api.medium.com/v1/users/" + MEDIUM_USER_ID + "/posts"
	var tags []string
	for _, tag := range post.Tags {
		tags = append(tags, tag.Name)
	}
	mediumPost := blogStructs.SMediumPost{
		Title:         post.Title,
		ContentFormat: "html",
		Content:       GhostToMediumHtmlConverter(post.Html, post.Title),
		PublishStatus: "draft",
		CanonicalUrl:  "https://banano.cc/blog/" + post.Slug,
		Tags:          tags,
	}

	mediumPostJson, err := json.Marshal(mediumPost)
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", mediumPostEndpoint, bytes.NewBuffer(mediumPostJson))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", "Bearer "+MEDIUM_SECRET)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}
	json.NewDecoder(resp.Body).Decode(&resp)

	r := blogStructs.SBlogResponse{
		Data: blogStructs.SBlogResponseData{
			Title: post.Title,
		},
	}
	log.Printf(`Submitted the post to Medium with the title "%s"...`, post.Title)

	// Typesense stuff starts here
	fields := [...]string{
		"id",
		"title",
		"slug",
		"published_at",
		"excerpt",
		"custom_excerpt",
		"feature_image",
	}
	formats := [...]string{"mobiledoc", "html", "plaintext"}
	include := [...]string{"tags"}
	blogApiUrl := "https://ghost.banano.cc/ghost/api/content"
	fieldsStr := strings.Join(fields[:], ",")
	formatsStr := strings.Join(formats[:], ",")
	includeStr := strings.Join(include[:], ",")
	blogEndpoint := fmt.Sprintf(`%s/posts?key=%s&fields=%s&formats=%s&include=%s`, blogApiUrl, GHOST_API_KEY, fieldsStr, formatsStr, includeStr)

	resp, err = http.Get(blogEndpoint)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}
	var ghostPosts blogStructs.SGhostPostsRepsonse
	json.NewDecoder(resp.Body).Decode(&ghostPosts)

	log.Println("BlogHandler: Got Ghost blog posts...")

	blogPostsForTypesense := []interface{}{
		struct {
			Id            string                      `json:"id"`
			Title         string                      `json:"title"`
			Slug          string                      `json:"slug"`
			PublishedAt   string                      `json:"publishedAt"`
			Excerpt       string                      `json:"excerpt"`
			CustomExcerpt string                      `json:"customExcerpt"`
			FeatureImage  string                      `json:"featureImage"`
			Tags          []blogStructs.SGhostPostTag `json:"tags"`
		}{},
	}

	for _, blogPost := range ghostPosts.Posts {
		blogPostsForTypesense = append(blogPostsForTypesense, blogStructs.SBlogPostForTypesense{
			Id:            blogPost.Id,
			Title:         blogPost.Title,
			Slug:          blogPost.Slug,
			PublishedAt:   blogPost.PublishedAt,
			Excerpt:       blogPost.Excerpt,
			CustomExcerpt: blogPost.CustomExcerpt,
			FeatureImage:  blogPost.FeatureImage,
			Tags:          blogPost.Tags,
		})
	}

	typesenseClient := typesense.NewClient(
		typesense.WithServer("https://typesense.banano.cc"),
		typesense.WithAPIKey(TYPESENSE_ADMIN_API_KEY))
	typesenseClient.Collection("blog-posts").Delete()
	log.Println("BlogHandler: Typesense collection deleted...")

	schema := &api.CollectionSchema{
		Name: "blog-posts",
		Fields: []api.Field{
			{
				Name: "id",
				Type: "string",
			},
			{
				Name: "title",
				Type: "string",
			},
			{
				Name: "slug",
				Type: "string",
			},
			{
				Name: "published_at",
				Type: "int64",
			},
			{
				Name: "excerpt",
				Type: "string",
			},
			{
				Name: "custom_excerpt",
				Type: "string",
			},
			{
				Name: "feature_image",
				Type: "string",
			},
			{
				Name:  "tags",
				Type:  "string[]",
				Facet: newTrue(),
			},
		},
		DefaultSortingField: defaultSortingField(),
	}
	typesenseClient.Collections().Create(schema)
	log.Println("BlogHandler: New Typesense collection created...")

	params := &api.ImportDocumentsParams{
		Action:    action(),
		BatchSize: batchSize(),
	}
	res, err := typesenseClient.Collection("blog-posts").Documents().Import(blogPostsForTypesense, params)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}
	log.Printf("BlogHandler: Imported %v documents to Typesense...", res)

	return c.JSON(r)
}

func GhostToMediumHtmlConverter(html string, title string) string {
	resHtml := fmt.Sprintf(`<h1>%s</h1>%s`, title, html)
	return resHtml
}

func newTrue() *bool {
	b := true
	return &b
}

func defaultSortingField() *string {
	f := "published_at"
	return &f
}

func batchSize() *int {
	f := 50
	return &f
}

func action() *string {
	f := "create"
	return &f
}
