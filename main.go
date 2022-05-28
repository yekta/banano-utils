package main

import (
	"log"
	"flag"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/yekta/banano-price-service/sources"
	"github.com/yekta/banano-price-service/structs"
	"github.com/robfig/cron/v3"
)

var prices structs.SPrices;

func main() {
	serverPort := flag.Int("port", 8080, "Port to listen on")

	app := fiber.New()

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
	log.Println("Getting prices...")
	coinGecko := sources.GetCoinGecko();
	coinex := sources.GetCoinex();
	res := structs.SPrices {
		Main: coinex,
		Coinex: coinex,
		CoinGecko: coinGecko,
	}
	return res;
}