package config

import (
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// config is a dynamic configuration for the twitter bot
type config struct {
	retweetSentimentRequirement  float64
	favoriteSentimentRequirement float64
	trackedWords                 []string
	excludedWords                []string
	limitTags                    float64
}

// Live wraps the dynamic config for use
type Live struct {
	conf *config
	mu   *sync.Mutex
}

// NewLiveConfig returns a new Live
func NewLiveConfig(filePath string) (Live, error) {
	live := Live{conf: &config{}, mu: &sync.Mutex{}}
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(filePath) // path to look for the config file in
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return Live{}, errors.Wrap(err, "could not read in twitter bot configuration")
	}

	live.conf.retweetSentimentRequirement = viper.GetFloat64("retweet_sentiment_requirement")
	live.conf.favoriteSentimentRequirement = viper.GetFloat64("favorite_sentiment_requirement")
	live.conf.trackedWords = viper.GetStringSlice("tracked_words")
	live.conf.excludedWords = viper.GetStringSlice("exclude_words")
	live.conf.limitTags = viper.GetFloat64("limit_tags")

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		viper.ReadInConfig()
		live.mu.Lock()
		live.conf.retweetSentimentRequirement = viper.GetFloat64("retweet_sentiment_requirement")
		live.conf.favoriteSentimentRequirement = viper.GetFloat64("favorite_sentiment_requirement")
		live.conf.trackedWords = viper.GetStringSlice("tracked_words")
		live.conf.excludedWords = viper.GetStringSlice("exclude_words")
		live.conf.limitTags = viper.GetFloat64("limit_tags")
		live.mu.Unlock()

	})

	return live, nil
}

// GetRetweetSentimentRequirement gets the sentiment requirement for retweets
func (l *Live) GetRetweetSentimentRequirement() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.conf.retweetSentimentRequirement
}

// GetFavoriteSentimentRequirement gets the sentiment requirement for favorites
func (l *Live) GetFavoriteSentimentRequirement() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.conf.favoriteSentimentRequirement
}

// GetTrackedWords gets the words to track for a twitter stream
func (l *Live) GetTrackedWords() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.conf.trackedWords
}

// GetExcludedWords gets the words to exclude in a twitter stream
func (l *Live) GetExcludedWords() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.conf.excludedWords
}

// GetLimitTags gets a tag limitation limiting tweets that contain a certain percentage of tags
func (l *Live) GetLimitTags() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.conf.limitTags
}
