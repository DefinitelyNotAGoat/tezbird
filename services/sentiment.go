package services

import (
	"fmt"
	"strings"

	"github.com/DefinitelyNotAGoat/littlebird/twitter"
	tapi "github.com/dghubble/go-twitter/twitter"
	"github.com/pkg/errors"
	sentiment "gopkg.in/vmarkovtsev/BiDiSentiment.v1"
)

// Sentiment is a sentimental analysis service for tezos twitter
type Sentiment struct {
	minScoreRetweet  float32
	minScoreFavorite float32
	twitterBot       *twitter.Bot
}

// NewSentiment returns a new Sentiment
func NewSentiment(twitterBot *twitter.Bot, minRetweet, minScoreFavorite float32) *Sentiment {
	return &Sentiment{twitterBot: twitterBot, minScoreRetweet: minRetweet, minScoreFavorite: minScoreFavorite}
}

// Start starts a new sentiment service
func (s *Sentiment) Start() {
	errch := make(chan error, 10)
	s.analyze(errch)

	go func() {
		for {
			select {
			case err := <-errch:
				fmt.Println(err)
			}
		}
	}()
}

func (s *Sentiment) analyze(errch chan error) {
	go func() {
		stream, err := s.twitterBot.Subscribe([]string{
			"tezos",
			"Tezos",
			"xtz",
			"XTZ",
			"Arthur Brietman",
			"Kathlean Brietman",
		})
		if err != nil {
			errch <- err
			return
		}

		session, _ := sentiment.OpenSession()

		demux := tapi.NewSwitchDemux()
		demux.Tweet = func(tweet *tapi.Tweet) {
			if containsTezosReference(tweet.Text) {
				result, err := sentiment.Evaluate(
					[]string{tweet.Text},
					session)
				if err != nil {
					errch <- errors.Wrap(err, "could not eval sentiment")
				} else {
					if containsBlockReference(tweet.Text) {
						return
					}
					if result[0] <= s.minScoreFavorite && tweet.User.IDStr != s.twitterBot.UserID {
						s.twitterBot.Like(tweet.ID)
					}
					if result[0] <= s.minScoreRetweet && tweet.User.IDStr != s.twitterBot.UserID { //positive sentiment
						s.twitterBot.Retweet(tweet.ID, nil)
					}

				}
			}
		}
		go demux.HandleChan(stream.Messages)
	}()
}

func containsTezosReference(msg string) bool {
	if strings.Contains(msg, "Tezos") ||
		strings.Contains(msg, "tezos") ||
		strings.Contains(msg, "XTZ") ||
		strings.Contains(msg, "xtz") ||
		strings.Contains(msg, "arthur brietman") ||
		strings.Contains(msg, "kathleen brietman") ||
		strings.Contains(msg, "Arthur Brietman") ||
		strings.Contains(msg, "Kathleen Brietman") {
		return true
	} else {
		return false
	}
}

func containsBlockReference(msg string) bool {
	if strings.Contains(msg, "Magnum") {
		return true
	} else {
		return false
	}
}
