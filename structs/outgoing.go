package structs

type SPriceSet struct {
	USD float64 `json:"USD"`
	BTC float64 `json:"BTC"`
}

type SPrices struct {
	Main SPriceSet `json:"main"`
	Coinex SPriceSet `json:"coinex"`
	CoinGecko SPriceSet `json:"coinGecko"`
}