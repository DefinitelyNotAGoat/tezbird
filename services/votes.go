package services

import (
	"fmt"
	"strconv"
	"time"

	gotezos "github.com/DefinitelyNotAGoat/go-tezos"
	"github.com/DefinitelyNotAGoat/tezbird/twitter"
)

type Vote struct {
	bot      *twitter.Bot
	gt       *gotezos.GoTezos
	addrBook *AddressBook
}

func NewVote(bot *twitter.Bot, gt *gotezos.GoTezos, addrBook *AddressBook) *Vote {
	return &Vote{bot: bot, gt: gt, addrBook: addrBook}
}

func (v *Vote) Start() {
	errch := make(chan error, 10)
	v.watchVotesProposals(errch)

	go func() {
		for {
			select {
			case err := <-errch:
				fmt.Println(err)
			}
		}
	}()
}

func (v *Vote) watchVotesProposals(errch chan error) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				block, err := v.gt.Block.GetHead()
				if err != nil {
					errch <- err
					continue
				}

				for _, opout := range block.Operations {
					for _, opin := range opout {
						for _, op := range opin.Contents {
							if op.Kind == "ballot" {
								var tweet string
								src := v.addrBook.Lookup(op.Source)
								delegate, err := v.gt.Delegate.GetDelegate(src)
								if err != nil {
									errch <- err
									tweet = fmt.Sprintf("[Vote Alert] %s voted %s for proposal %s: http://tzscan.io/%s #tezos", src, op.Ballot, op.Proposal, opin.Hash)
								} else {
									rolls, _ := strconv.Atoi(delegate.StakingBalance)
									rolls = rolls / 8000
									tweet = fmt.Sprintf("[Vote Alert] %s voted %s for proposal %s with %d rolls: http://tzscan.io/%s #tezos", src, op.Ballot, op.Proposal, rolls, opin.Hash)
								}

								err = v.bot.Tweet(tweet)
								if err != nil {
									errch <- err
								}

							} else if op.Kind == "proposals" {
								src := v.addrBook.Lookup(op.Source)
								tweet := fmt.Sprintf("[Proposal Alert] %s submited new proposal(s)", src)

								l := len(op.Proposals)
								for i, prop := range op.Proposals {
									if i == l-1 {
										tweet = fmt.Sprintf("%s and %s", tweet, prop)
									} else {
										tweet = fmt.Sprintf("%s %s,", tweet, prop)
									}
								}

								tweet = fmt.Sprintf("%s: http://tzscan.io/%s #tezos", tweet, opin.Hash)
								err = v.bot.Tweet(tweet)
								if err != nil {
									errch <- err
								}

							}
						}
					}
				}

			}
		}
	}()
}
