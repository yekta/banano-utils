package blog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	blogStructs "github.com/yekta/banano-price-service/blog/structs"
)

var lastPostToMedium = ""
var lastBlogIndex = ""

func GhostToMediumHandler(c *fiber.Ctx) error {
	start := time.Now()
	key := c.Query("key")
	if key != GHOST_TO_MEDIUM_SECRET {
		log.Println("GhostToMediumHandler: Not authorized")
		return c.Status(http.StatusUnauthorized).SendString("Not authorized")
	}

	log.Println("-- GhostToMediumHandler: Triggered...")

	var payload blogStructs.SGhostPostWebhook
	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	if lastPostToMedium == payload.Post.Previous.Title+payload.Post.Previous.UpdatedAt {
		log.Println("GhostToMediumHandler: Sent already, skipping")
		return c.Status(http.StatusTooManyRequests).SendString("Sent already, skipping")
	}
	lastPostToMedium = payload.Post.Previous.Title + payload.Post.Previous.UpdatedAt

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

	fmt.Print("\n\n\n")
	fmt.Printf("GhostToMediumHandler: Endpoint: %s\n", mediumPostEndpoint)
	fmt.Printf("GhostToMediumHandler: Medium Post: %s\n", mediumPost)
	fmt.Printf("GhostToMediumHandler: Payload: %s\n", payload)
	fmt.Print("\n\n\n")

	req.Header.Add("Authorization", "Bearer "+MEDIUM_SECRET)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("GhostToMedium error: %s", err)
		return c.Status(http.StatusInternalServerError).SendString("Something went wrong")
	}
	defer resp.Body.Close()

	fmt.Printf("GhostToMediumHandler: Response Status: %s\n", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Printf("GhostToMedium error: %s", err)
		return c.Status(http.StatusInternalServerError).SendString("Something went wrong")
	}

	fmt.Println(string(body))

	r := blogStructs.SBlogResponse{
		Data: blogStructs.SBlogResponseData{
			Title: post.Title,
		},
	}

	log.Printf(`GhostToMediumHandler: Submitted to Medium -> "%s"...`, post.Title)
	log.Printf("-- GhostToMediumHandler: Finished in %s!", time.Since(start))

	return c.JSON(r)
}

func BlogPostHandler(c *fiber.Ctx) error {
	start := time.Now()
	key := c.Query("key")
	if key != GHOST_API_KEY {
		log.Println("BlogPostHandler: Not authorized")
		return c.Status(http.StatusUnauthorized).SendString("Not authorized")
	}

	slug := c.Params("slug")
	log.Printf(`-- BlogPostHandler: Triggered for "%s"...`, slug)
	post, ok := blogSlugToPost[slug]

	if !ok {
		return c.Status(http.StatusNotFound).SendString("Not found")
	}

	fieldsStr := c.Query("fields")
	if fieldsStr != "" {
		fields := strings.Split(fieldsStr, ",")
		post = filterByFields(post, fields)
	} else {
		post = blogStructs.SGhostPost{
			Title:         post.Title,
			Slug:          post.Slug,
			Html:          post.Html,
			PublishedAt:   post.PublishedAt,
			Excerpt:       post.Excerpt,
			CustomExcerpt: post.CustomExcerpt,
			Tags:          post.Tags,
			FeatureImage:  post.FeatureImage,
			ReadingTime:   post.ReadingTime,
			Similars:      post.Similars,
		}
	}

	log.Printf(`-- BlogPostHandler: Finished for "%s" in %s!`, slug, time.Since(start))

	return c.JSON(post)
}

func BlogPostsHandler(c *fiber.Ctx) error {
	start := time.Now()
	key := c.Query("key")
	if key != GHOST_API_KEY {
		log.Println("BlogPostsHandler: Not authorized")
		return c.Status(http.StatusUnauthorized).SendString("Not authorized")
	}

	log.Println("-- BlogPostsHandler: Triggered...")

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

	log.Printf("-- BlogPostsHandler: Finished in %s!", time.Since(start))

	return c.JSON(postsRes)
}

func IndexBlogHandler(c *fiber.Ctx) error {
	start := time.Now()
	key := c.Query("key")
	if key != GHOST_TO_MEDIUM_SECRET {
		log.Println("IndexBlogHandler: Not authorized")
		return c.Status(http.StatusUnauthorized).SendString("Not authorized")
	}

	log.Println("-- IndexBlogHandler: Triggered...")

	var payload blogStructs.SGhostPostWebhook
	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	if lastBlogIndex == payload.Post.Previous.Title+payload.Post.Previous.UpdatedAt {
		log.Println("IndexBlogHandler: Indexed already, skipping")
		return c.Status(http.StatusTooManyRequests).SendString("Indexed already, skipping")
	}
	lastBlogIndex = payload.Post.Previous.Title + payload.Post.Previous.UpdatedAt

	IndexBlog(false)

	log.Printf("-- IndexBlogHandler: Finished in %s!", time.Since(start))

	return c.Status(http.StatusOK).SendString("OK")
}
