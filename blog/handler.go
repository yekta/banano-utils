package blog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	blogStructs "github.com/yekta/banano-price-service/blog/structs"
)

func BlogHandler(c *fiber.Ctx, MEDIUM_SECRET string, MEDIUM_USER_ID string) error {
	log.Println("BlogHandler triggered...")

	var payload blogStructs.SGhostPostWebhook
	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	post := payload.Post.Current

	mediumPostEndpoint := "https://api.medium.com/v1/users/" + MEDIUM_USER_ID + "/posts"
	mediumPost := blogStructs.SMediumPost{
		Title:         post.Title,
		ContentFormat: "html",
		Content:       GhostToMediumHtmlConverter(post.Html, post.Title),
		PublishStatus: "draft",
		CanonicalUrl:  "https://banano.cc/blog/" + post.Slug,
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
	return c.JSON(r)
}

func GhostToMediumHtmlConverter(html string, title string) string {
	resHtml := "<h1>" + title + "</h1>" + html
	return resHtml
}
