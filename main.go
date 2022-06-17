package main

import (
	"log"
	"flag"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/yekta/banano-price-service/prices/sources"
	"github.com/yekta/banano-price-service/prices/structs"
	"github.com/yekta/banano-price-service/medium-posts/sources"
	"github.com/yekta/banano-price-service/medium-posts/structs"
	"github.com/yekta/banano-price-service/shared/structs"
	"github.com/robfig/cron/v3"
)

var prices priceStructs.SPrices;
var mediumPosts mediumPostsStructs.MediumPosts;

func main() {
	serverPort := flag.Int("port", 3000, "Port to listen on")

	app := fiber.New()
	cors := cors.New()
	app.Use(cors)

	cron := cron.New()
	cron.AddFunc("@every 15s", GetAndSetPrices)
	cron.AddFunc("@every 3m", GetAndSetMediumPosts)
	cron.Start()

	GetAndSetPrices()
	GetAndSetMediumPosts()

	app.Get("/prices", func(c *fiber.Ctx) error {
		return c.JSON(prices)
	})

	app.Get("/medium-posts", func(c *fiber.Ctx) error {
		var shallowPosts mediumPostsStructs.MediumPosts;
		shallowPosts.LastBuildTimestamp = mediumPosts.LastBuildTimestamp;
		for _, post := range mediumPosts.Posts {
			shallowPosts.Posts = append(shallowPosts.Posts, mediumPostsStructs.MediumPost{
				Title: post.Title,
				Description: post.Description,
				PublishTimestamp: post.PublishTimestamp,
				LastUpdateTimestamp: post.LastUpdateTimestamp,
				Slug: post.Slug,
				Tags: post.Tags,
				Image: post.Image,
			});
		}
		return c.JSON(shallowPosts)
	})

	app.Get("/medium-posts/:slug", func(c *fiber.Ctx) error {
		slug := c.Params("slug")
		for _, post := range mediumPosts.Posts {
			if post.Slug == slug {
				return c.JSON(post)
			}
		}
		var err sharedStructs.ErrorResponse = sharedStructs.ErrorResponse{
			Error: "Post not found",
		}
		return c.Status(404).JSON(err)
	})
	
	log.Fatal(app.Listen(fmt.Sprintf(":%d", *serverPort)))
}

func GetAndSetPrices() {
	res := GetPrices();
	prices = res;
}

func GetPrices() priceStructs.SPrices {
	log.Println("\n\nGetting prices...")

	var res priceStructs.SPrices;

	c1 := make(chan priceStructs.SPriceSet);
	c2 := make(chan priceStructs.SPriceSet);

	go func(){
		defer close(c1)
		c1 <- priceSources.GetCoinGecko();
	}();
	go func(){
		defer close(c2)
		c2 <- priceSources.GetCoinex();
	}();

	res.CoinGecko = <- c1;
	res.Coinex = <- c2;
	res.Main = res.Coinex;
	return res;
}

func GetAndSetMediumPosts() {
	res := GetMediumPosts();
	mediumPosts = res;
}

func GetMediumPosts() mediumPostsStructs.MediumPosts{
	log.Println("\n\nGetting Medium posts...")

	var res mediumPostsStructs.MediumPosts;

	c1 := make(chan mediumPostsStructs.MediumPosts);

	go func(){
		defer close(c1)
		c1 <- mediumPostsSources.GetMediumPosts();
	}();

	res = <- c1;
	return res;
}