package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/heindl/googlephotos"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	var OAuth2AccessToken string
	var Verbose bool
	var AlbumTitles []string

	var rootCmd = &cobra.Command{
		Use:   "googlephotos",
		Short: "Access a user's Google Photo Library",
	}

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&OAuth2AccessToken, "token", "t", "", "Google oauth2 access token")
	rootCmd.PersistentFlags().StringArrayVarP(&AlbumTitles, "album", "a", []string{}, "Album titles to restrict results")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Return the user's Google Photo Library as a json array",
		RunE: func(cmd *cobra.Command, args []string) error {
			if Verbose {
				logrus.SetLevel(logrus.DebugLevel)
				logrus.SetFormatter(&logrus.JSONFormatter{})
			} else {
				logrus.SetLevel(logrus.WarnLevel)
			}
			media, err := googlephotos.FetchList(googlephotos.Params{
				OAuth2AccessToken: OAuth2AccessToken,
				AlbumTitles:       AlbumTitles,
			})
			if err != nil {
				return err
			}
			b, err := json.Marshal(media)
			if err != nil {
				return errors.Wrap(err, 0)
			}
			fmt.Println(string(b))
			return nil
		},
	})

	rootCmd.Execute()
}
