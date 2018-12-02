package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
)

var (
	tickerURL = "https://api.coinmarketcap.com/v1/ticker"
)

// Response used to hold response data from cmc
type Response struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Symbol             string `json:"symbol"`
	Rank               string `json:"rank"`
	PriceUsd           string `json:"price_usd"`
	PriceBtc           string `json:"price_btc"`
	TwentyFourHrVolume string `json:"24h_volume_usd"`
	MarketCapUsd       string `json:"market_cap_usd"`
	AvailableSupply    string `json:"available_supply"`
	TotalSupply        string `json:"total_supply"`
	MaxSupply          string `json:"null"`
	PercentChange1h    string `json:"percent_change_1h"`
	PercentChange24h   string `json:"percent_change_24h"`
	PercentChange7d    string `json:"percent_change_7d"`
	LastUpdate         string `json:"last_updated"`
}

// RetrieveUsdPrice is used to retrieve the USD price for a coin from CMC
func RetrieveUsdPrice(coin string) (float64, error) {
	url := fmt.Sprintf("%s/%s", tickerURL, coin)
	response, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	var decode []Response
	if err = json.Unmarshal(body, &decode); err != nil {
		return 0, err
	}
	return strconv.ParseFloat(decode[0].PriceUsd, 64)
}

// RetrieveEthUsdPriceNoDecimals is used to retrieve the eth usd price without decimals
func RetrieveEthUsdPriceNoDecimals() (int64, error) {
	response, err := http.Get("https://api.coinmarketcap.com/v1/ticker/ethereum/")
	if err != nil {
		return 0, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	var decode []Response
	err = json.Unmarshal(body, &decode)
	if err != nil {
		return 0, err
	}

	f, err := strconv.ParseFloat(decode[0].PriceUsd, 64)
	if err != nil {
		return 0, err
	}
	bigF := big.NewFloat(f)
	bigFloatString := bigF.String()
	var s string
	for _, v := range bigFloatString {
		if string(v) == "." {
			break
		}
		s += string(v)
	}
	return strconv.ParseInt(s, 10, 64)
}
