package socialgraph

import (
	"context"
	"fmt"

	"github.com/NamalSanjaya/nexster/pkgs/models/event"
	pb "github.com/NamalSanjaya/nexster/pkgs/models/posted_by"
	"github.com/NamalSanjaya/nexster/pkgs/models/user"
	tp "github.com/NamalSanjaya/nexster/space/pkg/types"
)

type socialGraph struct {
	userCtrler     user.Interface
	eventCtrler    event.Interface
	postedByCtrler pb.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewGraph(evIntfce event.Interface, pbIntfce pb.Interface, userIntfce user.Interface) *socialGraph {
	return &socialGraph{
		eventCtrler:    evIntfce,
		postedByCtrler: pbIntfce,
		userCtrler:     userIntfce,
	}
}

func (gr *socialGraph) CreateEvent(ctx context.Context, userKey string, data *tp.Event) (string, string, error) {
	eventKey, err := gr.eventCtrler.CreateDocument(ctx, &event.Event{
		Link:        data.Link,
		ImgType:     data.ImgType,
		Title:       data.Title,
		Date:        data.Date,
		Description: data.Description,
		Venue:       data.Venue,
		Mode:        data.Mode,
		EventLink:   data.EventLink,
	})
	if err != nil {
		return "", "", err
	}
	postedByKey, err := gr.postedByCtrler.CreateDocument(ctx, gr.eventCtrler.MkDocumentId(eventKey), gr.userCtrler.MkUserDocId(userKey), pb.TypeEvent)
	if err != nil {
		// TODO:
		// Can remove event node as clean up.
		return "", "", fmt.Errorf("failed to create postedBy edge: %s", err)
	}

	return eventKey, postedByKey, nil
}
