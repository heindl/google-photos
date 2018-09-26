// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "google-photos",
	Short: "Read a user's Google Photos library",
	Long: `
The [Google Photos API](https://developers.google.com/photos/) requires an OAuth2 Access Token with the scopes:
- https://www.googleapis.com/auth/photoslibrary.location
- https://www.googleapis.com/auth/photoslibrary.readonly

Soon I'll add a method for generating a token locally, but for now grab one from the [OAuth Sandbox](https://developers.google.com/oauthplayground).
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

var flagOAuth2AccessToken string
var flagVerbose bool
var flagAlbumTitles []string

func init() {

	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Print debug logs.")
	rootCmd.PersistentFlags().StringVarP(&flagOAuth2AccessToken, "token", "t", "", "Google oauth2 access token.")
	rootCmd.PersistentFlags().StringArrayVarP(&flagAlbumTitles, "album", "a", []string{}, "Album titles to limit results.")

	if flagVerbose {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

}
