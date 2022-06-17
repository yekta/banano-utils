package priceSources

import (
	"encoding/json"
	"log"
	"net/http"
	
	"github.com/yekta/banano-price-service/prices/structs"
)

func GetCoinGecko() priceStructs.SPriceSet {
	coinGeckoURL := "https://api.coingecko.com/api/v3/simple/price?ids=banano&vs_currencies=usd,btc"
	respCoinGecko, errCoinGecko := http.Get(coinGeckoURL)
	if errCoinGecko != nil {
		log.Fatalln(errCoinGecko)
	}
	var resultCoinGecko priceStructs.SCoinGeckoResponse
	json.NewDecoder(respCoinGecko.Body).Decode(&resultCoinGecko)
	
	priceSet := priceStructs.SPriceSet{
		USD: resultCoinGecko.Banano.USD,
		BTC: resultCoinGecko.Banano.BTC,
	}

	log.Println("CoinGecko done")

	return priceSet
}