package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	vdtor "github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	"github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
	"github.com/NamalSanjaya/nexster/pkgs/crypto/hmac"
	"github.com/NamalSanjaya/nexster/pkgs/errors"
	umail "github.com/NamalSanjaya/nexster/pkgs/utill/mail"
	ustr "github.com/NamalSanjaya/nexster/pkgs/utill/string"
	tm "github.com/NamalSanjaya/nexster/pkgs/utill/time"
	socigr "github.com/NamalSanjaya/nexster/usrmgmt/pkg/social_graph"
	typ "github.com/NamalSanjaya/nexster/usrmgmt/pkg/types"
)

const (
	failed          string = "failed"
	success         string = "success"
	defaultPageNo   int    = 1
	defaultPageSize int    = 20
)

// email related
const subjectOfMail = "Welcome to Nexster - Create Your Account Now!"
const htmlMailBody string = `
<p>Hi,</p>
<p>Welcome to Nexster! We're excited to have you on board. To get started, please click the link below to create your account:</p>
<p><a href="%s">Link to Account Creation</a></p>
<p>Best regards,<br>The Nexster Team</p>
`

// Auth Provider Related Configs
const authProvider string = "usrmgmt"
const timeline string = "timeline"
const spaceAsAud string = "space"
const imageAsAud string = "image"
const searchAsAud string = "search"

type server struct {
	config     *ServerConfig
	scGraph    socigr.Interface
	logger     *lg.Logger
	mailClient umail.Interface
}

var _ Interface = (*server)(nil)

func New(cfg *ServerConfig, sgrInterface socigr.Interface, logger *lg.Logger, mailIntfce umail.Interface) *server {
	return &server{
		config:     cfg,
		scGraph:    sgrInterface,
		logger:     logger,
		mailClient: mailIntfce,
	}
}

func (s *server) ListFriendReqs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	headers := map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}
	respBody := map[string]interface{}{
		"state":         failed,
		"data":          []map[string]string{},
		"results_count": 0,
	}

	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Infof("failed list friend requests: unsupported user_key type in JWT token: unauthorized request: user_key=%v", r.Context().Value(jwt.JwtUserKey))
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}
	pageNo, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		pageNo = defaultPageNo
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil {
		pageSize = defaultPageSize
	}
	respBody["page"] = pageNo
	respBody["page_size"] = pageSize

	results, err := s.scGraph.ListFriendReqs(r.Context(), jwtUserKey, (pageNo-1)*pageSize, pageSize)
	if err != nil {
		s.logger.Errorf("failed to list friend requests: %v: userKey=%s", err, jwtUserKey)
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}
	respBody["state"] = success
	respBody["data"] = results
	respBody["results_count"] = len(results)
	s.sendRespMsg(w, http.StatusOK, headers, respBody)
}

func (s *server) GetAllFriendReqsCount(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{
		"state": failed,
		"data":  map[string]int{},
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Infof("failed to count friend req: unsupported user_key type in JWT token: unauthorized request: user_key=%v", r.Context().Value(jwt.JwtUserKey))
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}
	count, err := s.scGraph.GetAllFriendReqsCount(r.Context(), jwtUserKey)
	if err != nil {
		s.logger.Errorf("failed to count all friend requests: %v: userKey=%s", err, jwtUserKey)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	respBody["state"] = success
	respBody["data"] = map[string]int{"count": count}
	s.sendRespDefault(w, http.StatusOK, respBody)
}

func (s *server) CreateNewFriendReq(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{
		"state": failed,
		"data":  map[string]int{},
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Infof("failed create friend request: unsupported user_key type in JWT token: unauthorized request: user_key=%v", r.Context().Value(jwt.JwtUserKey))
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}

	data, err := s.readFriendReqJson(r)
	if err != nil {
		s.logger.Infof("failed to read json content in friend req, Error: %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}
	if err = vdtor.New().Struct(data); err != nil {
		s.logger.Infof("required fields are not in friend req json content, Error: %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}
	results, err := s.scGraph.CreateFriendReq(r.Context(), jwtUserKey, data.To, data.Mode, data.State, currentUTCTime())
	if errors.IsNotEligibleError(err) {
		s.logger.Infof("failed to create friend req: %v", err)
		s.sendRespDefault(w, http.StatusConflict, respBody)
		return
	}
	if err != nil {
		s.logger.Errorf("failed to create friend req edge in db, Error: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	s.sendRespDefault(w, http.StatusCreated, map[string]interface{}{
		"state": success,
		"data":  results,
	})
}

func (s *server) RemovePendingFriendReq(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{
		"state": failed,
		"data":  map[string]string{},
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed remove friend request: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}
	var otherUserKey string
	if otherUserKey = r.URL.Query().Get("other_id"); otherUserKey == "" {
		s.logger.Info("failed to remove friend request: reqstor_id query parameter is empty")
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}
	friendReqId := p.ByName("friend_req_id")
	if friendReqId == "" {
		s.logger.Info("unable to remove friend request edge since friend_request_id is empty")
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}
	err := s.scGraph.RemoveFriendRequest(r.Context(), friendReqId, jwtUserKey, otherUserKey)
	if errors.IsUnAuthError(err) {
		s.logger.Infof("failed remove friend request: %v", err)
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}
	if errors.IsNotFoundError(err) {
		s.logger.Infof("failed remove friend request: %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}
	if err != nil {
		s.logger.Errorf("failed to remove friend request edge due to %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	respBody["state"] = success
	respBody["data"] = map[string]string{"friend_req_id": friendReqId}
	s.sendRespDefault(w, http.StatusOK, respBody)
}

// Create a friendship upon an accept of a friend request.
// TODO: This function can refactor.
// 1. Don't need to bring user2Key since it is in JWT token.
// 2. AcceptAt timestamp should be system generated.
func (s *server) CreateFriendLink(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	headers := map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}
	respBody := map[string]interface{}{
		"state": failed,
		"data":  map[string]string{},
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed create friendship: unsupported user_key type in JWT token: unauthorized request")
		respBody["message"] = "unauthorized resource access"
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}

	friendReqId := p.ByName("friend_req_id")
	if friendReqId == "" {
		s.logger.Info("unable to create friend edge since friend_request_id is empty")
		respBody["message"] = "friend_request_id is empty"
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}
	data, err := s.readFriendReqAccptJson(r)
	if err != nil {
		s.logger.Errorf("unable to create friend request edge since invalid request body due to %v", err)
		respBody["message"] = "invalid request body"
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}
	if err = vdtor.New().Struct(data); err != nil {
		s.logger.Errorf("unable to create friend request edge since some mandadary fields are missing in request body due to %v", err)
		respBody["message"] = "mandadory fields are missing"
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}
	results, err := s.scGraph.CreateFriend(r.Context(), friendReqId, data.User1Key, jwtUserKey, currentUTCTime())
	if err != nil {
		s.logger.Errorf("unable to create friend request edge since server failed to create required resources due to %v", err)
		respBody["message"] = "server failed to create friend link"
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}
	s.sendRespMsg(w, http.StatusCreated, headers, map[string]interface{}{
		"state":   success,
		"message": "friendship created",
		"data":    results,
	})
}

// TODO: Think whether we need friend_edge ids or id of two users.
func (s *server) RemoveFriendship(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{
		"state": failed,
		"data":  map[string]string{},
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed remove friend edge: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}
	friendId := p.ByName("friend_id")
	if friendId == "" {
		s.logger.Info("failed to remove friend edge: friend_id is empty")
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}
	var result map[string]string
	var err error
	if result, err = s.scGraph.RemoveFriendV2(r.Context(), jwtUserKey, friendId); err != nil {
		s.logger.Errorf("failed to remove friend edge of %s: %v", friendId, err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	respBody["state"] = success
	respBody["data"] = result
	s.sendRespDefault(w, http.StatusOK, respBody)
}

func (s *server) ListFriendInfo(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{
		"state": failed,
		"data":  map[string]string{},
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed list friends: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}
	pageNo, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		pageNo = defaultPageNo
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil {
		pageSize = defaultPageSize
	}
	results, err := s.scGraph.ListFriends(r.Context(), jwtUserKey, (pageNo-1)*pageSize, pageSize)
	if err != nil {
		s.logger.Errorf("failed to list friends info due to %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	totalCount, err := s.scGraph.CountFriends(r.Context(), jwtUserKey)
	if err != nil {
		s.logger.Errorf("failed to count the friends. Err: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	respBody = map[string]interface{}{
		"state":         success,
		"page":          pageNo,
		"page_size":     pageSize,
		"results_count": len(results),
		"total_count":   totalCount,
		"data":          results,
	}
	s.sendRespDefault(w, http.StatusOK, respBody)
}

// permission : Both (owner, viewer)
func (s *server) GetProfile(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userKey := p.ByName("user_id")

	info, err := s.scGraph.GetProfileInfo(r.Context(), userKey)
	if err != nil {
		s.logger.Errorf("failed to get profile info: userKey: %s: %v", userKey, err)
		s.sendRespMsg(w, http.StatusInternalServerError, map[string]string{
			ContentType: ApplicationJson_Utf8,
			Date:        "",
		}, map[string]interface{}{
			"state": failed,
			"data":  map[string]string{},
		})
		return
	}
	s.sendRespMsg(w, http.StatusOK, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}, map[string]interface{}{
		"state": success,
		"data":  info,
	})
}

func (s *server) GetFriendsCount(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userKey := p.ByName("user_id")
	count, err := s.scGraph.CountFriendsV2(r.Context(), userKey)
	if err != nil {
		s.logger.Errorf("failed to get friend count: userKey: %s: %v", userKey, err)
		s.sendRespMsg(w, http.StatusInternalServerError, map[string]string{
			ContentType: ApplicationJson_Utf8,
			Date:        "",
		}, map[string]interface{}{
			"state": failed,
			"data":  map[string]string{},
		})
		return
	}
	s.sendRespMsg(w, http.StatusOK, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}, map[string]interface{}{
		"state": success,
		"data":  map[string]int{"count": count},
	})
}

// owner permission
func (s *server) GetUserKeyByIndexNo(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	indexNo := p.ByName("index_no")
	headers := map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}
	respBody := map[string]interface{}{
		"state": failed,
		"data":  map[string]string{"key": ""},
	}
	userKey, err := s.scGraph.GetUserKeyByIndexNo(r.Context(), indexNo)
	if err != nil {
		s.logger.Errorf("failed to get userKey for given indexNo=%s: %v", indexNo, err)
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}
	respBody["state"] = success
	respBody["data"] = map[string]string{"key": userKey}
	s.sendRespMsg(w, http.StatusOK, headers, respBody)
}

// TODO: This endpoint handler should be removed when the login logic handler implemented.
func (s *server) SetAuthToken(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := p.ByName("user_id")
	if userId == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed: userid is missing"))
		s.logger.Error("falied to Set-cookie: user_id is missing")
		return
	}
	aud := []string{authProvider, timeline, spaceAsAud, imageAsAud, searchAsAud}
	token, err := jwt.GenJwtToken(authProvider, userId, aud)

	if err != nil {
		// log the error
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed: server failed to generate a token"))
		s.logger.Errorf("falied to Set-cookie: %v", err)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authorization header set in response"))
}

func (s *server) EditBasicProfileInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{"state": failed}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed update basic profile info: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}

	body, err := typ.ReadJsonBody[typ.Profile](r)
	if err != nil {
		s.logger.Infof("failed update basic profile info: failed to read request body: %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}
	data := typ.RemoveEmptyFields[typ.Profile](body)

	err = s.scGraph.UpdateUser(r.Context(), jwtUserKey, data)

	if errors.IsNotFoundError(err) {
		s.logger.Infof("failed update basic profile info: failed to find document: %v", err)
		s.sendRespDefault(w, http.StatusNotFound, respBody)
		return
	}
	if err != nil {
		s.logger.Errorf("failed update basic profile info: failed to update user: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	s.sendRespDefault(w, http.StatusOK, map[string]interface{}{"state": success})
}

func (s *server) DeleteUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{"state": failed}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed delete user profile: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}

	err := s.scGraph.DeleteUser(r.Context(), jwtUserKey)

	if errors.IsNotFoundError(err) {
		s.logger.Infof("failed delete user profile: failed to find document: %v", err)
		s.sendRespDefault(w, http.StatusNotFound, respBody)
		return
	}
	if err != nil {
		s.logger.Errorf("failed delete user profile: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	s.sendRespDefault(w, http.StatusOK, map[string]interface{}{"state": success})
}

func (s *server) ResetPassword(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{"state": failed}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed reset password: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}

	data, err := typ.ReadJsonBody[typ.PasswordResetInfo](r)
	if err != nil {
		s.logger.Infof("failed reset password: failed to read request body: %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}
	if err = vdtor.New().Struct(data); err != nil {
		s.logger.Infof("failed reset password: required fields are not in password reset json content, %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}

	err = s.scGraph.ResetPassword(r.Context(), jwtUserKey, data.CurrentPassword, data.NewPassword)

	if errors.IsNotEligibleError(err) {
		s.logger.Infof("failed reset password: %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}

	if errors.IsNotFoundError(err) {
		s.logger.Infof("failed reset password: failed to find user: %v", err)
		s.sendRespDefault(w, http.StatusNotFound, respBody)
		return
	}
	if errors.IsNotConflictError(err) {
		s.logger.Infof("failed reset password: many users exist with usere key=%s: %v", jwtUserKey, err)
		s.sendRespDefault(w, http.StatusConflict, respBody)
		return
	}
	if err != nil {
		s.logger.Errorf("failed reset password: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	s.sendRespDefault(w, http.StatusOK, map[string]interface{}{"state": success})
}

func (s *server) GetAccessToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{
		"state": failed,
		"data":  map[string]string{},
	}

	data, err := typ.ReadJsonBody[typ.AccessTokenBody](r)
	if err != nil {
		s.logger.Infof("failed get access token: failed to read request body: %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}
	if err = vdtor.New().Struct(data); err != nil {
		s.logger.Infof("failed get access token: required fields are not in password reset json content, %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}

	// if validate success, this will return relavent user key.
	userKey, err := s.scGraph.ValidatePasswordForToken(r.Context(), data.IndexNo, data.Password)

	if errors.IsUnAuthError(err) {
		s.logger.Infof("failed get access token: %v", err)
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}
	if errors.IsNotFoundError(err) {
		s.logger.Infof("failed get access token: %v", err)
		s.sendRespDefault(w, http.StatusNotFound, respBody)
		return
	}
	if errors.IsNotConflictError(err) {
		s.logger.Infof("failed get access token: many users exist with index no=%s: %v", data.IndexNo, err)
		s.sendRespDefault(w, http.StatusConflict, respBody)
		return
	}
	if err != nil {
		s.logger.Errorf("failed get access token: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}

	aud := []string{authProvider, timeline, spaceAsAud, imageAsAud, searchAsAud}
	accessToken, err := jwt.GenJwtToken(authProvider, userKey, aud)

	if err != nil {
		s.logger.Errorf("failed get access token: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}

	s.sendRespDefault(w, http.StatusOK, map[string]interface{}{
		"state": success,
		"data": map[string]string{
			"access_token": accessToken,
			"id":           userKey,
		},
	})
}

func (s *server) EmailAccountCreationLink(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{"state": failed}

	data, err := typ.ReadJsonBody[typ.AccountCreationLinkBody](r)
	if err != nil {
		s.logger.Infof("failed to send account creation link: failed to read request body: %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}
	if err = vdtor.New().Struct(data); err != nil {
		s.logger.Infof("failed to send account creation link: required fields are not in password reset json content, %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}

	// TODO:
	// Allow only for predetermined index no: (eg: by looking at first three characters in index number.)

	userKey, err := s.scGraph.GetUserKeyByIndexNo(r.Context(), data.IndexNo)
	if err != nil {
		s.logger.Errorf("failed to send account creation link: unable to get user key: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	if userKey != "" {
		s.logger.Info("failed to send account creation link: account is already exist.")
		s.sendRespDefault(w, http.StatusConflict, respBody)
		return
	}

	// 30 min account creation link
	expiredAt := strconv.FormatInt(tm.AddMinToCurrentTime(30), 10)

	accountCreationLink := fmt.Sprintf("%s/%s?index=%s&exp=%s&hmac=%s", s.config.FrontendDomain, s.config.FrontendPath,
		data.IndexNo, expiredAt, hmac.CalculateHMAC(s.config.SecretHmacKey, data.IndexNo, expiredAt),
	)

	if err = s.mailClient.SendEmail(fmt.Sprintf("%s@uom.lk", data.IndexNo), subjectOfMail, fmt.Sprintf(htmlMailBody, accountCreationLink)); err != nil {
		s.logger.Errorf("failed to send account creation link: unable to send the mail: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}

	s.sendRespDefault(w, http.StatusOK, map[string]interface{}{"state": success})

}

func (s *server) ValidateLinkCreationParams(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{"state": failed}

	data, err := typ.ReadJsonBody[typ.LinkCreationParams](r)
	if err != nil {
		s.logger.Infof("failed to verify link creation params: failed to read request body: %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}
	if err = vdtor.New().Struct(data); err != nil {
		s.logger.Infof("failed to verify link creation params: required fields are not in password reset json content, %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}

	// hmac validation
	if !hmac.ValidateHMAC(s.config.SecretHmacKey, data.Hmac, data.IndexNo, data.ExpiredAt) {
		s.logger.Info("failed to verify link creation params: hmac valdiation failed")
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}

	expiredAt, err := ustr.StrToInt64(data.ExpiredAt)
	if err != nil {
		s.logger.Infof("failed to verify link creation params: %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}

	if tm.HasUnixTimeExceeded(expiredAt) {
		s.logger.Infof("failed to verify link creation params: link is expired")
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}

	userKey, err := s.scGraph.GetUserKeyByIndexNo(r.Context(), data.IndexNo)
	if err != nil {
		s.logger.Errorf("failed to verify link creation params: unable to get user key: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	if userKey != "" {
		s.logger.Info("failed to verify link creation params: account is already exist.")
		s.sendRespDefault(w, http.StatusConflict, respBody)
		return
	}
	s.sendRespDefault(w, http.StatusOK, map[string]interface{}{"state": success})
}

// TODO: This endpoint handler should be removed when the login logic handler implemented.
func (s *server) SetCookie(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	subject := "482191" // TODO: change to user_key of authenticated user.
	aud := []string{authProvider, timeline}
	token, err := jwt.GenJwtToken(authProvider, subject, aud)

	if err != nil {
		// log the error
		s.logger.Errorf("falied to Set-cookie: %v", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: token,
		// Secure:   true, // TODO: Enable Secure: true, once you have the https connection.
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/", // Cookie is valid for all paths
		MaxAge:   600, // Valid only for 10min (only in development)
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("New jwt token enabled...!"))
}

func (s *server) readFriendReqJson(r *http.Request) (*FriendRequest, error) {
	data := &FriendRequest{}
	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return data, err
	}
	if err = json.Unmarshal(b, &data); err != nil {
		return data, err
	}
	return data, nil
}

func (s *server) readFriendReqAccptJson(r *http.Request) (*FriendReqAcceptance, error) {
	data := &FriendReqAcceptance{}
	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return data, err
	}
	if err = json.Unmarshal(b, &data); err != nil {
		return data, err
	}
	return data, nil
}

func (s *server) sendRespMsg(w http.ResponseWriter, statusCode int, headers map[string]string, body map[string]interface{}) {
	for key, val := range headers {
		w.Header().Add(key, val)
	}
	w.WriteHeader(statusCode)
	resp, _ := json.Marshal(body)
	w.Write(resp)
}

// similar to `sendRespMsg` but only have predefined headers
func (s *server) sendRespDefault(w http.ResponseWriter, statusCode int, body map[string]interface{}) {
	w.Header().Add(ContentType, ApplicationJson_Utf8)
	w.Header().Add(Date, "")
	w.WriteHeader(statusCode)
	resp, _ := json.Marshal(body)
	w.Write(resp)
}

func currentUTCTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}
