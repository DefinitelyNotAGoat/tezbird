package services

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
)

// AddressBook is a service to keep track of wallets
type AddressBook struct {
	ledger map[string]string
	client *http.Client
	mu     sync.Mutex
}

// MyTezosBaker contains api data from https://api.mytezosbaker.com/v1/bakers/
type MyTezosBaker struct {
	Bakers []Baker `json:"bakers"`
}

// Baker contains api data from https://api.mytezosbaker.com/v1/bakers/
type Baker struct {
	Rank                int    `json:"rank"`
	BakerName           string `json:"baker_name"`
	DelegationCode      string `json:"delegation_code"`
	Fee                 string `json:"fee"`
	BakerEfficiency     string `json:"baker_efficiency"`
	AvailableCapacity   string `json:"available_capacity"`
	AcceptingDelegation string `json:"accepting_delegation"`
	ExpectedRoi         string `json:"expected_roi"`
}

// NewAddressBook creates a new AddressBook
func NewAddressBook() *AddressBook {
	m := make(map[string]string)
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
	return &AddressBook{ledger: m, client: client}
}

// Start starts a new AddressBook service
func (a *AddressBook) Start() {
	errch := make(chan error, 10)
	a.warmAddressBookCache(errch)
}

func (a *AddressBook) warmAddressBookCache(errch chan error) {
	all := "https://api.mytezosbaker.com/v1/bakers/"

	req, err := http.NewRequest("GET", all, nil)
	if err != nil {
		errch <- err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		errch <- err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errch <- err
	}

	var m MyTezosBaker
	m, err = m.unmarshalJSON(bytes)
	if err != nil {
		errch <- err
	}

	a.mu.Lock()
	for _, each := range m.Bakers {
		a.ledger[each.DelegationCode] = each.BakerName
	}
	a.mu.Unlock()
}

func (a *AddressBook) lookupMyTezosBaker(addr string) (Baker, error) {
	var b Baker
	url := "https://api.mytezosbaker.com/v1/bakers/" + addr

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return b, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return b, err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return b, err
	}

	b, err = b.unmarshalJSON(bytes)
	if err != nil {
		return b, err
	}

	return b, nil
}

// Lookup looks for the address in the AddressBook ledger
func (a *AddressBook) Lookup(addr string) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	alias, ok := a.ledger[addr]
	if !ok {
		b, err := a.lookupMyTezosBaker(addr)
		if err != nil || b.BakerName == "" {
			return addr
		}
		a.ledger[addr] = b.BakerName
		alias = b.BakerName
	}
	return alias
}

func (m *MyTezosBaker) unmarshalJSON(v []byte) (MyTezosBaker, error) {
	var myTezosBaker MyTezosBaker
	err := json.Unmarshal(v, &myTezosBaker)
	return myTezosBaker, err
}

func (b *Baker) unmarshalJSON(v []byte) (Baker, error) {
	var baker Baker
	err := json.Unmarshal(v, &baker)
	return baker, err
}
