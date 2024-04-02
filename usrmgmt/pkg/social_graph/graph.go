package socialgraph

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"

	contapi "github.com/NamalSanjaya/nexster/pkgs/client/content_api"
	errs "github.com/NamalSanjaya/nexster/pkgs/errors"
	avtr "github.com/NamalSanjaya/nexster/pkgs/models/avatar"
	bdo "github.com/NamalSanjaya/nexster/pkgs/models/boardingOwner"
	fac "github.com/NamalSanjaya/nexster/pkgs/models/faculty"
	intrsIn "github.com/NamalSanjaya/nexster/pkgs/models/interestsIn"

	frnd "github.com/NamalSanjaya/nexster/pkgs/models/friend"
	freq "github.com/NamalSanjaya/nexster/pkgs/models/friend_request"
	gnd "github.com/NamalSanjaya/nexster/pkgs/models/genders"
	hgen "github.com/NamalSanjaya/nexster/pkgs/models/hasGender"
	stdt "github.com/NamalSanjaya/nexster/pkgs/models/student"
	usr "github.com/NamalSanjaya/nexster/pkgs/models/user"
	pwd "github.com/NamalSanjaya/nexster/pkgs/utill/password"
	strg "github.com/NamalSanjaya/nexster/pkgs/utill/string"
	typ "github.com/NamalSanjaya/nexster/usrmgmt/pkg/types"
)

const userColl string = "users" // Need to be changed once `users` repo bring to common level

const gettFriendReqEdgeQuery string = `FOR v,e IN 1..1 ANY
	@reqstorNode friendRequest
	OPTIONS { uniqueVertices: "path" }
	FILTER e.kind == "friend_request" && v._id == @friendNode
	return e._key`

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

const getUserKeyByIndexEmailQry string = `FOR user IN users
	FILTER user.index_no == @indexNo || user.email == @email
	RETURN user._key`

const getUserKeyByEmailQry string = `FOR user IN users
	FILTER user.email == @email
	RETURN user._key`

const listFriendReqs string = `FOR v,e IN 1..1 INBOUND
	@userNode friendRequest
	SORT e.req_date DESC
	LIMIT @offset, @count
	RETURN { "user_key": v._key, "username" : v.username, "image_url" : v.image_url, 
	"batch": v.batch,"faculty": v.faculty, "field" : v.field, "indexNo": v.index_no,
	"req_date": e.req_date, "req_key": e._key }`

const allFriendReqsCountQry string = `FOR doc IN friendRequest
	FILTER doc._to == @userNode
	COLLECT WITH COUNT INTO len
	RETURN len`

const friendReqPairQry string = `FOR doc IN friendRequest
	FILTER doc._key == @friendReqKey
	RETURN {"from" : doc._from, "to" : doc._to }`

const getPasswordQry string = `FOR v IN users
	filter v._key == @givenUserKey
	RETURN v.password`

const getLoginInfoForIndxQry string = `FOR v IN users
	filter v.index_no == @givenIndexNo
	RETURN {"key": v._key, "password":  v.password, "roles": v.roles  }`

const getBdOwnersForLogin string = `FOR v IN boardingOwners
	FILTER v.status == "active"
	FILTER v.mainContact == @mainContact
	RETURN {"key": v._key, "password":  v.password, "roles": v.roles  }`

type socialGraph struct {
	fReqCtrler        freq.Interface
	frndCtrler        frnd.Interface
	usrCtrler         usr.Interface
	conentClient      contapi.Interface
	avatarCtrler      avtr.Interface
	studentCtrler     stdt.Interface
	facultyCtrler     fac.Interface
	hasGenderCtrler   hgen.Interface
	bdOwnerCtrler     bdo.Interface
	interestsInCtrler intrsIn.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewGrphCtrler(frIntfce freq.Interface, frndIntfce frnd.Interface, usrIntfce usr.Interface, contentIntfce contapi.Interface, avtrIntfce avtr.Interface,
	stIntfce stdt.Interface, facIntface fac.Interface, hGenIntface hgen.Interface, bdOwnerIntfce bdo.Interface, interestsInIntfce intrsIn.Interface) *socialGraph {
	return &socialGraph{
		fReqCtrler:        frIntfce,
		frndCtrler:        frndIntfce,
		usrCtrler:         usrIntfce,
		conentClient:      contentIntfce,
		avatarCtrler:      avtrIntfce,
		studentCtrler:     stIntfce,
		facultyCtrler:     facIntface,
		hasGenderCtrler:   hGenIntface,
		bdOwnerCtrler:     bdOwnerIntfce,
		interestsInCtrler: interestsInIntfce,
	}
}

func (sgr *socialGraph) ListFriendReqs(ctx context.Context, userKey string, offset, count int) ([]*map[string]string, error) {
	friendReqs, err := sgr.fReqCtrler.ListStringValueJson(ctx, listFriendReqs, map[string]interface{}{
		"userNode": sgr.usrCtrler.MkUserDocId(userKey),
		"offset":   offset,
		"count":    count,
	})
	if err != nil {
		return []*map[string]string{}, err
	}
	for _, friendReq := range friendReqs {
		imgUrl, err := sgr.conentClient.CreateImageUrl((*friendReq)["image_url"], contapi.Viewer)
		if err != nil {
			log.Println("error when listing friend reqs: failed to create image url", err)
			continue
		}
		(*friendReq)["image_url"] = imgUrl
	}
	return friendReqs, nil
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
// 3. if from == to then reject creating req link
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
		return results, errs.NewNotEligibleError(fmt.Sprintf("friend req already exist for reqstorKey=%s, friendKey=%s", reqstorKey, friendKey))
	}

	if isExist, err = sgr.frndCtrler.IsFriendEdgeExist(ctx, reqstorId, friendId); err != nil {
		return results, fmt.Errorf("failed to check the existance of friend edge [from %s, to %s]. Error: %v", reqstorId, friendId, err)
	}
	if isExist {
		return results, errs.NewNotEligibleError(fmt.Sprintf("friendship already exist for reqstorKey=%s, friendKey=%s", reqstorKey, friendKey))
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

func (sgr *socialGraph) RemoveFriendRequest(ctx context.Context, friendkey, user1Key, user2Key string) error {
	// 1. Get from, to for that friendKey. check with user1Key and user2Key.
	// if from, to differ don't delete it. return unAuthorized actions.(new custom error)
	pair, err := sgr.fReqCtrler.ListStringValueJson(ctx, friendReqPairQry, map[string]interface{}{
		"friendReqKey": friendkey,
	})
	if err != nil {
		return fmt.Errorf("failed to check two ends of edge: friendId=%s: %v", friendkey, err)
	}

	ln := len(pair)
	if ln == 0 {
		// TODO: Return NotFoundError
		return errs.NewNotFoundError(fmt.Sprintf("no friend req doc found for friendKey=%s", friendkey))
	}
	if ln > 1 {
		return fmt.Errorf("found more than one doc for given key=%s", friendkey)
	}
	fromId := (*pair[0])["from"]
	toId := (*pair[0])["to"]
	user1Id := sgr.usrCtrler.MkUserDocId(user1Key)
	user2Id := sgr.usrCtrler.MkUserDocId(user2Key)

	notAuth := true
	if fromId == user1Id {
		if toId == user2Id {
			notAuth = false
		}
	}
	if fromId == user2Id {
		if toId == user1Id {
			notAuth = false
		}
	}
	if notAuth {
		// TODO: Return UnAuthError
		return errs.NewUnAuthError(fmt.Sprintf("%s, %s users, don't belong to friendKey=%s", user1Key, user2Key, friendkey))
	}
	return sgr.fReqCtrler.RemoveFriendReqEdge(ctx, friendkey)
}

// ISSUES:
// 1. even if users are not exist it will create the friend link with non-existing node.
// 2. check is the given friend_req coming from given requestor_id. [HIGH]
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

func (sgr *socialGraph) RemoveFriendV2(ctx context.Context, userKey1, userKey2 string) (map[string]string, error) {
	return sgr.frndCtrler.RemoveFriendship(ctx, sgr.usrCtrler.MkUserDocId(userKey1), sgr.usrCtrler.MkUserDocId(userKey2))
}

func (sgr *socialGraph) ListFriends(ctx context.Context, userKey string, offset, count int) ([]*map[string]string, error) {
	friends, err := sgr.frndCtrler.ListFriends(ctx, sgr.usrCtrler.MkUserDocId(userKey), offset, count)
	if err != nil {
		return []*map[string]string{}, err
	}
	for _, friend := range friends {
		imgUrl, err := sgr.conentClient.CreateImageUrl((*friend)["image_url"], contapi.Viewer)
		if err != nil {
			log.Println("failed to create post url: ", err)
			continue
		}
		(*friend)["image_url"] = imgUrl
	}
	return friends, nil
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
	imgUrl, err := sgr.conentClient.CreateImageUrl(info.ImageUrl, contapi.Viewer)
	if err != nil {
		log.Println("failed to create image url when creating profile info: ", err)
	}
	return map[string]string{
		"key": userKey, "username": info.Username, "faculty": info.Faculty, "field": info.Field, "batch": info.Batch,
		"img_url": imgUrl, "about": info.About, "firstName": info.FirstName, "secondName": info.SecondName, "gender": info.Gender, "birthday": info.Birthday,
	}, nil
}

func (sgr *socialGraph) CountFriendsV2(ctx context.Context, userId string) (int, error) {
	return sgr.frndCtrler.CountFriends(ctx, totalFriendsV2, map[string]interface{}{
		"startNode": sgr.usrCtrler.MkUserDocId(userId),
	})
}

func (sgr *socialGraph) GetUserKeyByIndexNo(ctx context.Context, indexNo string) (string, error) {
	indexNo = strings.ToLower(indexNo)
	res, err := sgr.usrCtrler.ListStrings(ctx, getUserKey, map[string]interface{}{
		"indexNo": indexNo,
	})
	resLn := len(res)
	if resLn == 0 {
		return "", nil
	}
	if resLn > 1 {
		return "", fmt.Errorf("indexNo=%s is not unique, array of userkeys exists", indexNo)
	}
	return *res[0], err
}

// return error if not unique
func (sgr *socialGraph) ExistUserForIndexEmail(ctx context.Context, indexNo, email string) (bool, error) {
	indexNo = strings.ToLower(indexNo)
	res, err := sgr.usrCtrler.ListStrings(ctx, getUserKeyByIndexEmailQry, map[string]interface{}{
		"indexNo": indexNo,
		"email":   email,
	})
	if err != nil {
		return false, err
	}
	resLn := len(res)
	if resLn == 0 {
		return false, nil
	}
	if resLn > 1 {
		return true, errs.NewConflictError(fmt.Sprintf("multiple users exist for index=%s, email=%s", indexNo, email))
	}
	return true, nil
}

func (sgr *socialGraph) ExistUserForEmail(ctx context.Context, email string) (bool, error) {
	res, err := sgr.usrCtrler.ListStrings(ctx, getUserKeyByEmailQry, map[string]interface{}{
		"email": email,
	})
	if err != nil {
		return false, err
	}
	resLn := len(res)
	if resLn == 0 {
		return false, nil
	}
	if resLn > 1 {
		return true, errs.NewConflictError(fmt.Sprintf("multiple users exist for email=%s", email))
	}
	return true, nil
}

func (sgr *socialGraph) UpdateUser(ctx context.Context, userId string, data map[string]interface{}) error {
	return sgr.usrCtrler.UpdateUser(ctx, userId, data)
}

// TODO:
// Need to remove necessary edges when deleting a user node.
func (sgr *socialGraph) DeleteUser(ctx context.Context, userId string) error {
	return sgr.usrCtrler.DeleteUser(ctx, userId)
}

func (sgr *socialGraph) ResetPassword(ctx context.Context, userKey, givenOldPasswd, newPasswd string) error {
	results, err := sgr.usrCtrler.ListStrings(ctx, getPasswordQry, map[string]interface{}{
		"givenUserKey": userKey,
	})
	if err != nil {
		return err
	}

	ln := len(results)
	if ln == 0 {
		return errs.NewNotFoundError(fmt.Sprintf("user with key %s not found", userKey))
	}
	if ln > 1 {
		return errs.NewConflictError(fmt.Sprintf("more than one user found for user key=%s", userKey))
	}
	curPasswdHash := *results[0]
	if !pwd.CheckPasswordHash(givenOldPasswd, curPasswdHash) {
		return errs.NewNotEligibleError("password is not matched")
	}

	newPasswdHash, err := pwd.HashPassword(newPasswd)
	if err != nil {
		return fmt.Errorf("failed to hash the password: %v", err)
	}

	return sgr.usrCtrler.UpdateUser(ctx, userKey, map[string]interface{}{
		"password": newPasswdHash,
	})
}

func (sgr *socialGraph) ForgotPasswordReset(ctx context.Context, email, newPasswd string) error {
	results, err := sgr.usrCtrler.ListStrings(ctx, getUserKeyByEmailQry, map[string]interface{}{
		"email": email,
	})
	if err != nil {
		return err
	}

	ln := len(results)
	if ln == 0 {
		return errs.NewNotFoundError(fmt.Sprintf("user not found for email=%s", email))
	}
	if ln > 1 {
		return errs.NewConflictError(fmt.Sprintf("more than one user found for email=%s", email))
	}
	newPasswdHash, err := pwd.HashPassword(newPasswd)
	if err != nil {
		return fmt.Errorf("failed to hash the password: %v", err)
	}
	return sgr.usrCtrler.UpdateUser(ctx, *results[0], map[string]interface{}{
		"password": newPasswdHash,
	})
}

// if password is match userKey, if not UnAuth error will be returned.
func (sgr *socialGraph) ValidatePasswordForToken(ctx context.Context, id, givenPasswd, consumerType string) (userKey string, roles []string, err error) {
	roles = []string{}
	results := []*map[string]interface{}{}

	if consumerType == typ.Student {
		results, err = sgr.usrCtrler.ListUsersAnyJsonValue(ctx, getLoginInfoForIndxQry, map[string]interface{}{
			"givenIndexNo": id, // id --> index no for student consumer.
		})
	} else if consumerType == typ.BoardingOwner {
		results, err = sgr.bdOwnerCtrler.ListAnyJsonValue(ctx, getBdOwnersForLogin, map[string]interface{}{
			"mainContact": id, // id --> main contact phone no for boarding owner consumer.
		})
	} else {
		err = fmt.Errorf("invalid consumer type given, %s", consumerType)
		return
	}

	if err != nil {
		return
	}
	ln := len(results)
	if ln == 0 {
		err = errs.NewNotFoundError(fmt.Sprintf("user is not found: id=%s, consumerType=%s", id, consumerType))
		return
	}
	if ln > 1 {
		err = errs.NewConflictError(fmt.Sprintf("more than one user found: id=%s, consumerType=%s", id, consumerType))
		return
	}
	result := *results[0]

	curPasswdHash, err := strg.InterfaceToString(result["password"])
	if err != nil {
		err = fmt.Errorf("password is not found: %v", err)
		return
	}

	if !pwd.CheckPasswordHash(givenPasswd, curPasswdHash) {
		err = errs.NewUnAuthError("password is not matched")
		return
	}
	// get user key
	userKey, err = strg.InterfaceToString(result["key"])
	if err != nil {
		return
	}
	// get roles array
	roles, err = strg.InterfaceToStringArray(result["roles"])
	return
}

// TODO: Need to add roles (students)
func (sgr *socialGraph) CreateUserNode(ctx context.Context, data *typ.AccCreateBody, defaultRoles []string) (string, error) {
	newPasswdHash, err := pwd.HashPassword(data.Password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}
	// TODO: data validation(birthday, faculty, field, batch, gender)
	userKey, err := sgr.usrCtrler.CreateDocument(ctx, &usr.UserCreateInfo{
		Key:        "",
		FirstName:  data.FirstName,
		SecondName: data.SecondName,
		Username:   "",
		IndexNo:    data.IndexNo,
		Email:      data.Email,
		ImageUrl:   data.ImageId, // avatar/47193434.jpg
		Birthday:   data.Birthday,
		Faculty:    data.Faculty,
		Field:      data.Field,
		Batch:      data.Batch,
		About:      data.About,
		Gender:     data.Gender,
		Password:   newPasswdHash,
		Roles:      defaultRoles,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create user document: %v", err)
	}

	value, ext, err := getExtensionAndNumber(data.ImageId)
	if err != nil {
		return "", fmt.Errorf("err parsing imageId: %v", err)
	}
	// Create avatar node
	_, err = sgr.avatarCtrler.Create(ctx, &avtr.Avatar{
		Key:    value,
		Format: ext,
		View:   avtr.PublicView,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create avatar node: %v", err)
	}

	// Create student edge
	_, err = sgr.studentCtrler.Create(ctx, &stdt.Student{
		Key:  "",
		From: usr.MkUserDocId(userKey),
		To:   sgr.facultyCtrler.MkFacultyDocId(strings.ToLower(data.Faculty)),
		Kind: "",
	})
	if err != nil {
		return "", fmt.Errorf("failed to create student edge: %v", err)
	}

	// Create hasGender edge
	_, err = sgr.hasGenderCtrler.Create(ctx, &hgen.HasGender{
		Key:  "",
		From: usr.MkUserDocId(userKey),
		To:   gnd.MkGenderId(data.Gender), // TODO: Gender will go data transform
		Kind: "",
	})

	if err != nil {
		return "", fmt.Errorf("failed to create hasGender edge: %v", err)
	}

	// create interestsIn edge
	facDepName := data.Faculty
	if facDepName == "Engineering" || data.Field != "" {
		facDepName = data.Field
	}
	if err = sgr.interestsInCtrler.InsertByFacDepName(ctx, facDepName, userKey); err != nil {
		log.Printf("user creation: failed to create interestIn edge: %v\n", err)
	}

	return userKey, nil
}

func getExtensionAndNumber(input string) (valuePart, extension string, err error) {
	arr1 := strings.Split(input, ".")
	if len(arr1) != 2 {
		err = fmt.Errorf("invalid input")
		return
	}
	extension = arr1[1]
	arr2 := strings.Split(arr1[0], "/")

	if len(arr2) != 2 {
		err = fmt.Errorf("invalid input")
		extension = ""
		return
	}
	valuePart = arr2[1]
	return
}
