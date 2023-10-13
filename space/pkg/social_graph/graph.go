package socialgraph

import (
	"context"
	"fmt"
	"log"

	contapi "github.com/NamalSanjaya/nexster/pkgs/client/content_api"
	"github.com/NamalSanjaya/nexster/pkgs/models/event"
	pb "github.com/NamalSanjaya/nexster/pkgs/models/posted_by"
	"github.com/NamalSanjaya/nexster/pkgs/models/user"
	rp "github.com/NamalSanjaya/nexster/space/pkg/repository"
	tp "github.com/NamalSanjaya/nexster/space/pkg/types"
)

type socialGraph struct {
	userCtrler     user.Interface
	eventCtrler    event.Interface
	postedByCtrler pb.Interface
	conentClient   contapi.Interface
	repo           rp.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewGraph(evIntfce event.Interface, pbIntfce pb.Interface, userIntfce user.Interface, contentClient contapi.Interface,
	repoIntface rp.Interface) *socialGraph {
	return &socialGraph{
		eventCtrler:    evIntfce,
		postedByCtrler: pbIntfce,
		userCtrler:     userIntfce,
		conentClient:   contentClient,
		repo:           repoIntface,
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

func (gr *socialGraph) ListUpcomingEvents(ctx context.Context, userKey string, offset, count int) ([]*map[string]interface{}, error) {
	events, err := gr.repo.ListUpcomingEvents(ctx, offset, count)
	if err != nil {
		return []*map[string]interface{}{}, fmt.Errorf("falied to list latest events: %v", err)
	}
	results := []*map[string]interface{}{}

	for _, event := range events {
		eventKey, ok := (*event)["key"].(string)
		if !ok {
			log.Printf("[Error]: failed to convert a event key to string: eventKey=%v", (*event)["key"])
			continue
		}

		// create poster image url
		imgLink, ok := (*event)["link"].(string)
		if !ok {
			log.Printf("[Error]: failed to convert link to string: eventKey=%s", eventKey)
			continue
		}
		posterLink, err := gr.conentClient.CreateImageUrl(imgLink, contapi.Viewer)
		if err != nil {
			log.Printf("[Error]: list latest events: failed to create event poster url: eventKey=%s, %v", eventKey, err)
		}
		// if we are failed to create poster link, still we add that to our list.
		(*event)["link"] = posterLink

		// prepare postedBy field properly
		postedBy, ok := (*event)["postedBy"].([]interface{})
		if !ok {
			log.Printf("[Error]: failed to convert postedBy to []interface{}: eventKey=%s", eventKey)
			continue
		}
		postedByLn := len(postedBy)
		if postedByLn == 0 {
			log.Printf("[Error]: no owner existed for the event: eventKey=%s", eventKey)
			continue
		}
		if postedByLn > 1 {
			log.Printf("[Warn]: more than one owner exist for the event: eventKey=%s", eventKey)
			continue
		}
		(*event)["postedBy"] = postedBy[0].(map[string]interface{})

		// calculate reaction count for "love" and "going"
		reactionStates, ok := (*event)["reactionStates"].([]interface{})
		if !ok {
			log.Printf("[Error]: failed to convert reactionStates to []interface{}: eventKey=%s", eventKey)
			continue
		}
		var goingCount, loveCount int
		for _, state := range reactionStates {
			mapState, isMap := state.(map[string]interface{})
			if !isMap {
				log.Printf("[Error]: failed to convert a reaction state to map[string]interface{}: eventKey=%s", eventKey)
				continue
			}
			countFloat64, isFloat64 := mapState["count"].(float64)
			if !isFloat64 {
				log.Printf("[Error]: failed to convert a reaction state count to float64: eventKey=%s", eventKey)
				continue
			}
			going, isGoingBool := mapState["going"].(bool)
			if !isGoingBool {
				log.Printf("[Error]: failed to convert a reaction state going to bool: eventKey=%s", eventKey)
				continue
			}
			love, isLoveBool := mapState["love"].(bool)
			if !isLoveBool {
				log.Printf("[Error]: failed to convert a reaction state love to bool: eventKey=%s", eventKey)
				continue
			}
			count := int(countFloat64)
			if going {
				goingCount += count
			}
			if love {
				loveCount += count
			}
		}
		(*event)["love"] = loveCount
		(*event)["going"] = goingCount
		delete((*event), "reactionStates")

		// Get reaction key. Empty if the key is not existing
		reactionKey, err := gr.repo.GetEventReactionKey(ctx, userKey, eventKey)
		if err != nil {
			reactionKey = "none" // To indicate API user, there is an error while retriving the reactionoKey.
			log.Printf("[Error]: failed to get reaction key of viewing user: userKey=%s, eventKey=%s", userKey, eventKey)
		}
		(*event)["reactionKey"] = reactionKey
		results = append(results, event)
	}
	return results, nil
}
