package main

import (
	"log"
	"flag"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/yekta/banano-price-service/sources"
	"github.com/yekta/banano-price-service/structs"
	"github.com/robfig/cron/v3"
)

var prices structs.SPrices;

func main() {
	serverPort := flag.Int("port", 3000, "Port to listen on")

	app := fiber.New()
	cors := cors.New()
	app.Use(cors)

	GetAndSetPrices()
	cron := cron.New()
	cron.AddFunc("@every 15s", GetAndSetPrices)
	cron.Start()

	app.Get("/prices", func(c *fiber.Ctx) error {
		return c.JSON(prices)
	})
	
	log.Fatal(app.Listen(fmt.Sprintf(":%d", *serverPort)))
}

func GetAndSetPrices() {
	res := GetPrices();
	prices = res;
}

func GetPrices() structs.SPrices {
	log.Println("\n\nGetting prices...")

	var res structs.SPrices;

	c1 := make(chan structs.SPriceSet);
	c2 := make(chan structs.SPriceSet);

	go func(){
		defer close(c1)
		c1 <- sources.GetCoinGecko();
	}();
	go func(){
		defer close(c2)
		c2 <- sources.GetCoinex();
	}();

	res.CoinGecko = <- c1;
	res.Coinex = <- c2;
	res.Main = res.Coinex;
	return res;
}