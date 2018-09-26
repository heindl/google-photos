// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package cmd

import (
	"fmt"
	"os"

	"github.com/heindl/google-photos/library"
	"github.com/spf13/cobra"
)

func init() {

	var flagOutputPath string

	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download photos to a directory.",
		Run: func(cmd *cobra.Command, args []string) {

			if flagOutputPath == "" {
				fmt.Fprint(os.Stderr, "Download path required to save photos")
				return
			}

			media, err := library.FetchList(library.Params{
				OAuth2AccessToken: flagOAuth2AccessToken,
				AlbumTitles:       flagAlbumTitles,
			})
			if err != nil {
				fmt.Fprint(os.Stderr, err.Error())
				return
			}

			if err := library.Download(media, flagOutputPath); err != nil {
				fmt.Fprint(os.Stderr, err.Error())
				return
			}
		},
	}

	downloadCmd.LocalFlags().StringVarP(&flagOutputPath, "output-path", "o", "", "output path for download")

	rootCmd.AddCommand(downloadCmd)

}
