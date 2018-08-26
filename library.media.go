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

type filters struct {
	ContentFilter *contentFilter `json:"contentFilter,omitempty"`
	DateFilter    *dateFilter    `json:"dateFilter,omitempty"`
}

type query struct {
	AlbumId   string   `json:"albumId,omitempty"`
	Filters   *filters `json:"filters,omitempty"`
	PageToken string   `json:"pageToken,omitempty"`
}

func safeClose(c io.Closer, err *error) {
	if closeErr := c.Close(); closeErr != nil && *err == nil {
		*err = closeErr
	}
}

const endpointAlbumList = "https://photoslibrary.googleapis.com/v1/albums"
const endpointMediaItemList = "https://photoslibrary.googleapis.com/v1/mediaItems:search"

// PhotoLibraryMedia is the representation of a photo or video in Google Photos.
type PhotoLibraryMedia struct {
	// Identifier for the media item. This is a persistent identifier that can be used between sessions to identify this media item.
	ID string `json:"id,omitempty"`
	// Description of the media item. This is shown to the user in the item's info section in the Google Photos app.
	Description string `json:"description,omitempty"`
	// Google Photos URL for the media item. This link is available to the user only if they're signed in.
	ProductURL string `json:"productUrl,omitempty"`
	// A URL to the media item's bytes. This shouldn't be used directly to access the media item.
	// For example, '=w2048-h1024' will set the dimensions of a media item of type photo to have a width of 2048 px and height of 1024 px.
	BaseURL string `json:"baseUrl,omitempty"`
	// MIME type of the media item. For example, image/jpeg.
	MimeType string `json:"mimeType,omitempty"`
	// Filename of the media item. This is shown to the user in the item's info section in the Google Photos app.
	Filename string `json:"filename,omitempty"`
	// Metadata related to the media item, such as, height, width, or creation time.
	MediaMetadata struct {
		Width        string    `json:"width,omitempty"`
		Height       string    `json:"height,omitempty"`
		CreationTime time.Time `json:"creationTime,omitempty"`
		Photo        struct {
			CameraMake      string  `json:"cameraMake,omitempty"`
			CameraModel     string  `json:"cameraModel,omitempty"`
			FocalLength     float64 `json:"focalLength,omitempty"`
			ApertureFNumber float64 `json:"apertureFNumber,omitempty"`
			IsoEquivalent   float64 `json:"isoEquivalent,omitempty"`
			ExposureTime    string  `json:"exposureTime,omitempty"`
		} `json:"photo,omitempty"`
		Video struct {
			CameraMake  string  `json:"cameraMake,omitempty"`
			CameraModel string  `json:"cameraModel,omitempty"`
			Fps         float64 `json:"fps,omitempty"`
			Status      string  `json:"status,omitempty"`
		} `json:"video,omitempty"`
	} `json:"mediaMetadata,omitempty"`
	// Information about the user who created this media item.
	ContributorInfo struct {
		ProfilePictureBaseURL string `json:"profilePictureBaseUrl,omitempty"`
		DisplayName           string `json:"displayName,omitempty"`
	} `json:"contributorInfo,omitempty"`
	// Not yet available. Location of the media item.
	Location interface{} `json:"location,omitempty"` // Current not returned by Google Photos
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
		"requestBody": string(filterBytes),
		"endpoint":    endpointMediaItemList,
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
