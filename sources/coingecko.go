package sources

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/yekta/banano-price-service/structs"
)

func GetCoinGecko() structs.SPriceSet {
	coinGeckoURL := "https://api.coingecko.com/api/v3/simple/price?ids=banano&vs_currencies=usd,btc"
	respCoinGecko, errCoinGecko := http.Get(coinGeckoURL)
	if errCoinGecko != nil {
		log.Fatalln(errCoinGecko)
	}
	var resultCoinGecko structs.SCoinGeckoResponse
	json.NewDecoder(respCoinGecko.Body).Decode(&resultCoinGecko)
	
	priceSet := structs.SPriceSet{
		USD: resultCoinGecko.Banano.USD,
		BTC: resultCoinGecko.Banano.BTC,
	}
	return priceSet
}