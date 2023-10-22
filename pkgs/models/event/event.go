package event

import (
	"context"
	"fmt"
	"log"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	"github.com/NamalSanjaya/nexster/pkgs/errors"
	utm "github.com/NamalSanjaya/nexster/pkgs/utill/time"
	"github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
	"github.com/arangodb/go-driver"
)

const listUpcomingByDateQry = `FOR doc IN events
  FILTER DATE_TIMESTAMP(doc.date) >= DATE_NOW()
  SORT DATE_TIMESTAMP(doc.date) ASC
  LIMIT @offset, @count
  RETURN { "key": doc._key, "link": doc.link, "title": doc.title, "date": doc.date, "description": doc.description, 
  "venue": doc.venue, "mode": doc.mode, "eventLink": doc.eventLink, "createdAt": doc.createdAt }`

type eventCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*eventCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *eventCtrler {
	return &eventCtrler{argClient: argClient}
}

func (ev *eventCtrler) MkDocumentId(key string) string {
	return fmt.Sprintf("%s/%s", EventColl, key)
}

func MkEventDocId(key string) string {
	return fmt.Sprintf("%s/%s", EventColl, key)
}

func (ev *eventCtrler) CreateDocument(ctx context.Context, doc *Event) (string, error) {
	// setting default parameters
	doc.Key = uuid.GenUUID4()
	doc.Kind = kind
	doc.Visibility = visibility
	doc.CreatedAt = utm.CurrentUTCTimeTillMinutes()

	meta, err := ev.argClient.Coll.CreateDocument(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("failed to create event node: %v", err)
	}
	return meta.Key, nil
}

func (ev *eventCtrler) ListUpcomingsByDate(ctx context.Context, offset, count int) ([]*map[string]string, error) {
	return ev.listJsonValues(ctx, listUpcomingByDateQry, map[string]interface{}{
		"offset": offset,
		"count":  count,
	})
}

// Return [{}, {}, {}]. json objects can have string type of values for fields.
func (ev *eventCtrler) listJsonValues(ctx context.Context, query string, bindVars map[string]interface{}) ([]*map[string]string, error) {
	results := []*map[string]string{}
	cursor, err := ev.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return results, err
	}
	defer cursor.Close()

	for {
		var result map[string]string
		_, err := cursor.ReadDocument(ctx, &result)
		if driver.IsNoMoreDocuments(err) {
			return results, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		results = append(results, &result)
	}
}

func (ev *eventCtrler) Get(ctx context.Context, key string) (*Event, error) {
	event := &Event{}
	_, err := ev.argClient.Coll.ReadDocument(ctx, key, event)
	if driver.IsNotFoundGeneral(err) {
		return nil, errors.NewNotFoundError("document not found")
	}
	return event, err
}
