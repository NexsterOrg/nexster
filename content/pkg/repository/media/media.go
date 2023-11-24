package media

import (
	"context"
	"fmt"

	md "github.com/NamalSanjaya/nexster/pkgs/models/media"
)

type mediaRepository struct {
	ctrler md.Interface
}

var _ Interface = (*mediaRepository)(nil)

func NewRepository(mediaIntfce md.Interface) *mediaRepository {
	return &mediaRepository{
		ctrler: mediaIntfce,
	}
}

func (mr *mediaRepository) GetView(ctx context.Context, key string) (string, error) {
	data, err := mr.ctrler.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return data.Visibility, nil
}

// link attribute is unique.
func (mr *mediaRepository) GetViewForLink(ctx context.Context, link string) (string, error) {
	links, err := mr.ctrler.ListStrings(ctx, getVisibilitForLink, map[string]interface{}{
		"link": link,
	})
	if err != nil {
		return "", err
	}
	ln := len(links)
	if ln == 1 {
		return links[0], nil
	}
	if ln == 0 {
		return "", nil
	}
	return "", fmt.Errorf("found more than one document for link %s. data is in error state since link of media should be unique", link)
}
