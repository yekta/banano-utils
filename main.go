package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/robfig/cron/v3"
	blog "github.com/yekta/banano-price-service/blog"
	prices "github.com/yekta/banano-price-service/prices"
)

func main() {
	serverPort := flag.Int("port", 3000, "Port to listen on")

	app := fiber.New()
	cors := cors.New()
	app.Use(cors)
	//

	cron := cron.New()
	cron.AddFunc("@every 15s", prices.GetAndSetPrices)
	cron.Start()

	go prices.GetAndSetPrices()
	go blog.IndexBlog(true)

	app.Get("/prices", prices.PricesHandler)
	app.Get("/blog/posts", blog.BlogPostsHandler)
	app.Get("/blog/posts/slug/:slug", blog.BlogPostHandler)
	app.Post("/blog/index", blog.IndexBlogHandler)
	app.Post("/blog/ghost-to-medium", blog.GhostToMediumHandler)

	log.Fatal(app.Listen(fmt.Sprintf(":%d", *serverPort)))
}
