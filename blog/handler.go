package blog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	blogStructs "github.com/yekta/banano-price-service/blog/structs"
)

func BlogHandler(c *fiber.Ctx) error {
	fmt.Println("\nBlogHandler triggered...")
	MEDIUM_SECRET := GetEnv("MEDIUM_SECRET")
	MEDIUM_USER_ID := GetEnv("MEDIUM_USER_ID")

	var payload blogStructs.SGhostPostWebhook
	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	post := payload.Post.Current

	mediumPostEndpoint := "https://api.medium.com/v1/users/" + MEDIUM_USER_ID + "/posts"
	mediumPost := blogStructs.SMediumPost{
		Title:         post.Title,
		ContentFormat: "html",
		Content:       "<h1>" + post.Title + "</h1>" + post.Html,
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
	fmt.Println(req)
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
	return c.JSON(r)
}

func GetEnv(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}
	return os.Getenv(key)
}
