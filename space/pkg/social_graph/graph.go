package socialgraph

import (
	"context"
	"fmt"
	"log"

	contapi "github.com/NamalSanjaya/nexster/pkgs/client/content_api"
	"github.com/NamalSanjaya/nexster/pkgs/errors"
	"github.com/NamalSanjaya/nexster/pkgs/models/event"
	erec "github.com/NamalSanjaya/nexster/pkgs/models/event_reaction"
	pb "github.com/NamalSanjaya/nexster/pkgs/models/posted_by"
	"github.com/NamalSanjaya/nexster/pkgs/models/user"
	rp "github.com/NamalSanjaya/nexster/space/pkg/repository"
	tp "github.com/NamalSanjaya/nexster/space/pkg/types"
)

type socialGraph struct {
	eventCtrler    event.Interface
	postedByCtrler pb.Interface
	userCtrler     user.Interface
	reactionCtrler erec.Interface
	conentClient   contapi.Interface
	repo           rp.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewGraph(evIntfce event.Interface, pbIntfce pb.Interface, userIntfce user.Interface, rectIntfce erec.Interface, contentClient contapi.Interface,
	repoIntface rp.Interface) *socialGraph {
	return &socialGraph{
		eventCtrler:    evIntfce,
		postedByCtrler: pbIntfce,
		userCtrler:     userIntfce,
		reactionCtrler: rectIntfce,
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
		parsedEvent, err := gr.parseEventInfo(ctx, userKey, event)
		if err != nil {
			log.Println("[Error]: ", err)
			continue
		}
		results = append(results, &parsedEvent)
	}
	return results, nil
}

// List events for given user
func (gr *socialGraph) ListMyEvents(ctx context.Context, userKey string, offset, count int) ([]*map[string]interface{}, error) {
	events, err := gr.repo.ListEventsForUser(ctx, userKey, offset, count)
	if err != nil {
		return []*map[string]interface{}{}, fmt.Errorf("falied to list latest events: %v", err)
	}
	results := []*map[string]interface{}{}

	for _, event := range events {
		parsedEvent, err := gr.parseEventInfo(ctx, userKey, event)
		if err != nil {
			log.Println("[Error]: ", err)
			continue
		}
		results = append(results, &parsedEvent)
	}
	return results, nil
}

func (gr *socialGraph) parseEventInfo(ctx context.Context, userKey string, event *map[string]interface{}) (map[string]interface{}, error) {
	emptyResult := map[string]interface{}{}
	eventKey, ok := (*event)["key"].(string)
	if !ok {
		return emptyResult, fmt.Errorf("failed to convert a event key to string: eventKey=%v", (*event)["key"])
	}

	// create poster image url
	imgLink, ok := (*event)["link"].(string)
	if !ok {
		return emptyResult, fmt.Errorf("failed to convert link to string: eventKey=%s", eventKey)
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
		return emptyResult, fmt.Errorf("failed to convert postedBy to []interface{}: eventKey=%s", eventKey)
	}
	postedByLn := len(postedBy)
	if postedByLn == 0 {
		return emptyResult, fmt.Errorf("no owner existed for the event: eventKey=%s", eventKey)
	}
	if postedByLn > 1 {
		return emptyResult, fmt.Errorf("more than one owner exist for the event: eventKey=%s", eventKey)
	}
	owner, ok := postedBy[0].(map[string]interface{})
	if !ok {
		return emptyResult, fmt.Errorf("failed to convert owner info to map[string]interface{}: eventKey=%s", eventKey)
	}
	ownerKey, isStr := owner["key"].(string)
	if !isStr {
		return emptyResult, fmt.Errorf("failed to convert owner key to string: eventKey=%s", eventKey)
	}
	(*event)["postedBy"] = owner

	if ownerKey == userKey {
		(*event)["perm"] = "owner"
	} else {
		(*event)["perm"] = "viewer"
	}
	// calculate reaction count for "love" and "going"
	reactionStates, ok := (*event)["reactionStates"].([]interface{})
	if !ok {
		return emptyResult, fmt.Errorf("failed to convert reactionStates to []interface{}: eventKey=%s", eventKey)
	}
	var goingCount, loveCount int
	for _, state := range reactionStates {
		mapState, isMap := state.(map[string]interface{})
		if !isMap {
			return emptyResult, fmt.Errorf("failed to convert a reaction state to map[string]interface{}: eventKey=%s", eventKey)
		}
		countFloat64, isFloat64 := mapState["count"].(float64)
		if !isFloat64 {
			return emptyResult, fmt.Errorf("failed to convert a reaction state count to float64: eventKey=%s", eventKey)
		}
		going, isGoingBool := mapState["going"].(bool)
		if !isGoingBool {
			return emptyResult, fmt.Errorf("failed to convert a reaction state going to bool: eventKey=%s", eventKey)
		}
		love, isLoveBool := mapState["love"].(bool)
		if !isLoveBool {
			return emptyResult, fmt.Errorf("failed to convert a reaction state love to bool: eventKey=%s", eventKey)
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
	reaction, err := gr.repo.GetEventReaction(ctx, userKey, eventKey)
	if err != nil {
		log.Printf("[Error]: failed to get reaction key of viewing user: userKey=%s, eventKey=%s", userKey, eventKey)
	}
	(*event)["reaction"] = reaction
	return *event, nil
}

func (gr *socialGraph) GetEvent(ctx context.Context, userKey, eventKey string) (map[string]interface{}, error) {
	event, err := gr.repo.GetEvent(ctx, eventKey)
	if err != nil {
		return event, err
	}
	return gr.parseEventInfo(ctx, userKey, &event)
}

// typ -> love, going
func (gr *socialGraph) ListEventReactUsersForType(ctx context.Context, eventKey, typ string, offset, count int) ([]*map[string]interface{}, error) {
	results := []*map[string]interface{}{}
	var err error
	if typ == "love" {
		results, err = gr.repo.ListEventLovers(ctx, eventKey, offset, count)
	} else if typ == "going" {
		results, err = gr.repo.ListEventAttendees(ctx, eventKey, offset, count)
	} else {
		return results, fmt.Errorf("invalid type is given for typ parameter")
	}

	if err != nil {
		return results, err
	}

	for _, result := range results {
		imgLink, ok := (*result)["imageUrl"].(string)
		if !ok {
			log.Printf("[Error]: failed to convert imageUrl to string: eventKey=%s, userKey=%v", eventKey, (*result)["key"])
			continue
		}
		imgLink, err = gr.conentClient.CreateImageUrl(imgLink, contapi.Viewer)
		if err != nil {
			log.Printf("[Error]: failed to create avatar url: eventKey=%s, userKey=%v", eventKey, (*result)["key"])
			continue
		}
		(*result)["imageUrl"] = imgLink
	}
	return results, nil
}

func (gr *socialGraph) GetEventOwnerKey(ctx context.Context, eventKey string) (string, error) {
	return gr.repo.GetEventOwnerKey(ctx, eventKey)
}

func (sgr *socialGraph) GetRole(authUserKey, userKey string) user.UserRole {
	if authUserKey != userKey {
		return user.Viewer
	}
	return user.Owner
}

func (sgr *socialGraph) CreateEventReactionEdge(ctx context.Context, reactorKey, eventKey string, data *tp.EventReaction) (string, error) {
	// check the event exist
	_, err := sgr.eventCtrler.Get(ctx, eventKey)
	if err != nil {
		return "", err
	}
	// check any existing event reaction edge
	curReactKey, err := sgr.repo.GetKeyOfUserReaction(ctx, eventKey, reactorKey)
	if err != nil {
		return "", err
	}
	// event reaction edge already exists
	if curReactKey != "" {
		return "", errors.NewConflictError("event reaction edge already exists")
	}
	return sgr.reactionCtrler.Create(ctx, &erec.EventReaction{
		From:  sgr.userCtrler.MkUserDocId(reactorKey),
		To:    sgr.eventCtrler.MkDocumentId(eventKey),
		Love:  data.Love,
		Going: data.Going,
	})
}

func (sgr *socialGraph) SetEventReactionState(ctx context.Context, reactorKey, reactionEdgeKey string, data map[string]bool) error {
	edgeOwner, err := sgr.reactionCtrler.Get(ctx, reactionEdgeKey)
	if err != nil {
		return err
	}
	if edgeOwner.From != sgr.userCtrler.MkUserDocId(reactorKey) {
		return errors.NewUnAuthError("not allowed to access")
	}
	return sgr.reactionCtrler.UpdateState(ctx, reactionEdgeKey, data)
}

// TODO:
// Not found & Unauthorized access errors need to handle properly.
func (sgr *socialGraph) DeleteEvent(ctx context.Context, userKey, eventKey string) error {
	var err error
	if err = sgr.repo.DelPostedByGivenFromAndTo(ctx, event.MkEventDocId(eventKey), user.MkUserDocId(userKey)); err != nil {
		return fmt.Errorf("failed to remove postedBy edge: %v", err)
	}
	deletedEventDoc, err := sgr.eventCtrler.Delete(ctx, eventKey)
	if err != nil {
		return fmt.Errorf("failed to remove event node: %v", err)
	}

	// delete image from azure blob storage
	link, ok := (*deletedEventDoc)["link"].(string) // eg: link = event-posters/3818349.png
	if !ok {
		return fmt.Errorf("failed to delete event poster from azure blob storage: unable to find image link")
	}
	if err = sgr.conentClient.DeleteImage(ctx, link); err != nil {
		return fmt.Errorf("failed to delete event poster from azure blob storage: %v", err)
	}
	return nil
}
