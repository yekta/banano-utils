package prices

import (
	"log"

	"github.com/gofiber/fiber/v2"
	priceSources "github.com/yekta/banano-price-service/prices/sources"
	priceStructs "github.com/yekta/banano-price-service/prices/structs"
)

var prices priceStructs.SPrices

func PricesHandler(c *fiber.Ctx) error {
	return c.JSON(prices)
}

func GetAndSetPrices() {
	res := GetPrices()
	prices = res
}

func GetPrices() priceStructs.SPrices {
	log.Println("Prices: Getting...")

	var res priceStructs.SPrices

	c1 := make(chan priceStructs.SPriceSet)
	c2 := make(chan priceStructs.SPriceSet)

	go func() {
		c1 <- priceSources.GetCoinGecko()
		defer close(c1)
	}()

	go func() {
		c2 <- priceSources.GetCoinex()
		defer close(c2)
	}()

	res.CoinGecko = <-c1
	res.Coinex = <-c2
	res.Main = res.Coinex
	log.Println("Prices: Got it!")
	return res
}
