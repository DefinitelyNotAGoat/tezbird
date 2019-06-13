package cmd

import (
	"fmt"
	"os"

	gotezos "github.com/DefinitelyNotAGoat/go-tezos"
	"github.com/DefinitelyNotAGoat/littlebird/config"
	lbs "github.com/DefinitelyNotAGoat/littlebird/services"
	"github.com/DefinitelyNotAGoat/littlebird/twitter"
	"github.com/DefinitelyNotAGoat/tezbird/services"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newStartCommand() *cobra.Command {
	var (
		path       string
		prefix     string
		node       string
		min        int
		configpath string
	)

	var start = &cobra.Command{
		Use:   "start",
		Short: "start starts a twitter bot for tezos data",
		Run: func(cmd *cobra.Command, args []string) {
			if path == "" {
				fmt.Printf("[tezbird][preflight] error: no api keys passed (e.g. --keys=./twitter.yml")
				os.Exit(1)
			}

			quit := make(chan int)

			gt, err := gotezos.NewGoTezos(node)
			if err != nil {
				fmt.Println(errors.Wrap(err, "could not find tezos node").Error())
				os.Exit(1)
			}

			conf, err := config.NewLiveConfig(configpath)
			if err != nil {
				fmt.Println(errors.Wrap(err, "unable to load sentiment config").Error())
				os.Exit(1)
			}

			gecko := services.NewCoinGecko()
			addressBook := services.NewAddressBook()
			twitterBot, err := twitter.NewTwitterSession(path, prefix)

			if err != nil {
				fmt.Println(errors.Wrap(err, "could not start twitter client").Error())
				os.Exit(1)
			}

			services := []services.Service{
				gecko,
				addressBook,
				services.NewTransferWatch(twitterBot, gt, gecko, addressBook, min),
				lbs.NewSentiment(twitterBot, conf),
				services.NewVote(twitterBot, gt, addressBook),
			}

			for _, serv := range services {
				serv.Start()
			}

			<-quit
		},
	}

	start.PersistentFlags().StringVar(&configpath, "config", "", "path to config.json file containing tezos key words to use and other relavent info (e.g. path/to/my/file/)")
	start.PersistentFlags().StringVar(&path, "keys", "", "path to twitter.yml file containing API keys if not in current dir (e.g. path/to/my/file/)")
	start.PersistentFlags().StringVar(&node, "node", "http://127.0.0.1:8732", "url to tezos node (e.g. http://127.0.0.1:8732)")
	start.PersistentFlags().StringVar(&prefix, "prefix", "", "prefix to all tweets (e.g. DefinitelyNotABot: -- will read DefinitelyNotABot: my tweet)")
	start.PersistentFlags().IntVar(&min, "transfer-min", 50000, "minimum threshold for transfers (e.g. 5000)")

	return start
}
