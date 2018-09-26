// Copyright (c) 2018 Parker Heindl. All rights reserved.
//
// Use of this source code is governed by the MIT License.
// Read LICENSE.md in the project root for information.

package library

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Album is the basic unit of organization in Google Photos.
type Album struct {
	ID                    string `json:"id"`
	Title                 string `json:"title"`
	ProductURL            string `json:"productUrl"`
	CoverPhotoBaseURL     string `json:"coverPhotoBaseUrl"`
	CoverPhotoMediaItemID string `json:"coverPhotoMediaItemId"`
	IsWriteable           bool   `json:"isWriteable"`
	MediaItemsCount       string `json:"mediaItemsCount"`
}

// Albums represents a list of Album.
type Albums []*Album

func (Ω Albums) filterToTitles(titles ...string) Albums {
	if len(titles) == 0 {
		return Ω
	}
	titlM := map[string]struct{}{}
	for _, t := range titles {
		titlM[t] = struct{}{}
	}
	y := Albums{}
	for _, album := range Ω {
		if _, ok := titlM[album.Title]; ok {
			y = y.addToSet(album)
		}
	}
	return y
}
func (Ω Albums) contains(albumID string) bool {
	for _, a := range Ω {
		if a.ID == albumID {
			return true
		}
	}
	return false
}

func (Ω Albums) addToSet(a *Album) Albums {
	if a == nil {
		return Ω
	}
	if Ω.contains(a.ID) {
		return Ω
	}
	return append(Ω, a)
}

type albumResponse struct {
	Albums        []*Album `json:"albums"`
	NextPageToken string   `json:"nextPageToken"`
}

func fetchAlbums(accessToken string, titles ...string) (Albums, error) {
	pageToken := ""
	albums := Albums{}
	for {
		albumPageResponse, err := fetchAlbumPage(accessToken, pageToken)
		if err != nil {
			return nil, err
		}
		albums = append(albums, albumPageResponse.Albums...)
		if albumPageResponse.NextPageToken == "" {
			filtered := albums.filterToTitles(titles...)
			logrus.Debugf("Resolving album fetch request with %d items, filtered from %d", len(filtered), len(albums))
			return filtered, nil
		}
		pageToken = albumPageResponse.NextPageToken
	}
}

func fetchAlbumPage(accessToken string, pageToken string) (*albumResponse, error) {

	logrus.WithFields(logrus.Fields{
		"pageToken": pageToken,
	}).Debug("Requesting album page")

	req, err := http.NewRequest("GET", endpointAlbumList, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	q := req.URL.Query()
	q.Add("pageToken", pageToken)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer safeClose(resp.Body, &err)

	res := &albumResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "could not decode album response")
	}
	return res, nil
}
