package socialgraph

import (
	"context"
	"fmt"

	frnd "github.com/NamalSanjaya/nexster/pkgs/models/friend"
	freq "github.com/NamalSanjaya/nexster/pkgs/models/friend_request"
)

const userColl string = "users" // Need to be changed once `users` repo bring to common level

const gettFriendReqEdgeQuery string = `FOR v,e IN 1..1 OUTBOUND
	@reqstorNode friendRequest
	OPTIONS { uniqueVertices: "path" }
	FILTER e.kind == "friend_request" && v._id == @friendNode
	return e._key`

type socialGraph struct {
	fReqCtrler freq.Interface
	frndCtrler frnd.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewGrphCtrler(frIntfce freq.Interface, frndIntfce frnd.Interface) *socialGraph {
	return &socialGraph{
		fReqCtrler: frIntfce,
		frndCtrler: frndIntfce,
	}
}

func (sgr *socialGraph) CreateFriendReq(ctx context.Context, reqstorKey, friendKey, mode, state, reqDate string) error {
	reqstorId := fmt.Sprintf("%s/%s", userColl, reqstorKey)
	friendId := fmt.Sprintf("%s/%s", userColl, friendKey)

	isExist, err := sgr.fReqCtrler.IsFriendReqExist(ctx, gettFriendReqEdgeQuery, map[string]interface{}{
		"reqstorNode": reqstorId,
		"friendNode":  friendId,
	})
	if err != nil {
		return fmt.Errorf("failed to get friend req from %s, to %s. Error: %v", reqstorId, friendId, err)
	}
	if isExist {
		return nil
	}
	return sgr.fReqCtrler.CreateFriendReqEdge(ctx, &freq.FriendRequest{
		From:    reqstorId,
		To:      friendId,
		Mode:    mode,
		State:   state,
		ReqDate: reqDate,
	})
}

func (sgr *socialGraph) RemoveFriendRequest(ctx context.Context, key string) error {
	return sgr.fReqCtrler.RemoveFriendReqEdge(ctx, key)
}

// ISSUES:
// 1. even if users are not exist it will create the friend link with non-existing node.
func (sgr *socialGraph) CreateFriend(ctx context.Context, friendReqKey, user1, user2, acceptedAt string) (map[string]string, error) {
	results := map[string]string{}
	// remove friend req edges
	if err := sgr.fReqCtrler.RemoveFriendReqEdge(ctx, friendReqKey); err != nil {
		return results, fmt.Errorf("error: failed to remove friend request due to %v", err)
	}
	friendId1, err := sgr.frndCtrler.CreateFriendEdge(ctx, &frnd.Friend{
		From:      fmt.Sprintf("%s/%s", userColl, user1),
		To:        fmt.Sprintf("%s/%s", userColl, user2),
		StartedAt: acceptedAt,
	})
	if err != nil {
		return results, fmt.Errorf("failed to create friend, fromUser: %s, toUser: %s due to %v", user1, user2, err)
	}

	friendId2, err := sgr.frndCtrler.CreateFriendEdge(ctx, &frnd.Friend{
		From:      fmt.Sprintf("%s/%s", userColl, user2),
		To:        fmt.Sprintf("%s/%s", userColl, user1),
		StartedAt: acceptedAt,
	})
	if err != nil {
		// remove previously created friendId1
		if err2 := sgr.frndCtrler.RemoveFriendEdge(ctx, friendId1); err2 != nil {
			return results, fmt.Errorf(`failed to delete friend, fromUser: %s, toUser: %s due to %v. 
				Uni directionaly edge will be remained`, user1, user2, err2)
		}
		return results, fmt.Errorf("failed to create friend, fromUser: %s, toUser: %s due to %v", user2, user1, err)
	}
	results["friend_id1"] = friendId1
	results["friend_id2"] = friendId2
	results["started_at"] = acceptedAt

	return results, nil
}
