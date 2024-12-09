package userInsight

import (
	"context"
	"fmt"
	"log"

	driver "github.com/arangodb/go-driver"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	errs "github.com/NamalSanjaya/nexster/pkgs/errors"
)

type userInsightCtrler struct {
	argClient *argdb.Client
}

var _ Interface = (*userInsightCtrler)(nil)

func NewCtrler(argClient *argdb.Client) *userInsightCtrler {
	return &userInsightCtrler{argClient: argClient}
}

func (uic *userInsightCtrler) MkUserInsightDocId(key string) string {
	return fmt.Sprintf("%s/%s", UserInsightsColl, key)
}

// To use outside places without User Insight instance
func MkUserInsightDocId(key string) string {
	return fmt.Sprintf("%s/%s", UserInsightsColl, key)

}

func (uic *userInsightCtrler) CreateUserInsight(ctx context.Context, doc *InsightData) (string, error) {
	doc.Key = fmt.Sprintf("%s-%s-%s", doc.Type, doc.UserId, doc.Year)

	_, err := uic.argClient.Coll.CreateDocument(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("failed to create user insight: %v", err)
	}

	return doc.Key, nil
}

func (uic *userInsightCtrler) GetUserInsight(ctx context.Context, key string) (*InsightData, error) {
	insight := &InsightData{}
	_, err := uic.argClient.Coll.ReadDocument(ctx, key, insight)
	if driver.IsNotFoundGeneral(err) {
		return nil, errs.NewNotFoundError(fmt.Sprintf("user insight with key=%s not found", key))
	}
	return insight, err
}

func (uic *userInsightCtrler) UpdateUserInsight(ctx context.Context, key string, updateFields map[string]interface{}) error {
	_, err := uic.argClient.Coll.UpdateDocument(ctx, key, updateFields)
	if driver.IsNotFoundGeneral(err) {
		return errs.NewNotFoundError(fmt.Sprintf("user insight with key=%s not found", key))
	}
	return err
}

func (uic *userInsightCtrler) DeleteUserInsight(ctx context.Context, key string) error {
	_, err := uic.argClient.Coll.RemoveDocument(ctx, key)
	if driver.IsNotFoundGeneral(err) {
		return errs.NewNotFoundError(fmt.Sprintf("user insight with key=%s not found", key))
	}
	return err
}

func (uic *userInsightCtrler) CountUsers(ctx context.Context, query string, bindVars map[string]interface{}) (int, error) {
	cursor, err := uic.argClient.Db.Query(ctx, query, bindVars)
	if err != nil {
		return 0, err
	}
	defer cursor.Close()

	for {
		var count int
		_, err := cursor.ReadDocument(ctx, &count)
		if driver.IsNoMoreDocuments(err) {
			return 0, nil
		} else if err != nil {
			log.Println(err)
			continue
		}
		return count, nil
	}
}

func (uic *userInsightCtrler) AppendLoginTimestamp(ctx context.Context, userKey, timestamp string) error {
	key := fmt.Sprintf("activeUser-%s-%s", userKey, timestamp[:4])
	insight, err := uic.GetUserInsight(ctx, key)
	if err != nil {
		return err
	}

	insight.LoginTimestamps = append(insight.LoginTimestamps, timestamp)

	updateFields := map[string]interface{}{
		"loginTimestamps": insight.LoginTimestamps,
	}

	err = uic.UpdateUserInsight(ctx, key, updateFields)
	if err != nil {
		return err
	}

	return nil
}

func (uic *userInsightCtrler) GetActiveUserCountForGivenTimeRange(ctx context.Context, from, to string) (int, error) {
	bindVars := map[string]interface{}{
		"from": from,
		"to":   to,
	}
	return uic.CountUsers(ctx, getActiveUserCountForGivenTimeRangeQry, bindVars)
}
