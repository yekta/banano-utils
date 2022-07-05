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

	cron := cron.New()
	cron.AddFunc("@every 15s", prices.GetAndSetPrices)
	cron.AddFunc("@every 10m", blog.GetAndSetBlogPostsForSitemap)
	cron.Start()

	go prices.GetAndSetPrices()
	go blog.IndexTypesense()
	go blog.GetAndSetBlogPostsForSitemap()

	app.Get("/prices", prices.PricesHandler)

	app.Post("/blog/ghost-to-medium", blog.GhostToMediumHandler)
	app.Get("/blog/typesense-reindex", blog.TypesenseReindexHandler)
	app.Get("/blog/posts-for-sitemap", blog.BlogPostsForSitemapHandler)

	log.Fatal(app.Listen(fmt.Sprintf(":%d", *serverPort)))
}
