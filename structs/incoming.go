package structs

type SCoinGeckoResponse struct {
	Banano struct {
		USD float64 `json:"usd"`
		BTC float64 `json:"btc"`
	} `json:"banano"`
}

type SCoinexTickerResponse struct {
	Code int `json:"code"`
	Data struct {
		Date   int64 `json:"date"`
		Ticker struct {
			Vol        string `json:"vol"`
			Low        string `json:"low"`
			Open       string `json:"open"`
			High       string `json:"high"`
			Last       string `json:"last"`
			Buy        string `json:"buy"`
			BuyAmount  string `json:"buy_amount"`
			Sell       string `json:"sell"`
			SellAmount string `json:"sell_amount"`
		} `json:"ticker"`
	} `json:"data"`
	Message string `json:"message"`
}