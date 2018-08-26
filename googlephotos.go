package googlephotos

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
)

type Params struct {
	OAuth2AccessToken string `validate:"required"`
	//StartAt           *time.Time
	//EndAt             *time.Time
	AlbumTitles []string
}

var validate = validator.New()

type Image struct {
	*PhotoLibraryMedia
	Albums     Albums     `json:"albums,omitempty"`
	Categories Categories `json:"categories,omitempty"`
}

type mediaSet struct {
	sync.Mutex
	m map[string]*Image
}

func (Ω *mediaSet) add(media *PhotoLibraryMedia, album *Album, category *Category) {
	Ω.Lock()
	defer Ω.Unlock()
	if _, ok := Ω.m[media.ID]; !ok {
		Ω.m[media.ID] = &Image{
			PhotoLibraryMedia: media,
			Albums:            Albums{},
			Categories:        Categories{},
		}
	}
	Ω.m[media.ID].Albums = Ω.m[media.ID].Albums.addToSet(album)
	Ω.m[media.ID].Categories = Ω.m[media.ID].Categories.addToSet(category)
}

func (Ω *mediaSet) toSlice(requireAlbum bool) (res []*Image) {
	for _, v := range Ω.m {
		if requireAlbum && len(v.Albums) == 0 {
			continue
		}
		res = append(res, v)
	}
	return
}

func Download(list []*Image, filePath string) error {

	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		return errors.Wrap(err, 0)
	}
	eg := errgroup.Group{}
	eg.Go(func() error {
		for _, _img := range list {
			img := _img
			eg.Go(func() (resErr error) {

				imgResponse, err := http.Get(img.BaseURL)
				if err != nil {
					return errors.Wrap(err, 0)
				}
				defer safeClose(imgResponse.Body, &resErr)

				imgFile, err := os.Create(path.Join(filePath, fmt.Sprintf("%s.jpg", img.ID)))
				if err != nil {
					return errors.Wrap(err, 0)
				}
				defer safeClose(imgFile, &resErr)

				_, err = io.Copy(imgFile, imgResponse.Body)
				if err != nil {
					return errors.Wrap(err, 0)
				}

				return nil
			})
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, 0)
	}

	return nil

}

// FetchList returns a list of images from Google Photos.
func FetchList(params Params) ([]*Image, error) {
	if err := validate.Struct(params); err != nil {
		return nil, errors.Wrap(err, 0)
	}

	albums, err := fetchAlbums(params.OAuth2AccessToken, params.AlbumTitles...)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{"albums": len(albums)}).Infof("Received album list")

	media := mediaSet{m: map[string]*Image{}}

	g := errgroup.Group{}
	g.Go(func() error {

		for _, _album := range albums {
			album := _album
			g.Go(func() error {
				items, err := fetchLibraryMedia(params.OAuth2AccessToken, &query{
					AlbumId: album.ID,
				})

				logrus.WithFields(logrus.Fields{
					"mediaItems": len(items),
					"album":      album.Title,
				}).Infof("Received media items for album")

				if err != nil {
					return err
				}
				for _, item := range items {
					media.add(item, album, nil)
				}
				return nil
			})
		}

		for _, _category := range knownCategories {
			category := _category
			g.Go(func() error {
				items, err := fetchLibraryMedia(params.OAuth2AccessToken, &query{
					Filters: &filters{
						ContentFilter: &contentFilter{
							IncludedContentCategories: []string{string(category)},
						},
					},
				})
				logrus.WithFields(logrus.Fields{
					"mediaItems": len(items),
					"category":   category,
				}).Infof("Received media items for category")
				if err != nil {
					return err
				}
				for _, item := range items {
					media.add(item, nil, &category)
				}
				return nil
			})
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return media.toSlice(len(params.AlbumTitles) > 0), nil

}
