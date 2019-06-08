package twitter

import (
	"fmt"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Bot is a wrapper for a twitter session to post payout results to
type Bot struct {
	title   string
	session *twitter.Client
	UserID  string
	isset   bool
}

// NewTwitterSession creates a new twitter bot based off a twitter.yml file located in the current path or path passed.
// can pass an optional title for the bot
func NewTwitterSession(path string, title string) (*Bot, error) {
	bot := Bot{title: title}
	viper.SetConfigName("twitter")
	if path != "" {
		viper.AddConfigPath(path)
	}
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		return &bot, errors.Wrap(err, "could not find twitter.yml")
	}

	key := viper.GetString("consumerKey")
	keySecret := viper.GetString("consumerKeySecret")
	access := viper.GetString("accessToken")
	accessSecret := viper.GetString("accessTokenSecret")
	if key == "" || access == "" || keySecret == "" || accessSecret == "" {
		return &bot, fmt.Errorf("could not read key or access token")
	}

	config := oauth1.NewConfig(key, keySecret)
	token := oauth1.NewToken(access, accessSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	bot.session = client

	return &bot, nil
}

// Post posts a tweet
func (bot *Bot) Tweet(twt string) error {
	tweet := bot.title + " " + twt
	rtn, _, err := bot.session.Statuses.Update(tweet, nil)
	if err != nil {
		return err
	}
	if !bot.isset {
		bot.UserID = rtn.User.IDStr
	}
	fmt.Println("Tweeted:" + twt)

	return nil
}

// Retweet retweets a tweet
func (bot *Bot) Retweet(id int64, params *twitter.StatusRetweetParams) error {
	tweet, _, err := bot.session.Statuses.Retweet(id, params)
	if err != nil {
		return err
	}
	fmt.Println("Retweeted: " + tweet.Text)

	return nil
}

// Like favorites a tweet at id
func (bot *Bot) Like(id int64) error {
	tweet, _, err := bot.session.Favorites.Create(&twitter.FavoriteCreateParams{ID: id})
	if err != nil {
		return err
	}
	fmt.Println("Favorited: " + tweet.Text)

	return nil
}

// Subscribe to a twitter stream
func (bot *Bot) Subscribe(subjects []string) (*twitter.Stream, error) {
	params := &twitter.StreamFilterParams{
		Track:         subjects,
		StallWarnings: twitter.Bool(true),
	}
	stream, err := bot.session.Streams.Filter(params)
	return stream, err
}
