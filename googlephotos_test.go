package googlephotos

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

func TestMediaFetch(t *testing.T) {

	logrus.SetLevel(logrus.DebugLevel)

	accessToken := os.Getenv("GOOGLE_OAUTH_ACCESS_TOKEN")

	media, err := fetchLibraryMedia(accessToken, &query{
		Filters: &filters{
			ContentFilter: &contentFilter{
				IncludedContentCategories: []string{"LANDSCAPES"},
			},
		},
	})
	assert.NoError(t, err)
	assert.Len(t, media, 289)

	albums, err := fetchAlbums(accessToken)
	assert.NoError(t, err)
	assert.Len(t, albums, 37)

	filteredAlbums, err := fetchAlbums(accessToken, "Wander.Haus", "Workplaces")
	assert.NoError(t, err)
	assert.Len(t, filteredAlbums, 2)

	totalMediaItems, err := strconv.Atoi(albums[0].TotalMediaItems)
	assert.NoError(t, err)
	assert.True(t, totalMediaItems > 0)

	albumMedia, err := fetchLibraryMedia(accessToken, &query{
		AlbumId: albums[0].ID,
	})
	assert.NoError(t, err)
	assert.Equal(t, totalMediaItems, len(albumMedia))

}
