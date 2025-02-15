package priceSources

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/yekta/banano-price-service/prices/structs"
)

func GetCoinex() priceStructs.SPriceSet {
	coinexTickerURL := "https://api.coinex.com/v1/market/ticker"
	coinexBanBtcURL := fmt.Sprintf("%s%s", coinexTickerURL, "?market=BANANOBTC")
	coinexBanUsdtURL := fmt.Sprintf("%s%s", coinexTickerURL, "?market=BANANOUSDT")

	respCoinexBtc, errCoinexBtc := http.Get(coinexBanBtcURL)
	if errCoinexBtc != nil {
		log.Fatalln(errCoinexBtc)
	}
	var resultCoinexBtc priceStructs.SCoinexTickerResponse
	json.NewDecoder(respCoinexBtc.Body).Decode(&resultCoinexBtc)

	respCoinexUsdt, errCoinexUsdt := http.Get(coinexBanUsdtURL)
	if errCoinexUsdt != nil {
		log.Fatalln(errCoinexUsdt)
	}
	var resultCoinexUsdt priceStructs.SCoinexTickerResponse
	json.NewDecoder(respCoinexUsdt.Body).Decode(&resultCoinexUsdt)

	var coinexBANUSDT float64
	var coinexBANBTC float64
	if s, err := strconv.ParseFloat(resultCoinexUsdt.Data.Ticker.Last, 64); err == nil {
		coinexBANUSDT = s
	}
	if s, err := strconv.ParseFloat(resultCoinexBtc.Data.Ticker.Last, 64); err == nil {
		coinexBANBTC = s
	}

	log.Println("GetPrices: Coinex done...")

	return priceStructs.SPriceSet{
		USD: coinexBANUSDT,
		BTC: coinexBANBTC,
	}
}
