package socialgraph

import (
	"context"
	"fmt"

	freq "github.com/NamalSanjaya/nexster/pkgs/models/friend_request"
)

const gettFriendReqEdgeQuery string = `FOR v,e IN 1..1 OUTBOUND
	@reqstorNode friendRequest
	OPTIONS { uniqueVertices: "path" }
	FILTER e.kind == "friend_request" && v._id == @friendNode
	return e._key`

type socialGraph struct {
	fReqCtrler freq.Interface
}

func NewGrphCtrler(frIntfce freq.Interface) *socialGraph {
	return &socialGraph{
		fReqCtrler: frIntfce,
	}
}

func (sgr *socialGraph) CreateFriendReq(ctx context.Context, reqstorKey, friendKey, mode, state, reqDate string) error {
	reqstorId := fmt.Sprintf("users/%s", reqstorKey)
	friendId := fmt.Sprintf("users/%s", friendKey)

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
