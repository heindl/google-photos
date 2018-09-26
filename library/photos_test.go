// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package library

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMediaFetch(t *testing.T) {

	//logrus.SetLevel(logrus.DebugLevel)

	accessToken := os.Getenv("GOOGLE_OAUTH_ACCESS_TOKEN")

	media, err := fetchLibraryMedia(accessToken, &query{
		Filters: &filters{
			ContentFilter: &contentFilter{
				IncludedContentCategories: []string{"LANDSCAPES"},
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 290, len(media))

	albums, err := fetchAlbums(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, 37, len(albums))

	filteredAlbums, err := fetchAlbums(accessToken, "Wander.Haus", "Workplaces")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(filteredAlbums))

	totalMediaItems, err := strconv.Atoi(albums[0].MediaItemsCount)
	assert.NoError(t, err)
	assert.True(t, totalMediaItems > 0)

	albumMedia, err := fetchLibraryMedia(accessToken, &query{
		AlbumID: albums[0].ID,
	})
	assert.NoError(t, err)
	assert.Equal(t, totalMediaItems, len(albumMedia))

}
