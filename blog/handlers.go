package blog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
	blogStructs "github.com/yekta/banano-price-service/blog/structs"
	sharedUtils "github.com/yekta/banano-price-service/shared"
	"golang.org/x/exp/slices"
)

var TYPESENSE_ADMIN_API_KEY = sharedUtils.GetEnv("TYPESENSE_ADMIN_API_KEY")
var GHOST_API_KEY = sharedUtils.GetEnv("GHOST_API_KEY")
var GHOST_TO_MEDIUM_SECRET = sharedUtils.GetEnv("GHOST_TO_MEDIUM_SECRET")
var MEDIUM_SECRET = sharedUtils.GetEnv("MEDIUM_SECRET")
var MEDIUM_USER_ID = sharedUtils.GetEnv("MEDIUM_USER_ID")
var blogApiUrl = "https://ghost.banano.cc/ghost/api/content"
var blogPostsForSitemap blogStructs.SGhostPostsForSitemapResponse
var blogPosts blogStructs.SGhostPostsResponse
var blogSlugToPost = make(map[string]blogStructs.SGhostPost)

var fields = [...]string{
	"id",
	"title",
	"slug",
	"created_at",
	"updated_at",
	"published_at",
	"excerpt",
	"custom_excerpt",
	"feature_image",
	"featured",
	"reading_time",
}
var formats = [...]string{"html", "plaintext"}
var include = [...]string{"tags"}
var fieldsStr = strings.Join(fields[:], ",")
var formatsStr = strings.Join(formats[:], ",")
var includeStr = strings.Join(include[:], ",")
var limit = 1000
var blogEndpoint = fmt.Sprintf(`%s/posts/?key=%s&fields=%s&formats=%s&include=%s&limit=%v`, blogApiUrl, GHOST_API_KEY, fieldsStr, formatsStr, includeStr, limit)
var typesenseClient = typesense.NewClient(
	typesense.WithServer("https://typesense.banano.cc"),
	typesense.WithAPIKey(TYPESENSE_ADMIN_API_KEY),
	typesense.WithConnectionTimeout(60*time.Second))
var typesenseParams = &api.ImportDocumentsParams{
	Action:    action(),
	BatchSize: batchSize(),
}

const defaultPostLimit = 15

var schema = &api.CollectionSchema{
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

const secondThreshold = 60

var lastPostToMedium = time.Now().Add(time.Second * -1 * secondThreshold)

func GhostToMediumHandler(c *fiber.Ctx) error {
	key := c.Query("key")
	if key != GHOST_TO_MEDIUM_SECRET {
		log.Println("GhostToMediumHandler: Not authorized")
		return c.Status(http.StatusUnauthorized).SendString("Not authorized")
	}
	if lastPostToMedium.Add(time.Second * secondThreshold).After(time.Now()) {
		log.Println("GhostToMediumHandler: Too many requests, skipping")
		return c.Status(http.StatusTooManyRequests).SendString("Too many requests")
	}

	log.Println("GhostToMediumHandler: Triggered...")

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

	content := GhostToMediumHtmlConverter(post.Html, post.Title)

	mediumPost := blogStructs.SMediumPost{
		Title:         post.Title,
		ContentFormat: "html",
		Content:       content,
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
		log.Printf("Got error: %s", err)
		return c.Status(http.StatusInternalServerError).SendString("Something went wrong")
	}
	json.NewDecoder(resp.Body).Decode(&resp)

	r := blogStructs.SBlogResponse{
		Data: blogStructs.SBlogResponseData{
			Title: post.Title,
		},
	}

	log.Printf(`Submitted the post to Medium with the title "%s"...`, post.Title)
	lastPostToMedium = time.Now()

	IndexTypesense()

	return c.JSON(r)
}

func TypesenseIndexHandler(c *fiber.Ctx) error {
	key := c.Query("key")
	if key != GHOST_TO_MEDIUM_SECRET {
		log.Println("TypesenseIndexHandler: Not authorized")
		return c.Status(http.StatusUnauthorized).SendString("Not authorized")
	}
	log.Println("TypesenseIndexHandler: Triggered...")
	IndexTypesense()
	log.Println("TypesenseIndexHandler finished executing...")
	return c.JSON("ok")
}

func BlogPostsForSitemapHandler(c *fiber.Ctx) error {
	log.Println("BlogPostsForSitemapHandler: Triggered...")
	return c.JSON(blogPostsForSitemap)
}

func BlogPostHandler(c *fiber.Ctx) error {
	key := c.Query("key")
	if key != GHOST_API_KEY {
		log.Println("BlogPostHandler: Not authorized")
		return c.Status(http.StatusUnauthorized).SendString("Not authorized")
	}

	slug := c.Params("slug")
	log.Printf(`BlogPostHandler: Triggered for "%s"`, slug)
	post, ok := blogSlugToPost[slug]

	if !ok {
		return c.Status(http.StatusNotFound).SendString("Not found")
	}

	fieldsStr := c.Query("fields")
	if fieldsStr != "" {
		fields := strings.Split(fieldsStr, ",")
		post = filterByFields(post, fields)
	}

	return c.JSON(post)
}

func BlogPostsHandler(c *fiber.Ctx) error {
	key := c.Query("key")
	if key != GHOST_API_KEY {
		log.Println("BlogPostsHandler: Not authorized")
		return c.Status(http.StatusUnauthorized).SendString("Not authorized")
	}

	log.Println("BlogPostsHandler: Triggered...")

	var postsRes blogStructs.SGhostPostsResponse

	page := c.Query("page")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 1
	}

	limit := c.Query("limit")
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = defaultPostLimit
	}
	if limitInt > len(blogPosts.Posts) {
		limitInt = len(blogPosts.Posts)
	}

	pages := int(math.Ceil(float64(len(blogPosts.Posts)) / float64(limitInt)))
	total := len(blogPosts.Posts)
	if pageInt < 1 {
		pageInt = 1
	} else if pageInt > pages {
		pageInt = pages
	}
	next := pageInt + 1
	if next > pages {
		next = 0
	}
	prev := pageInt - 1
	if prev < 1 {
		prev = 0
	}
	postsRes.Posts = []blogStructs.SGhostPost{}
	min := (pageInt - 1) * limitInt
	max := pageInt * limitInt
	if max > total {
		max = total
	}
	postsRes.Posts = append(postsRes.Posts, blogPosts.Posts[min:max]...)
	postsRes.Meta = blogStructs.SGhostMeta{
		Pagination: blogStructs.SGhostPagination{
			Page:  pageInt,
			Pages: pages,
			Limit: limitInt,
			Total: total,
			Next:  next,
			Prev:  prev,
		},
	}

	fieldsStr := c.Query("fields")
	if fieldsStr != "" {
		fields := strings.Split(fieldsStr, ",")
		for index, post := range postsRes.Posts {
			newPost := filterByFields(post, fields)
			postsRes.Posts[index] = newPost
		}
	}
	return c.JSON(postsRes)
}

func filterByFields(post blogStructs.SGhostPost, fields []string) blogStructs.SGhostPost {
	if !slices.Contains(fields, "id") {
		post.Id = ""
	}
	if !slices.Contains(fields, "slug") {
		post.Slug = ""
	}
	if !slices.Contains(fields, "title") {
		post.Title = ""
	}
	if !slices.Contains(fields, "published_at") {
		post.PublishedAt = ""
	}
	if !slices.Contains(fields, "excerpt") {
		post.Excerpt = ""
	}
	if !slices.Contains(fields, "custom_excerpt") {
		post.CustomExcerpt = ""
	}
	if !slices.Contains(fields, "html") {
		post.Html = ""
	}
	if !slices.Contains(fields, "plaintext") {
		post.Plaintext = ""
	}
	if !slices.Contains(fields, "slug") {
		post.Slug = ""
	}
	if !slices.Contains(fields, "feature_image") {
		post.FeatureImage = ""
	}
	if !slices.Contains(fields, "tags") {
		post.Tags = []blogStructs.SGhostPostTag{}
	}
	return post
}

func GetAndSetBlogPosts() error {
	log.Println("GetAndSetBlogPosts: Getting...")

	resp, err := http.Get(blogEndpoint)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}

	var ghostPosts blogStructs.SGhostPostsResponse
	json.NewDecoder(resp.Body).Decode(&ghostPosts)

	blogPosts = ghostPosts

	for _, post := range blogPosts.Posts {
		blogSlugToPost[post.Slug] = post
	}

	log.Println("GetAndSetBlogPosts: Set!")
	return err
}

func IndexTypesense() error {
	resp, err := http.Get(blogEndpoint)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}

	var ghostPosts blogStructs.SGhostPostsResponse
	json.NewDecoder(resp.Body).Decode(&ghostPosts)

	log.Println("TypesenseHandler: Got Ghost blog posts!")

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

	_, errDel := typesenseClient.Collection("blog-posts").Delete()

	if errDel != nil {
		log.Println("TypesenseHandler: Error deleting collection:", errDel)
	} else {
		log.Println("TypesenseHandler: Typesense collection deleted...")
	}

	_, errCreate := typesenseClient.Collections().Create(schema)
	if errCreate != nil {
		log.Printf("Got error %s", errCreate)
	} else {
		log.Println("TypesenseHandler: New Typesense collection created...")
	}

	_, errImport := typesenseClient.Collection("blog-posts").Documents().Import(blogPostsForTypesense, typesenseParams)

	if errImport != nil {
		log.Printf("Got error %s", errImport)
	} else {
		log.Printf("TypesenseHandler: Imported documents to Typesense...")
	}
	return errImport
}

func GetAndSetBlogPostsForSitemap() {
	log.Println("BlogPostsForSitemap: Getting...")
	fields := [...]string{
		"slug",
		"updated_at",
	}
	fieldsStr := strings.Join(fields[:], ",")
	const limit = 1000
	blogEndpoint := fmt.Sprintf(`%s/posts/?key=%s&fields=%s&limit=%v`, blogApiUrl, GHOST_API_KEY, fieldsStr, limit)
	resp, err := http.Get(blogEndpoint)
	if err != nil {
		log.Printf("Got error: %s", err)
	} else {
		log.Println("BlogPostsForSitemap: Got it!")
		var ghostPosts blogStructs.SGhostPostsForSitemapResponse
		json.NewDecoder(resp.Body).Decode(&ghostPosts)
		blogPostsForSitemap = ghostPosts
		log.Println("BlogPostsForSitemap: Set!")
	}
}

func GhostToMediumHtmlConverter(html string, title string) string {
	s1 := strings.ReplaceAll(html, "<h2", "<h1")
	s2 := strings.ReplaceAll(s1, "</h2>", "</h1>")
	s3 := strings.ReplaceAll(s2, "<h3", "<h2")
	s4 := strings.ReplaceAll(s3, "</h3>", "</h2>")
	resHtml := fmt.Sprintf(`<h1>%s</h1>%s`, title, s4)
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
