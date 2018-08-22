package googlephotos

import (
	"bytes"
	"encoding/json"
	"github.com/go-errors/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type contentFilter struct {
	IncludedContentCategories []string `json:"includedContentCategories,omitempty"`
	ExcludedContentCategories []string `json:"excludedContentCategories,omitempty"`
}

type dateFilter struct {
	Dates  []ymd `json:"dates,omitempty"`
	Ranges []struct {
		StartDate ymd `json:"startDate,omitempty"`
		EndDate   ymd `json:"endDate,omitempty"`
	} `json:"ranges,omitempty"`
}

type ymd struct {
	Year  string `json:"year,omitempty"`
	Month string `json:"month,omitempty"`
	Day   string `json:"day,omitempty"`
}

type filters struct{
	ContentFilter *contentFilter `json:"contentFilter,omitempty"`
	DateFilter    *dateFilter    `json:"dateFilter,omitempty"`
}

type query struct {
	AlbumId string `json:"albumId,omitempty"`
	Filters *filters `json:"filters,omitempty"`
	PageToken string `json:"pageToken,omitempty"`
}

func safeClose(c io.Closer, err *error) {
	if closeErr := c.Close(); closeErr != nil && *err == nil {
		*err = closeErr
	}
}

const endpointAlbumList = "https://photoslibrary.googleapis.com/v1/albums"
const endpointMediaItemList = "https://photoslibrary.googleapis.com/v1/mediaItems:search"

type PhotoLibraryMedia struct {
	ID            string `json:"id,omitempty"`
	Description   string `json:"description,omitempty"`
	ProductURL    string `json:"productUrl,omitempty"`
	BaseURL       string `json:"baseUrl,omitempty"`
	MimeType      string `json:"mimeType,omitempty"`
	Filename      string `json:"filename,omitempty"`
	MediaMetadata struct {
		Width        string `json:"width,omitempty"`
		Height       string `json:"height,omitempty"`
		CreationTime time.Time `json:"creationTime,omitempty"`
		Photo        struct {
			CameraMake      string `json:"cameraMake,omitempty"`
			CameraModel     string `json:"cameraModel,omitempty"`
			FocalLength     float64 `json:"focalLength,omitempty"`
			ApertureFNumber float64 `json:"apertureFNumber,omitempty"`
			IsoEquivalent   float64 `json:"isoEquivalent,omitempty"`
			ExposureTime    string `json:"exposureTime,omitempty"`
		} `json:"photo,omitempty"`
		Video struct {
			CameraMake  string `json:"cameraMake,omitempty"`
			CameraModel string `json:"cameraModel,omitempty"`
			Fps         float64 `json:"fps,omitempty"`
			Status      string `json:"status,omitempty"`
		} `json:"video,omitempty"`
	} `json:"mediaMetadata,omitempty"`
	ContributorInfo struct {
		ProfilePictureBaseURL string `json:"profilePictureBaseUrl,omitempty"`
		DisplayName           string `json:"displayName,omitempty"`
	} `json:"contributorInfo,omitempty"`
	Location interface{} `json:"location,omitempty"`
}

type mediaPageResponse struct {
	NextPageToken string               `json:"nextPageToken"`
	Media         []*PhotoLibraryMedia `json:"mediaItems"`
}

func fetchLibraryMedia(accessToken string, filter *query) ([]*PhotoLibraryMedia, error) {
	images := []*PhotoLibraryMedia{}
	for {
		page, err := fetchMediaPage(accessToken, filter)
		if err != nil {
			return nil, err
		}
		images = append(images, page.Media...)
		if page.NextPageToken == "" {
			break
		}
		filter.PageToken = page.NextPageToken
	}
	logrus.Debugf("Resolving media fetch request with %d items", len(images))
	return images, nil
}

func fetchMediaPage(accessToken string, filter *query) (*mediaPageResponse, error) {

	if accessToken == "" {
		return nil, errors.New("Invalid access token")
	}

	filterBytes, err := json.Marshal(filter)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	logrus.WithFields(logrus.Fields{
		"requestBody":    string(filterBytes),
		"endpoint":  endpointMediaItemList,
	}).Debug("Fetching media page")

	req, err := http.NewRequest("POST", endpointMediaItemList, bytes.NewReader(filterBytes))
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	defer safeClose(resp.Body, &err)

	if resp.StatusCode != 200 {
		return nil, errors.WrapPrefix("Could not fetch media page", resp.Status, 0)
	}

	res := &mediaPageResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.WrapPrefix(err, "Could not decode media page response", 0)
	}
	return res, nil
}
