package blog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
	blogStructs "github.com/yekta/banano-price-service/blog/structs"
	sharedUtils "github.com/yekta/banano-price-service/shared"
)

var TYPESENSE_ADMIN_API_KEY = sharedUtils.GetEnv("TYPESENSE_ADMIN_API_KEY")
var GHOST_API_KEY = sharedUtils.GetEnv("GHOST_API_KEY")
var GHOST_TO_MEDIUM_SECRET = sharedUtils.GetEnv("GHOST_TO_MEDIUM_SECRET")
var MEDIUM_SECRET = sharedUtils.GetEnv("MEDIUM_SECRET")
var MEDIUM_USER_ID = sharedUtils.GetEnv("MEDIUM_USER_ID")

func GhostToMediumHandler(c *fiber.Ctx) error {
	key := c.Query("key")
	if key != GHOST_TO_MEDIUM_SECRET {
		log.Println("GhostToMediumHandler: Not authorized")
		return c.Status(http.StatusUnauthorized).SendString("Not authorized")
	}
	log.Println("GhostToMediumHandler triggered...")

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

	// Typesense stuff
	IndexTypesense()

	return c.JSON(r)
}

func IndexTypesense() error {
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
	const limit = 1000
	blogEndpoint := fmt.Sprintf(`%s/posts?key=%s&fields=%s&formats=%s&include=%s&limit=%v`, blogApiUrl, GHOST_API_KEY, fieldsStr, formatsStr, includeStr, limit)

	resp, err := http.Get(blogEndpoint)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}
	var ghostPosts blogStructs.SGhostPostsRepsonse
	json.NewDecoder(resp.Body).Decode(&ghostPosts)

	log.Println("TypesenseHandler: Got Ghost blog posts...")

	var blogPostsForTypesense []interface{}
	for _, post := range ghostPosts.Posts {
		t, _ := time.Parse("2006-01-02T15:04:05.000+15:04", post.PublishedAt)
		blogPostsForTypesense = append(blogPostsForTypesense, blogStructs.SBlogPostForTypesense{
			Id:            post.Id,
			Title:         post.Title,
			Slug:          post.Slug,
			CustomExcerpt: post.CustomExcerpt,
			Excerpt:       post.Excerpt,
			PlainText:     post.Plaintext,
			FeatureImage:  post.FeatureImage,
			PublishedAt:   uint64(t.UnixMilli()),
		})
	}

	typesenseClient := typesense.NewClient(
		typesense.WithServer("https://typesense.banano.cc"),
		typesense.WithAPIKey(TYPESENSE_ADMIN_API_KEY),
		typesense.WithConnectionTimeout(60*time.Second))

	_, errDel := typesenseClient.Collection("blog-posts").Delete()

	if errDel != nil {
		log.Println("TypesenseHandler: Error deleting collection:", errDel)
	} else {
		log.Println("TypesenseHandler: Typesense collection deleted...")
	}

	schema := &api.CollectionSchema{
		Name: "blog-posts",
		Fields: []api.Field{
			{
				Name:  "title",
				Type:  "string",
				Infix: newTrue(),
			},
			{
				Name: "excerpt",
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
				Name:     "custom_excerpt",
				Type:     "string",
				Optional: newTrue(),
			},
			{
				Name: "plaintext",
				Type: "string",
			},
			{
				Name:     "feature_image",
				Type:     "string",
				Optional: newTrue(),
			},
		},
		DefaultSortingField: defaultSortingField(),
	}

	_, errCreate := typesenseClient.Collections().Create(schema)
	if errCreate != nil {
		log.Printf("Got error %s", errCreate)
	} else {
		log.Println("TypesenseHandler: New Typesense collection created...")
	}

	params := &api.ImportDocumentsParams{
		Action:    action(),
		BatchSize: batchSize(),
	}

	_, errImport := typesenseClient.Collection("blog-posts").Documents().Import(blogPostsForTypesense, params)

	if errImport != nil {
		log.Printf("Got error %s", errImport)
	} else {
		log.Printf("TypesenseHandler: Imported documents to Typesense...")
	}
	return errImport
}

func TypesenseReindexHandler(c *fiber.Ctx) error {
	key := c.Query("key")
	if key != TYPESENSE_ADMIN_API_KEY {
		log.Println("TypesenseReindexHandler: Not authorized")
		return c.Status(http.StatusUnauthorized).SendString("Not authorized")
	}
	log.Println("TypesenseReindexHandler triggered...")
	IndexTypesense()
	log.Println("TypesenseReindexHandler finished executing...")
	return c.JSON("ok")
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
	f := 500
	return &f
}

func action() *string {
	f := "create"
	return &f
}
