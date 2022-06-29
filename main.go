package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	blog "github.com/yekta/banano-price-service/blog"
	priceSources "github.com/yekta/banano-price-service/prices/sources"
	priceStructs "github.com/yekta/banano-price-service/prices/structs"
)

var prices priceStructs.SPrices

func main() {
	MEDIUM_SECRET := GetEnv("MEDIUM_SECRET")
	MEDIUM_USER_ID := GetEnv("MEDIUM_USER_ID")
	GHOST_TO_MEDIUM_SECRET := GetEnv("GHOST_TO_MEDIUM_SECRET")
	TYPESENSE_ADMIN_API_KEY := GetEnv("TYPESENSE_ADMIN_API_KEY")
	GHOST_API_KEY := GetEnv("GHOST_API_KEY")

	serverPort := flag.Int("port", 3000, "Port to listen on")

	app := fiber.New()
	cors := cors.New()
	app.Use(cors)

	cron := cron.New()
	cron.AddFunc("@every 15s", GetAndSetPrices)
	cron.Start()

	GetAndSetPrices()

	app.Get("/prices", func(c *fiber.Ctx) error {
		return c.JSON(prices)
	})

	app.Post("/blog", func(c *fiber.Ctx) error {
		return blog.BlogHandler(c, MEDIUM_SECRET, MEDIUM_USER_ID, GHOST_TO_MEDIUM_SECRET, TYPESENSE_ADMIN_API_KEY, GHOST_API_KEY)
	})

	app.Get("/typesense-reindex", func(c *fiber.Ctx) error {
		return blog.TypesenseReindexHandler(c, TYPESENSE_ADMIN_API_KEY, GHOST_API_KEY)
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%d", *serverPort)))
}

func GetAndSetPrices() {
	res := GetPrices()
	prices = res
}

func GetPrices() priceStructs.SPrices {
	log.Println("Getting prices...")

	var res priceStructs.SPrices

	c1 := make(chan priceStructs.SPriceSet)
	c2 := make(chan priceStructs.SPriceSet)

	go func() {
		defer close(c1)
		c1 <- priceSources.GetCoinGecko()
	}()

	go func() {
		defer close(c2)
		c2 <- priceSources.GetCoinex()
	}()

	res.CoinGecko = <-c1
	res.Coinex = <-c2
	res.Main = res.Coinex
	return res
}

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}

func GetEnv(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("\nNo .env file, will try to use env variables...")
	}
	return os.Getenv(key)
}
