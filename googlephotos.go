package googlephotos

import (
	"github.com/go-errors/errors"
	"golang.org/x/sync/errgroup"
	"gopkg.in/go-playground/validator.v9"
	"sync"
	"github.com/sirupsen/logrus"
)

type Params struct {
	OAuth2AccessToken string `validate:"required"`
	//StartAt           *time.Time
	//EndAt             *time.Time
	AlbumTitles []string
}

var validate = validator.New()

type Media struct {
	*PhotoLibraryMedia
	Albums     Albums     `json:"albums,omitempty"`
	Categories Categories `json:"categories,omitempty"`
}

type mediaSet struct {
	sync.Mutex
	m map[string]*Media
}

func (Ω *mediaSet) Add(media *PhotoLibraryMedia, album *Album, category *Category) {
	Ω.Lock()
	defer Ω.Unlock()
	if _, ok := Ω.m[media.ID]; !ok {
		Ω.m[media.ID] = &Media{
			PhotoLibraryMedia: media,
			Albums:            Albums{},
			Categories:        Categories{},
		}
	}
	Ω.m[media.ID].Albums = Ω.m[media.ID].Albums.addToSet(album)
	Ω.m[media.ID].Categories = Ω.m[media.ID].Categories.addToSet(category)
}

func (Ω *mediaSet) ToSlice(requireAlbum bool) (res []*Media) {
	for _, v := range Ω.m {
		if requireAlbum && len(v.Albums) == 0 {
			continue
		}
		res = append(res, v)
	}
	return
}

func List(params Params) ([]*Media, error) {
	if err := validate.Struct(params); err != nil {
		return nil, errors.Wrap(err, 0)
	}

	albums, err := fetchAlbums(params.OAuth2AccessToken, params.AlbumTitles...)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{"albums": len(albums)}).Infof("Received album list")

	media := mediaSet{m: map[string]*Media{}}

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
					"album": album.Title,
				}).Infof("Received media items for album")

				if err != nil {
					return err
				}
				for _, item := range items {
					media.Add(item, album, nil)
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
					"category": category,
				}).Infof("Received media items for category")
				if err != nil {
					return err
				}
				for _, item := range items {
					media.Add(item, nil, &category)
				}
				return nil
			})
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return media.ToSlice(len(params.AlbumTitles) > 0), nil

}
