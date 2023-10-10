package socialgraph

import (
	"context"
	"fmt"
	"log"

	contapi "github.com/NamalSanjaya/nexster/pkgs/client/content_api"
	contentapi "github.com/NamalSanjaya/nexster/pkgs/client/content_api"
	"github.com/NamalSanjaya/nexster/pkgs/models/event"
	pb "github.com/NamalSanjaya/nexster/pkgs/models/posted_by"
	"github.com/NamalSanjaya/nexster/pkgs/models/user"
	tp "github.com/NamalSanjaya/nexster/space/pkg/types"
)

type socialGraph struct {
	userCtrler     user.Interface
	eventCtrler    event.Interface
	postedByCtrler pb.Interface
	conentClient   contapi.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewGraph(evIntfce event.Interface, pbIntfce pb.Interface, userIntfce user.Interface, contentClient contapi.Interface) *socialGraph {
	return &socialGraph{
		eventCtrler:    evIntfce,
		postedByCtrler: pbIntfce,
		userCtrler:     userIntfce,
		conentClient:   contentClient,
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

func (gr *socialGraph) ListLatestEvents(ctx context.Context, offset, count int) ([]*map[string]string, error) {
	events, err := gr.eventCtrler.ListUpcomingsByDate(ctx, offset, count)
	if err != nil {
		return []*map[string]string{}, fmt.Errorf("falied to list latest events: %v", err)
	}
	for _, event := range events {
		posterLink, err := gr.conentClient.CreateImageUrl((*event)["link"], contentapi.Viewer)
		if err != nil {
			log.Println("list latest events: failed to create event poster url: ", err)
			continue
		}
		// If we are failed to create poster link, still we add that to our list.
		(*event)["link"] = posterLink
	}
	return events, nil
}
