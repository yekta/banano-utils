package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/robfig/cron/v3"
	blog "github.com/yekta/banano-price-service/blog"
	priceSources "github.com/yekta/banano-price-service/prices/sources"
	priceStructs "github.com/yekta/banano-price-service/prices/structs"
)

var prices priceStructs.SPrices

func main() {
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

	app.Post("/blog", blog.BlogHandler)

	log.Fatal(app.Listen(fmt.Sprintf(":%d", *serverPort)))
}

func GetAndSetPrices() {
	res := GetPrices()
	prices = res
}

func GetPrices() priceStructs.SPrices {
	log.Println("\n\nGetting prices...")

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
