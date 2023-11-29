package socialgraph

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	contapi "github.com/NamalSanjaya/nexster/pkgs/client/content_api"
	frnd "github.com/NamalSanjaya/nexster/pkgs/models/friend"
	freq "github.com/NamalSanjaya/nexster/pkgs/models/friend_request"
	uctr "github.com/NamalSanjaya/nexster/pkgs/models/user"
	umath "github.com/NamalSanjaya/nexster/pkgs/utill/math"
	rp "github.com/NamalSanjaya/nexster/search/pkg/repository"
	"github.com/NamalSanjaya/nexster/search/utill/algo"
)

const minAcceptanceIndex float32 = float32(0.25)
const maxSearchResults int = 20

type socialGraph struct {
	contentClient contapi.Interface
	repo          rp.Interface
	fReqCtrler    freq.Interface
	frndCtrler    frnd.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewGraph(contentClientIntfce contapi.Interface, repoIntface rp.Interface, fReqIntfce freq.Interface, frndIntfce frnd.Interface) *socialGraph {
	return &socialGraph{
		contentClient: contentClientIntfce,
		repo:          repoIntface,
		fReqCtrler:    fReqIntfce,
		frndCtrler:    frndIntfce,
	}
}

func (gr *socialGraph) SearchAmongUsers(ctx context.Context, keyword string) ([]*map[string]string, error) {
	users, err := gr.repo.ListAllUsers(ctx)
	if err != nil {
		return []*map[string]string{}, nil
	}

	keyword = strings.ToLower(keyword)
	seletedUsers := []*map[string]string{}
	selectedCount := 0
	for _, user := range users {
		score := algo.JaccardIndexBasedScore(strings.ToLower((*user)["username"]), keyword)
		if score < minAcceptanceIndex {
			continue
		}
		(*user)["score"] = strconv.FormatFloat(float64(score), 'f', -1, 32)
		seletedUsers = append(seletedUsers, user)
		selectedCount++
	}

	sort.Slice(seletedUsers, func(i, j int) bool {
		valI, err := strconv.ParseFloat((*seletedUsers[i])["score"], 32)
		if err != nil {
			return false
		}
		valJ, err := strconv.ParseFloat(((*seletedUsers[j])["score"]), 32)
		if err != nil {
			return false
		}
		return valI > valJ
	})

	selectedLastIndex := umath.Min[int](selectedCount, maxSearchResults)
	seletedUsers = seletedUsers[:selectedLastIndex]

	for _, user := range seletedUsers {
		imgUrl, err := gr.contentClient.CreateImageUrl((*user)["image_url"], contapi.Viewer)
		if err != nil {
			log.Println("failed at user search: failed to create avatar url: ", err)
			continue
		}
		(*user)["image_url"] = imgUrl
		delete(*user, "score")
	}
	return seletedUsers, nil
}

func (sgr *socialGraph) AttachFriendState(ctx context.Context, reqstorKey, friendKey string) (state string, reqId string, err error) {
	ln, err := sgr.frndCtrler.GetShortestDistance(ctx, uctr.MkUserDocId(reqstorKey), uctr.MkUserDocId(friendKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to get shortest distance: %v", err)
	}
	if ln == 1 {
		return "", "", fmt.Errorf("requestor and friend has the same key")
	}
	// already a friend
	if ln == 2 {
		return frnd.FriendType, "", nil
	}
	friendReqKey, err := sgr.fReqCtrler.GetFriendReqKey(ctx, uctr.MkUserDocId(reqstorKey), uctr.MkUserDocId(friendKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to attach friend state between %s and %s", reqstorKey, friendKey)
	}
	// pending-requestor friend
	if friendReqKey != "" {
		return frnd.PendingReqstorType, friendReqKey, nil
	}

	friendReqKey, err = sgr.fReqCtrler.GetFriendReqKey(ctx, uctr.MkUserDocId(friendKey), uctr.MkUserDocId(reqstorKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to attach friend state between %s and %s", reqstorKey, friendKey)
	}
	// pending-recipient friend
	if friendReqKey != "" {
		return frnd.PendingRecipientType, friendReqKey, nil
	}
	// not a friend
	return frnd.NotFriendType, "", nil
}
