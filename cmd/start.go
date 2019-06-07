package cmd

import (
	"fmt"
	"os"

	gotezos "github.com/DefinitelyNotAGoat/go-tezos"
	"github.com/DefinitelyNotAGoat/tezbird/services"
	"github.com/DefinitelyNotAGoat/tezbird/twitter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newStartCommand() *cobra.Command {
	var (
		path   string
		prefix string
		node   string
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
				services.NewTransferWatch(twitterBot, gt, gecko, addressBook, 0),
				services.NewSentiment(twitterBot),
			}

			for _, serv := range services {
				serv.Start()
			}

			<-quit
		},
	}

	start.PersistentFlags().StringVar(&path, "keys", "", "path to twitter.yml file containing API keys if not in current dir (e.g. path/to/my/file/)")
	start.PersistentFlags().StringVar(&node, "node", "http://127.0.0.1:8732", "url to tezos node (e.g. http://127.0.0.1:8732)")
	start.PersistentFlags().StringVar(&prefix, "prefix", "", "prefix to all tweets (e.g. DefinitelyNotABot: -- will read DefinitelyNotABot: my tweet)")
	return start
}
