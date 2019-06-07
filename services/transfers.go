package services

import (
	"fmt"
	"math"
	"strconv"
	"time"

	gotezos "github.com/DefinitelyNotAGoat/go-tezos"
	"github.com/DefinitelyNotAGoat/tezbird/twitter"
	humanize "github.com/dustin/go-humanize"
)

// TransferWatch is a structure to contain the service that watches for transfers
type TransferWatch struct {
	twitterBot *twitter.Bot
	gt         *gotezos.GoTezos
	gecko      *CoinGecko
	addrBook   *AddressBook
	min        int //MUTEZ
}

// NewTransferWatch returns a TransferWatch
func NewTransferWatch(twitterBot *twitter.Bot, gt *gotezos.GoTezos, gecko *CoinGecko, addrBook *AddressBook, min int) *TransferWatch {
	return &TransferWatch{twitterBot: twitterBot, gt: gt, gecko: gecko, addrBook: addrBook, min: min}
}

//Start starts the TransferWatch service
func (tw *TransferWatch) Start() {
	errch := make(chan error, 10)
	tw.watchTransfers(errch)

	go func() {
		for {
			select {
			case err := <-errch:
				fmt.Println(err)
			}
		}
	}()
}

func (tw *TransferWatch) watchTransfers(errch chan error) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				block, err := tw.gt.Block.GetHead()
				if err != nil {
					errch <- err
				}

				for _, opout := range block.Operations {
					for _, opin := range opout {
						for _, op := range opin.Contents {
							if op.Kind == "transaction" {
								amount, _ := strconv.Atoi(op.Amount)
								f := math.Round((float64(amount)/1000000)*100) / 100
								value := math.Round((tw.gecko.GetPrice()*f)*100) / 100

								if int(value) >= tw.min {
									src := tw.addrBook.Lookup(op.Source)
									dst := tw.addrBook.Lookup(op.Destination)

									tweet := fmt.Sprintf("[Transfer Alert] %s sent %s XTZ ($%s) to %s: http://tzscan.io/%s", src, humanize.Commaf(f), humanize.Commaf(value), dst, opin.Hash)
									fmt.Println(tweet)
									err := tw.twitterBot.Tweet(tweet)
									if err != nil {
										errch <- err
									}
								}

							}
						}
					}
				}
			}
		}
	}()
}
