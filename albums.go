package googlephotos

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"github.com/machinae/stringslice"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Album struct {
	ID                    string `json:"id"`
	Title                 string `json:"title"`
	ProductURL            string `json:"productUrl"`
	CoverPhotoBaseURL     string `json:"coverPhotoBaseUrl"`
	CoverPhotoMediaItemID string `json:"coverPhotoMediaItemId"`
	IsWriteable           bool   `json:"isWriteable"`
	TotalMediaItems       string    `json:"totalMediaItems"`
}

type Albums []*Album

func (Ω Albums) filterToTitles(titles ...string) Albums {
	if len(titles) == 0 {
		return Ω
	}
	res := Albums{}
	for _, album := range Ω {
		if stringslice.Contains(titles, album.Title) {
			res = append(res, album)
		}
	}
	return res
}
func (Ω Albums) contains(albumId string) bool {
	for _, a := range Ω {
		if a.ID == albumId {
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
		return nil, errors.Wrap(err, 0)
	}

	q := req.URL.Query()
	q.Add("pageToken", pageToken)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	defer safeClose(resp.Body, &err)

	res := &albumResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.WrapPrefix(err, "Could not decode album response", 0)
	}
	return res, nil
}
