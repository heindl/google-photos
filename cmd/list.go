// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/heindl/google-photos/library"
	"github.com/spf13/cobra"
)

func init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Print photo references to standard out as a JSON document.",
		Run: func(cmd *cobra.Command, args []string) {
			media, err := library.FetchList(library.Params{
				OAuth2AccessToken: flagOAuth2AccessToken,
				AlbumTitles:       flagAlbumTitles,
			})
			if err != nil {
				fmt.Fprint(os.Stderr, err.Error())
				return
			}
			b, err := json.Marshal(media)
			if err != nil {
				fmt.Fprint(os.Stderr, err.Error())
				return
			}
			fmt.Fprint(os.Stdout, string(b))
		},
	}
	rootCmd.AddCommand(listCmd)
}
