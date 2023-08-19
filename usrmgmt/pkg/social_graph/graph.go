package socialgraph

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	frnd "github.com/NamalSanjaya/nexster/pkgs/models/friend"
	freq "github.com/NamalSanjaya/nexster/pkgs/models/friend_request"
	usr "github.com/NamalSanjaya/nexster/pkgs/models/user"
)

const userColl string = "users" // Need to be changed once `users` repo bring to common level

const gettFriendReqEdgeQuery string = `FOR v,e IN 1..1 ANY
	@reqstorNode friendRequest
	OPTIONS { uniqueVertices: "path" }
	FILTER e.kind == "friend_request" && v._id == @friendNode
	return e._key`

const listFriends string = `FOR v,e IN 1..1 OUTBOUND
	@startNode friends
	OPTIONS { uniqueVertices: "path" }
	SORT e.started_at DESC
	LIMIT @offset, @count
	RETURN {
		"user_id": v._key,
		"name": v.username,
		"from_friend_id": e._key,
		"to_friend_id": e.other_friend_id,
		"image_url" : v.image_url,
		"started_at" : e.started_at,
		"headling" : v.headling
	}`

const totalFriends string = `RETURN LENGTH(
	FOR v IN 1..1 OUTBOUND @startNode friends
	  OPTIONS { uniqueVertices: "path" }
	  RETURN 1)`

const totalFriendsV2 string = `RETURN LENGTH(
    FOR f IN friends
     FILTER f._from == @startNode
     RETURN 1
)`

const getUserKey string = `FOR user IN users
	FILTER user.index_no == @indexNo
	LIMIT 1
	RETURN user._key`

const listFriendReqs string = `FOR v,e IN 1..1 INBOUND
	@userNode friendRequest
	SORT e.req_date DESC
	LIMIT @offset, @count
	RETURN { "user_key": v._key, "username" : v.username, "image_url" : v.image_url, 
	"batch": v.batch,"faculty": v.faculty, "field" : v.degree_info.field, 
	"req_date": e.req_date, "req_key": e._key }`

const allFriendReqsCountQry string = `FOR doc IN friendRequest
	FILTER doc._to == @userNode
	COLLECT WITH COUNT INTO len
	RETURN len`

type socialGraph struct {
	fReqCtrler freq.Interface
	frndCtrler frnd.Interface
	usrCtrler  usr.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewGrphCtrler(frIntfce freq.Interface, frndIntfce frnd.Interface, usrIntfce usr.Interface) *socialGraph {
	return &socialGraph{
		fReqCtrler: frIntfce,
		frndCtrler: frndIntfce,
		usrCtrler:  usrIntfce,
	}
}

func (sgr *socialGraph) ListFriendReqs(ctx context.Context, userKey string, offset, count int) ([]*map[string]string, error) {
	return sgr.fReqCtrler.ListStringValueJson(ctx, listFriendReqs, map[string]interface{}{
		"userNode": sgr.usrCtrler.MkUserDocId(userKey),
		"offset":   offset,
		"count":    count,
	})
}

func (sgr *socialGraph) GetAllFriendReqsCount(ctx context.Context, userKey string) (int, error) {
	res, err := sgr.fReqCtrler.ListStrings(ctx, allFriendReqsCountQry, map[string]interface{}{
		"userNode": sgr.usrCtrler.MkUserDocId(userKey),
	})
	if err != nil {
		return 0, err
	}
	if len(res) == 0 {
		return 0, nil
	}
	return res[0], nil
}

// TODO:
// 1. Need to check the existance of user nodes.
// 2. req_date should be system generated
func (sgr *socialGraph) CreateFriendReq(ctx context.Context, reqstorKey, friendKey, mode, state, reqDate string) (map[string]string, error) {
	results := map[string]string{}
	reqstorId := fmt.Sprintf("%s/%s", userColl, reqstorKey)
	friendId := fmt.Sprintf("%s/%s", userColl, friendKey)

	isExist, err := sgr.fReqCtrler.IsFriendReqExist(ctx, gettFriendReqEdgeQuery, map[string]interface{}{
		"reqstorNode": reqstorId,
		"friendNode":  friendId,
	})
	if err != nil {
		return results, fmt.Errorf("failed to get friend req from %s, to %s. Error: %v", reqstorId, friendId, err)
	}
	// Return Err, so that upper layer notice as resource not been created
	if isExist {
		return results, nil
	}

	if isExist, err = sgr.frndCtrler.IsFriendEdgeExist(ctx, reqstorId, friendId); err != nil {
		return results, fmt.Errorf("failed to check the existance of friend edge [from %s, to %s]. Error: %v", reqstorId, friendId, err)
	}
	if isExist {
		return results, nil
	}

	newFriendReqkey, err := sgr.fReqCtrler.CreateFriendReqEdge(ctx, &freq.FriendRequest{
		From:    reqstorId,
		To:      friendId,
		Mode:    mode,
		State:   state,
		ReqDate: reqDate,
		IsSeen:  false,
	})
	if err != nil {
		return results, fmt.Errorf("failed to create friend req [from %s, to %s]. Error: %v", reqstorId, friendId, err)
	}
	results["friend_req_id"] = newFriendReqkey
	return results, nil
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
	id1 := uuid.New().String() // Generate UUID key
	id2 := uuid.New().String()

	err := sgr.frndCtrler.CreateFriendEdge(ctx, &frnd.Friend{
		Key:           id1,
		From:          fmt.Sprintf("%s/%s", userColl, user1),
		To:            fmt.Sprintf("%s/%s", userColl, user2),
		OtherFriendId: id2,
		StartedAt:     acceptedAt,
	})
	if err != nil {
		return results, fmt.Errorf("failed to create friend, fromUser: %s, toUser: %s due to %v", user1, user2, err)
	}

	err = sgr.frndCtrler.CreateFriendEdge(ctx, &frnd.Friend{
		Key:           id2,
		From:          fmt.Sprintf("%s/%s", userColl, user2),
		To:            fmt.Sprintf("%s/%s", userColl, user1),
		OtherFriendId: id1,
		StartedAt:     acceptedAt,
	})
	if err != nil {
		// remove previously created friendId1
		if err2 := sgr.frndCtrler.RemoveFriendEdge(ctx, id1); err2 != nil {
			return results, fmt.Errorf(`failed to delete friend, fromUser: %s, toUser: %s due to %v. 
				Uni directionaly edge will be remained`, user1, user2, err2)
		}
		return results, fmt.Errorf("failed to create friend, fromUser: %s, toUser: %s due to %v", user2, user1, err)
	}
	results["friend_id1"] = id1
	results["friend_id2"] = id2
	results["started_at"] = acceptedAt

	return results, nil
}

// TODO:
// This operation should be atomic. If first one failed, whole operation should be canceled.
func (sgr *socialGraph) RemoveFriend(ctx context.Context, key1, key2 string) error {
	if err := sgr.frndCtrler.RemoveFriendEdge(ctx, key1); err != nil {
		return err
	}
	return sgr.frndCtrler.RemoveFriendEdge(ctx, key2)
}

func (sgr *socialGraph) ListFriends(ctx context.Context, userId string, offset, count int) ([]*map[string]string, error) {
	return sgr.usrCtrler.ListUsersV2(ctx, listFriends, map[string]interface{}{
		"startNode": sgr.usrCtrler.MkUserDocId(userId),
		"offset":    offset,
		"count":     count,
	})
}

func (sgr *socialGraph) CountFriends(ctx context.Context, userId string) (int, error) {
	return sgr.usrCtrler.CountUsers(ctx, totalFriends, map[string]interface{}{
		"startNode": sgr.usrCtrler.MkUserDocId(userId),
	})
}

func (sgr *socialGraph) GetRole(authUserKey, userKey string) usr.UserRole {
	if authUserKey != userKey {
		return usr.Viewer
	}
	return usr.Owner
}

// TODO: check this method again since I change the field struct format
func (sgr *socialGraph) GetProfileInfo(ctx context.Context, userKey string) (map[string]string, error) {
	info, err := sgr.usrCtrler.GetUser(ctx, userKey)
	if err != nil {
		return map[string]string{}, err
	}
	return map[string]string{
		"key": userKey, "username": info.Username, "faculty": info.Faculty, "field": info.DegreeInfo.Field, "batch": info.Batch,
		"img_url": info.ImageUrl, "about": info.About,
	}, nil
}

func (sgr *socialGraph) CountFriendsV2(ctx context.Context, userId string) (int, error) {
	return sgr.frndCtrler.CountFriends(ctx, totalFriendsV2, map[string]interface{}{
		"startNode": sgr.usrCtrler.MkUserDocId(userId),
	})
}

func (sgr *socialGraph) GetUserKeyByIndexNo(ctx context.Context, indexNo string) (string, error) {
	res, err := sgr.usrCtrler.ListStrings(ctx, getUserKey, map[string]interface{}{
		"indexNo": indexNo,
	})
	resLn := len(res)
	if resLn == 0 {
		return "", nil
	}
	if len(res) > 1 {
		return "", fmt.Errorf("indexNo=%s is not unique, array of userkeys exists", indexNo)
	}
	return *res[0], err
}
