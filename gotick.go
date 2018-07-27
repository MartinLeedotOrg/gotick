package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/buger/goterm"
)

const apiURL = "https://api.coinmarketcap.com/v2"

type listing struct {
	Data []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Symbol      string `json:"symbol"`
		WebsiteSlug string `json:"website_slug"`
	} `json:"data"`
	Metadata struct {
		Timestamp           int         `json:"timestamp"`
		NumCryptocurrencies int         `json:"num_cryptocurrencies"`
		Error               interface{} `json:"error"`
	} `json:"metadata"`
}

type ticker struct {
	Data struct {
		ID                int     `json:"id"`
		Name              string  `json:"name"`
		Symbol            string  `json:"symbol"`
		WebsiteSlug       string  `json:"website_slug"`
		Rank              int     `json:"rank"`
		CirculatingSupply float64 `json:"circulating_supply"`
		TotalSupply       float64 `json:"total_supply"`
		MaxSupply         float64 `json:"max_supply"`
		Quotes            struct {
			USD struct {
				Price            float64 `json:"price"`
				Volume24H        float64 `json:"volume_24h"`
				MarketCap        float64 `json:"market_cap"`
				PercentChange1H  float64 `json:"percent_change_1h"`
				PercentChange24H float64 `json:"percent_change_24h"`
				PercentChange7D  float64 `json:"percent_change_7d"`
			} `json:"USD"`
		} `json:"quotes"`
		LastUpdated int `json:"last_updated"`
	} `json:"data"`
	Metadata struct {
		Timestamp int         `json:"timestamp"`
		Error     interface{} `json:"error"`
	} `json:"metadata"`
}

var currencies = new(listing)
var symbols = symbolFlag{}
var interval = flag.Duration("w", 0, "watch interval '-w 5s'")
var tickerMap = map[string]int{}

type symbolFlag []string

func (i *symbolFlag) String() string {
	return fmt.Sprint(strings.Join(*i, " "))
}

func (i *symbolFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func init() {
	currencies.update()
}

func main() {
	flag.Var(&symbols, "t", "list of tokens '-t BTC -t BTH -t XRP'")
	flag.Parse()
	if len(symbols) == 0 {
		symbols = []string{"BTC", "ETH", "BCH"}
	}

	for _, v := range currencies.Data {
		tickerMap[v.Symbol] = v.ID
	}

	if *interval != 0 {
		for {
			output := makeOutput()
			goterm.Clear()
			goterm.MoveCursor(1, 1)
			goterm.Println(time.Now().Round(0))
			goterm.Printf(output)
			goterm.Flush()

			time.Sleep(*interval)
		}

	}
	output := makeOutput()
	fmt.Println(output)
}

func makeOutput() string {
	var output string

	for i, v := range symbols {
		v = strings.ToUpper(v)
		id, ok := tickerMap[v]
		if ok == false {
			log.Printf("symbol '%s' not found\nContinuing without it...", v)
			symbols = append(symbols[:i], symbols[i+1:]...)
			time.Sleep(time.Second * 3)
			break
		}
		result := getTicker(id)
		toPrint := fmt.Sprintf("%s: $%.2f (%.2f%s)\t", result.Data.Symbol,
			result.Data.Quotes.USD.Price, result.Data.Quotes.USD.PercentChange24H, "%%")
		output = output + toPrint
	}
	if len(symbols) == 0 {
		log.Fatal("No symbols left to try")
	}

	return output

}

func getTicker(id int) ticker {
	url := fmt.Sprintf("%s/ticker/%s", apiURL, strconv.Itoa(id))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	result := ticker{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}
	return result
}

func (record *listing) update() {
	url := fmt.Sprintf("%s/listings/", apiURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(record); err != nil {
		log.Fatal(err)
	}
}
