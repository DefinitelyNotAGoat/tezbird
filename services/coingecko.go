package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
)

//Simple is a structure to represent coingecko's simple api
type Simple struct {
	Tezos Currency `json:"tezos"`
}

// Currency apart of Simple
type Currency struct {
	USD float64 `json:"usd"`
}

// CoinGecko is a service to hit coingecko's api
type CoinGecko struct {
	price float64
	mu    sync.Mutex
}

// NewCoinGecko returns a CoinGecko
func NewCoinGecko() *CoinGecko {
	return &CoinGecko{}
}

// Start starts a CoinGecko
func (c *CoinGecko) Start() {
	errch := make(chan error, 10)
	c.watchPrice(errch)

	go func() {
		for {
			select {
			case err := <-errch:
				fmt.Println(err)
			}
		}
	}()
}

func (c *CoinGecko) watchPrice(errch chan error) {
	url := "https://api.coingecko.com/api/v3/simple/price"

	netTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errch <- err
	}

	q := req.URL.Query()
	q.Add("ids", "tezos")
	q.Add("vs_currencies", "usd")
	req.URL.RawQuery = q.Encode()

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-ticker.C:
				resp, err := client.Do(req)
				if err != nil {
					errch <- err
					continue
				}

				bytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					errch <- err
					continue
				}

				var s Simple
				s, err = s.unmarshalJSON(bytes)
				if err != nil {
					errch <- err
					continue
				}
				c.mu.Lock()
				c.price = s.Tezos.USD
				c.mu.Unlock()
			}
		}
	}()
}

// GetPrice gets the current tezos price in usd
func (c *CoinGecko) GetPrice() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.price
}

func (s *Simple) unmarshalJSON(v []byte) (Simple, error) {
	var simple Simple
	err := json.Unmarshal(v, &simple)
	return simple, err
}
