package services

import (
	"fmt"
	"strings"

	"github.com/DefinitelyNotAGoat/littlebird/config"
	"github.com/DefinitelyNotAGoat/littlebird/twitter"
	tapi "github.com/dghubble/go-twitter/twitter"
	"github.com/pkg/errors"
	sentiment "gopkg.in/vmarkovtsev/BiDiSentiment.v1"
)

// Sentiment is a sentimental analysis service for tezos twitter
type Sentiment struct {
	twitterBot *twitter.Bot
	conf       config.Live
}

// NewSentiment returns a new Sentiment
func NewSentiment(twitterBot *twitter.Bot, conf config.Live) *Sentiment {
	return &Sentiment{twitterBot: twitterBot, conf: conf}
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
		stream, err := s.twitterBot.Subscribe(s.conf.GetTrackedWords())
		if err != nil {
			errch <- err
			return
		}

		session, _ := sentiment.OpenSession()

		demux := tapi.NewSwitchDemux()
		demux.Tweet = func(tweet *tapi.Tweet) {
			if !s.blockMessage(tweet.Text) {
				result, err := sentiment.Evaluate(
					[]string{tweet.Text},
					session)
				if err != nil {
					errch <- errors.Wrap(err, "could not eval sentiment")
				} else {
					if result[0] <= float32(s.conf.GetFavoriteSentimentRequirement()) && tweet.User.IDStr != s.twitterBot.UserID {
						s.twitterBot.Like(tweet.ID)
					}
					if result[0] <= float32(s.conf.GetRetweetSentimentRequirement()) && tweet.User.IDStr != s.twitterBot.UserID { //positive sentiment
						s.twitterBot.Retweet(tweet.ID, nil)
					}

				}
			}

		}
		go demux.HandleChan(stream.Messages)
	}()
}

func (s *Sentiment) blockMessage(msg string) bool {
	fields := strings.Fields(strings.ToLower(msg))
	hashtags := 0
	for _, field := range fields {
		if isHashTag(field) {
			hashtags++
		}
		for _, word := range s.conf.GetExcludedWords() {
			if field == strings.ToLower(word) {
				return true
			}
		}
	}
	if (float64(hashtags))/float64(len(fields)) >= s.conf.GetLimitTags() {
		return true
	}

	return false
}

func isHashTag(word string) bool {
	if strings.HasPrefix(word, "#") {
		return true
	}
	return false
}
