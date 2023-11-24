package social_graph

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	contapi "github.com/NamalSanjaya/nexster/pkgs/client/content_api"
	fcrepo "github.com/NamalSanjaya/nexster/pkgs/models/faculty"
	frnd "github.com/NamalSanjaya/nexster/pkgs/models/friend"
	freq "github.com/NamalSanjaya/nexster/pkgs/models/friend_request"
	mrepo "github.com/NamalSanjaya/nexster/pkgs/models/media"
	mo "github.com/NamalSanjaya/nexster/pkgs/models/media_owner"
	rrepo "github.com/NamalSanjaya/nexster/pkgs/models/reaction"
	urepo "github.com/NamalSanjaya/nexster/pkgs/models/user"
	tp "github.com/NamalSanjaya/nexster/timeline/pkg/types"
)

const perMonth float64 = (24 * 30)
const (
	male           string = "male"
	female         string = "female"
	birthdayLayout string = "2006-01-02" // yy-mm-dd
)

// TODO
// 1. Change collection names, field names and other parameter names (eg: friends, mediaOwnerEdges)

// TODO: For users/482201 case return wrong results.
const recentMediaQuery string = `FOR v,e IN 1..2 INBOUND @userNode friends, mediaOwnerEdges
	FILTER e.kind == "media_owner" && v.visibility == @visibility
	&& v.created_date < DATE_ISO8601(@lastPostAt)
	SORT v.created_date DESC
	LIMIT @noOfPosts
	RETURN DISTINCT {"media": {"_key": v._key, "link" : v.link, "title" : v.title, 
	"description" : v.description,"created_date" : v.created_date, "size" : v.size}, "owner_id": e._to}`

const order2FriendsQuery string = `FOR v,e IN 2..2 OUTBOUND
	@userNode friends
	OPTIONS { uniqueVertices: "path" }
	COLLECT key = v._key INTO groups
	SORT null
	SORT groups[0].e.started_at
	RETURN {"key" : key, "username" : groups[0].v.username, "image_url": groups[0].v.image_url, "faculty": groups[0].v.faculty, 
	"field": groups[0].v.field, "batch": groups[0].v.batch }`

const getOrder1FriendsQuery string = `FOR v,e IN 1..1 OUTBOUND
	@userNode friends
	RETURN v._key`

const getReactionQuery string = `FOR v,e IN 1..1 INBOUND @mediaNode reactions
    RETURN { "like": e["like"], "love": e.love, "laugh": e.laugh,
    "sad": e.sad, "insightful": e.insightful }`

const getOwnersMediaQuery string = `FOR v,e IN 1..1 INBOUND @userNode mediaOwnerEdges
	FILTER v.created_date <= DATE_ISO8601(@lastPostAt)
	SORT v.created_date DESC
	LIMIT @noOfPosts
	RETURN DISTINCT {"_key": v._key, "link" : v.link, "title" : v.title, 
	"description" : v.description,"created_date" : v.created_date, "size" : v.size}`

const getViewerReactions string = `FOR r IN reactions
	FILTER r._from == @fromUser AND r._to == @toMedia
	LIMIT 1
	RETURN r`

const getAllMedia string = `FOR v IN 1..1 INBOUND
	@userNode mediaOwnerEdges
	SORT v.created_date DESC
	LIMIT @offset, @count
	RETURN {"key": v._key, "image_url": v.link}`

const listPublicMediaQuery string = `FOR v IN 1..1 INBOUND
	@userNode mediaOwnerEdges
	FILTER v.visibility == "public"
	SORT v.created_date DESC
	LIMIT @offset, @count
	RETURN {"key": v._key, "image_url": v.link}`

const getUserKey string = `FOR user IN users
	FILTER user.index_no == @indexNo
	LIMIT 1
	RETURN user._key`

// TODO: Add FILTER v._key != userKey
const listUsersBasedOnGenderQry = `FOR v IN 1..1 INBOUND @genderId hasGender
	FILTER v._key != @userKey
	RETURN  { "key": v._key, "username": v.username, "image_url": v.image_url, "batch": v.batch, 
		"field": v.field, "faculty": v.faculty, "birthday" : v.birthday, "gender" : v.gender}`

type socialGraph struct {
	mediaRepo     mrepo.Interface
	userRepo      urepo.Interface
	reactRepo     rrepo.Interface
	facRepo       fcrepo.Interface
	fReqCtrler    freq.Interface
	frndCtrler    frnd.Interface
	mdOwnerCtrler mo.Interface
	conentClient  contapi.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewRepo(mIntfce mrepo.Interface, uIntfce urepo.Interface, rIntfce rrepo.Interface, facIntfce fcrepo.Interface,
	frIntfce freq.Interface, frndIntfce frnd.Interface, mdOwnerIntfce mo.Interface, contentClient contapi.Interface) *socialGraph {
	return &socialGraph{
		mediaRepo:     mIntfce,
		userRepo:      uIntfce,
		reactRepo:     rIntfce,
		facRepo:       facIntfce,
		fReqCtrler:    frIntfce,
		frndCtrler:    frndIntfce,
		mdOwnerCtrler: mdOwnerIntfce,
		conentClient:  contentClient,
	}
}

func (sgr *socialGraph) ListRecentPosts(ctx context.Context, userId, lastPostTimestamp, visibility string, noOfPosts int) ([]*map[string]interface{}, error) {
	posts := []*map[string]interface{}{}
	bindVars := map[string]interface{}{
		"userNode":   sgr.userRepo.MkUserDocId(userId),
		"lastPostAt": lastPostTimestamp,
		"noOfPosts":  noOfPosts,
		"visibility": visibility,
	}
	medias, err := sgr.mediaRepo.ListMediaWithOwner(ctx, recentMediaQuery, bindVars)
	if err != nil {
		return posts, err
	}
	prefixLn := len(urepo.UsersColl) + 1 // length of "users/"

	for _, media := range medias {
		user, err2 := sgr.userRepo.GetUser(ctx, media.OwnerId[prefixLn:])
		if err2 != nil {
			log.Println(err2)
			continue
		}

		racts, err2 := sgr.reactRepo.GetReactionsCount(ctx, getReactionQuery, map[string]interface{}{
			"mediaNode": sgr.mediaRepo.MkMediaDocId(media.Media.Key),
		})
		if err2 != nil {
			log.Println(err2)
			continue
		}

		viewersReacts, err2 := sgr.reactRepo.GetViewersReactions(ctx, getViewerReactions, map[string]interface{}{
			"fromUser": sgr.userRepo.MkUserDocId(userId), "toMedia": sgr.mediaRepo.MkMediaDocId(media.Media.Key),
		})
		if err2 != nil {
			log.Println(err2)
			continue
		}
		permission := sgr.conentClient.GetPermission(user.UserId, userId)
		mediaLink, err := sgr.conentClient.CreateImageUrl(media.Media.Link, permission)
		if err != nil {
			log.Println("failed to create post url: ", err)
			continue
		}
		media.Media.Link = mediaLink

		imgUrl, err := sgr.conentClient.CreateImageUrl(user.ImageUrl, permission)
		if err != nil {
			log.Println("failed to create post url: ", err)
			continue
		}

		posts = append(posts, &map[string]interface{}{
			"media": media.Media, "owner": map[string]string{"_key": user.UserId, "name": user.Username, "Headling": user.Headling, "image_url": imgUrl},
			"reactions": racts, "viewer_reaction": map[string]interface{}{"key": viewersReacts.Key, "like": viewersReacts.Like, "love": viewersReacts.Love,
				"laugh": viewersReacts.Laugh},
		})
	}

	return posts, nil
}

func (sgr *socialGraph) ListOwnersPosts(ctx context.Context, userKey, lastPostTimestamp string, noOfPosts int) ([]*map[string]interface{}, error) {
	posts := []*map[string]interface{}{}
	bindVars := map[string]interface{}{
		"userNode":   sgr.userRepo.MkUserDocId(userKey),
		"lastPostAt": lastPostTimestamp,
		"noOfPosts":  noOfPosts,
	}
	medias, err := sgr.mediaRepo.ListMedia(ctx, getOwnersMediaQuery, bindVars)
	if err != nil {
		return posts, err
	}

	for _, media := range medias {
		racts, err2 := sgr.reactRepo.GetReactionsCount(ctx, getReactionQuery, map[string]interface{}{
			"mediaNode": sgr.mediaRepo.MkMediaDocId(media.Key),
		})
		if err2 != nil {
			log.Println(err2)
			continue
		}

		mediaLink, err := sgr.conentClient.CreateImageUrl(media.Link, contapi.Owner)
		if err != nil {
			log.Println("failed to create owner post url: ", err)
			continue
		}
		media.Link = mediaLink

		posts = append(posts, &map[string]interface{}{
			"media":     media,
			"reactions": racts,
		})
	}

	return posts, nil
}

func (sgr *socialGraph) ListFriendSuggestions(ctx context.Context, userId string, offset, count int) ([]*map[string]string, error) {
	if offset < 0 || count <= 0 {
		return []*map[string]string{}, nil
	}
	userDocId := sgr.userRepo.MkUserDocId(userId)
	order2Nodes, err := sgr.userRepo.ListUsersV2(ctx, order2FriendsQuery, map[string]interface{}{
		"userNode": userDocId,
	})
	if err != nil {
		return order2Nodes, fmt.Errorf("falied to list 2nd order user info due to %v", err)
	}
	order1Friends, err := sgr.userRepo.ListStrings(ctx, getOrder1FriendsQuery, map[string]interface{}{"userNode": userDocId})
	if err != nil {
		return order2Nodes, fmt.Errorf("failed to list 1st order users due to %v", err)
	}
	results := []*map[string]string{}
	resultsCount := 0
	// remove 1st order nodes from 2nd order nodes
	for _, node2 := range order2Nodes {
		notFound := true
		for _, key1 := range order1Friends {
			if *key1 == (*node2)["key"] {
				notFound = false
				break
			}
		}
		if notFound {

			imgUrl, err := sgr.conentClient.CreateImageUrl((*node2)["image_url"], contapi.Viewer)
			if err != nil {
				log.Println("failed to create url: ", err)
				continue
			}
			(*node2)["image_url"] = imgUrl

			results = append(results, node2)
			resultsCount++
		}
	}
	if offset >= resultsCount {
		return []*map[string]string{}, nil
	}
	endIndex := offset + count
	if endIndex > resultsCount {
		endIndex = resultsCount
	}
	return results[offset:endIndex], nil
}

func (sgr *socialGraph) UpdateMediaReaction(ctx context.Context, fromUserKey, toMediaKey, key string, newDoc map[string]interface{}) (string, error) {
	return sgr.reactRepo.UpdateReactions(ctx, sgr.userRepo.MkUserDocId(fromUserKey), sgr.mediaRepo.MkMediaDocId(toMediaKey), key, newDoc)
}

func (sgr *socialGraph) CreateMediaReaction(ctx context.Context, fromUserKey, toMediaKey string, newDoc map[string]interface{}) (string, error) {
	// first check whether a there is a link or not
	viewersReacts, err := sgr.reactRepo.GetViewersReactions(ctx, getViewerReactions, map[string]interface{}{
		"fromUser": sgr.userRepo.MkUserDocId(fromUserKey), "toMedia": sgr.mediaRepo.MkMediaDocId(toMediaKey),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create reaction link due to %v", err)
	}
	// Already there is a key
	if viewersReacts.Key != "" {
		return sgr.reactRepo.UpdateReactions(ctx, sgr.userRepo.MkUserDocId(fromUserKey), sgr.mediaRepo.MkMediaDocId(toMediaKey), viewersReacts.Key, newDoc)
	}
	return sgr.reactRepo.CreateReactionLink(ctx, sgr.userRepo.MkUserDocId(fromUserKey), sgr.mediaRepo.MkMediaDocId(toMediaKey), newDoc)
}

func (sgr *socialGraph) GetRole(authUserKey, userKey string) urepo.UserRole {
	if authUserKey != userKey {
		return urepo.Viewer
	}
	return urepo.Owner
}

func (sgr *socialGraph) ListAllMedia(ctx context.Context, userKey string, offset, count int) ([]*map[string]string, error) {
	medias, err := sgr.mediaRepo.ListMediaWithCustomFields(ctx, getAllMedia, map[string]interface{}{
		"userNode": sgr.userRepo.MkUserDocId(userKey),
		"offset":   offset,
		"count":    count,
	})
	if err != nil {
		return []*map[string]string{}, err
	}

	for _, media := range medias {
		imgUrl, err := sgr.conentClient.CreateImageUrl((*media)["image_url"], contapi.Owner)
		if err != nil {
			log.Println("failed at owner post listing: failed to create post url: ", err)
			continue
		}
		(*media)["image_url"] = imgUrl
	}
	return medias, nil
}

func (sgr *socialGraph) ListPublicMedia(ctx context.Context, userKey string, offset, count int) ([]*map[string]string, error) {
	medias, err := sgr.mediaRepo.ListMediaWithCustomFields(ctx, listPublicMediaQuery, map[string]interface{}{
		"userNode": sgr.userRepo.MkUserDocId(userKey),
		"offset":   offset,
		"count":    count,
	})
	if err != nil {
		return []*map[string]string{}, err
	}

	for _, media := range medias {
		imgUrl, err := sgr.conentClient.CreateImageUrl((*media)["image_url"], contapi.Viewer)
		if err != nil {
			log.Println("failed at owner post listing: failed to create post url: ", err)
			continue
		}
		(*media)["image_url"] = imgUrl
	}
	return medias, nil
}

func (sgr *socialGraph) GetUserKeyByIndexNo(ctx context.Context, indexNo string) (string, error) {
	res, err := sgr.userRepo.ListStrings(ctx, getUserKey, map[string]interface{}{
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

func (sgr *socialGraph) AttachFriendState(ctx context.Context, reqstorKey, friendKey string) (state string, reqId string, err error) {
	ln, err := sgr.frndCtrler.GetShortestDistance(ctx, sgr.userRepo.MkUserDocId(reqstorKey), sgr.userRepo.MkUserDocId(friendKey))
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
	friendReqKey, err := sgr.fReqCtrler.GetFriendReqKey(ctx, sgr.userRepo.MkUserDocId(reqstorKey), sgr.userRepo.MkUserDocId(friendKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to attach friend state between %s and %s", reqstorKey, friendKey)
	}
	// pending-requestor friend
	if friendReqKey != "" {
		return frnd.PendingReqstorType, friendReqKey, nil
	}

	friendReqKey, err = sgr.fReqCtrler.GetFriendReqKey(ctx, sgr.userRepo.MkUserDocId(friendKey), sgr.userRepo.MkUserDocId(reqstorKey))
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

// Suggest friends based on the birthday and faculty mapping.
// TODO: owner can be within the list as well.
func (sgr *socialGraph) ListFriendSuggsV2(ctx context.Context, userKey, birthday, faculty, gender string, page, pageSize int) ([]*map[string]string, error) {
	facWithGender, err := sgr.facRepo.GetFaculty(ctx, faculty, gender)
	if err != nil {
		return []*map[string]string{}, fmt.Errorf("failed to get faculty info: %v", err)
	}
	// prefer gender
	pfGender := preferGender(gender)
	pfUsers, err := sgr.userRepo.ListUsersV2(ctx, listUsersBasedOnGenderQry, map[string]interface{}{
		"genderId": fmt.Sprintf("genders/%s", pfGender),
		"userKey":  userKey,
	})
	if err != nil {
		return []*map[string]string{}, fmt.Errorf("failed to list %s gender users: %v", pfGender, err)
	}
	for _, pfUser := range pfUsers {
		score := sgr.facRepo.GetPriority(strings.ToLower((*pfUser)["faculty"]), pfGender, facWithGender) * ageMatch((*pfUser)["birthday"], birthday, pfGender, gender)
		(*pfUser)["score"] = strconv.Itoa(score)
	}
	sort.Slice(pfUsers, func(i, j int) bool {
		valI, err := strconv.Atoi((*pfUsers[i])["score"])
		if err != nil {
			return false
		}
		valJ, err := strconv.Atoi((*pfUsers[j])["score"])
		if err != nil {
			return false
		}
		return valI > valJ
	})

	// same gender
	otherUsers, err := sgr.userRepo.ListUsersV2(ctx, listUsersBasedOnGenderQry, map[string]interface{}{
		"genderId": fmt.Sprintf("genders/%s", gender),
		"userKey":  userKey,
	})
	if err != nil {
		return []*map[string]string{}, fmt.Errorf("failed to list %s gender users: %v", gender, err)
	}
	for _, otherUser := range otherUsers {
		score := sgr.facRepo.GetPriority(strings.ToLower((*otherUser)["faculty"]), gender, facWithGender) * ageMatch((*otherUser)["birthday"], birthday, gender, gender)
		(*otherUser)["score"] = strconv.Itoa(score)
	}
	sort.Slice(otherUsers, func(i, j int) bool {
		valI, err := strconv.Atoi((*otherUsers[i])["score"])
		if err != nil {
			return false
		}
		valJ, err := strconv.Atoi((*otherUsers[j])["score"])
		if err != nil {
			return false
		}
		return valI > valJ
	})
	// 3:1 ratio
	pfExpCount, otherExpCount := genderBasedCount(pageSize)

	pfResults, pfCount := Split(pfUsers, page, pfExpCount)
	otherResults, otherCount := Split(otherUsers, page, otherExpCount)

	combinedResult := make([]*map[string]string, pfCount+otherCount)
	copy(combinedResult, pfResults)
	copy(combinedResult[pfCount:], otherResults)

	// create image urls
	for _, user := range combinedResult {
		imgUrl, err := sgr.conentClient.CreateImageUrl((*user)["image_url"], contapi.Viewer)
		if err != nil {
			log.Println("failed at friend suggestions: failed to create avatar url: ", err)
			continue
		}
		(*user)["image_url"] = imgUrl
	}
	return combinedResult, nil
}

// default to male
func preferGender(gender string) string {
	if gender == male {
		return female
	}
	return male
}

// birthday format: yy-mm-dd
func ageMatch(birthdayUser, birthdayRef, genderUser, genderRef string) int {
	bdUser, err := time.Parse(birthdayLayout, birthdayUser)
	if err != nil {
		log.Println("Error parsing birthdayUser:", err)
		return 0
	}

	bdRef, err := time.Parse(birthdayLayout, birthdayRef)
	if err != nil {
		log.Println("Error parsing birthdayRef:", err)
		return 0
	}

	var diff float64
	if genderRef == female {
		diff = bdRef.Sub(bdUser).Hours() / perMonth

	} else if genderRef == male {
		if genderUser == female {
			diff = bdUser.Sub(bdRef).Hours() / perMonth
		} else {
			diff = bdRef.Sub(bdUser).Hours() / perMonth
		}

	} else {
		return 0
	}
	if diff < -120 {
		return 0
	}
	return int(diff + 120) // return y = x + (12months*10years)
}

func genderBasedCount(size int) (prefer int, other int) {
	// Ratio 3:1
	prefer = (3 * size) / 4
	other = size - prefer
	return
}

func Split(arr []*map[string]string, offset, count int) ([]*map[string]string, int) {
	ln := len(arr)
	end := offset + count
	if offset >= ln {
		return []*map[string]string{}, 0
	}
	if len(arr) < end {
		return arr[offset:], ln - offset
	}
	return arr[offset:end], count
}

func (sgr *socialGraph) CreateImagePost(ctx context.Context, userKey string, data *tp.Post) (string, string, error) {
	mediaKey, err := sgr.mediaRepo.CreateForGivenKey(ctx, &mrepo.Media{
		Link: data.Link, Title: data.Title, Description: data.Description, Visibility: data.Visibility,
		Key: "", CreateDate: "", Size: 0,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to create media node: %v", err)
	}
	mediaOwnerKey, err := sgr.mdOwnerCtrler.Create(ctx, sgr.mediaRepo.MkMediaDocId(mediaKey), sgr.userRepo.MkUserDocId(userKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to create media owner node: %v", err)
	}
	return mediaKey, mediaOwnerKey, nil
}
