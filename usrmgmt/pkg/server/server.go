package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	vdtor "github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	"github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
	jwtPrvdr "github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
	urepo "github.com/NamalSanjaya/nexster/pkgs/models/user"
	socigr "github.com/NamalSanjaya/nexster/usrmgmt/pkg/social_graph"
)

const (
	failed          string = "failed"
	success         string = "success"
	defaultPageNo   int    = 1
	defaultPageSize int    = 20
)

// Auth Provider Related Configs
const authProvider string = "usrmgmt"
const timeline string = "timeline"

type server struct {
	scGraph socigr.Interface
	logger  *lg.Logger
}

var _ Interface = (*server)(nil)

func New(sgrInterface socigr.Interface, logger *lg.Logger) *server {
	return &server{
		scGraph: sgrInterface,
		logger:  logger,
	}
}

func (s *server) HandleFriendReq(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	results := map[string]string{}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Errorf("failed create friend request: unsupported user_key type in JWT token: unauthorized request: user_key=%v", r.Context().Value(jwt.JwtUserKey))
		s.sendRespMsg(w, http.StatusUnauthorized, map[string]string{Date: ""}, map[string]interface{}{
			"state":   failed,
			"message": "unauthorized request",
			"data":    results,
		})
		return
	}

	data, err := s.readFriendReqJson(r)
	if err != nil {
		s.logger.Errorf("failed to read json content in friend req, Error: %v", err)
		s.sendRespMsg(w, http.StatusBadRequest, map[string]string{Date: ""}, map[string]interface{}{
			"state":   failed,
			"message": "request body is in wrong format",
			"data":    results,
		})
		return
	}
	if err = vdtor.New().Struct(data); err != nil {
		s.logger.Errorf("required fields are not in friend req json content, Error: %v", err)
		s.sendRespMsg(w, http.StatusBadRequest, map[string]string{Date: ""}, map[string]interface{}{
			"state":   failed,
			"message": "required fields are missing in request body",
			"data":    results,
		})
		return
	}

	if s.scGraph.GetRole(jwtUserKey, data.From) != urepo.Owner {
		s.logger.Error("failed create friend request: unauthorized request")
		s.sendRespMsg(w, http.StatusUnauthorized, map[string]string{Date: ""}, map[string]interface{}{
			"state":   failed,
			"message": "unauthorized request",
			"data":    results,
		})
		return
	}

	results, err = s.scGraph.CreateFriendReq(r.Context(), data.From, data.To, data.Mode, data.State, data.ReqDate)
	if err != nil {
		s.logger.Errorf("failed to create friend req edge in db, Error: %v", err)
		s.sendRespMsg(w, http.StatusInternalServerError, map[string]string{Date: ""}, map[string]interface{}{
			"state":   failed,
			"message": "failed to create required resources",
			"data":    results,
		})
		return
	}
	s.sendRespMsg(w, http.StatusOK, map[string]string{Date: ""}, map[string]interface{}{
		"state": success,
		"data":  results,
	})
}

func (s *server) RemovePendingFriendReq(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	headers := map[string]string{
		Date: "",
	}
	respBody := map[string]interface{}{
		"state": failed,
		"data":  map[string]string{},
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed remove friend request: unsupported user_key type in JWT token: unauthorized request")
		respBody["message"] = "unauthorized resource access"
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}
	var requestorId string
	if requestorId = r.URL.Query().Get("reqstor_id"); requestorId == "" {
		s.logger.Info("failed to remove friend request: reqstor_id query parameter is empty")
		respBody["message"] = "missing query parameter"
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}
	if s.scGraph.GetRole(jwtUserKey, requestorId) != urepo.Owner {
		s.logger.Info("failed remove friend request: unauthorized request")
		respBody["message"] = "unauthorized resource access"
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}
	friendReqId := p.ByName("friend_req_id")
	if friendReqId == "" {
		s.logger.Info("unable to remove friend request edge since friend_request_id is empty")
		respBody["message"] = "missing path parameter"
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}
	err := s.scGraph.RemoveFriendRequest(r.Context(), friendReqId)
	if err != nil {
		s.logger.Errorf("failed to remove friend request edge due to %v", err)
		respBody["message"] = "failed to remove friend request"
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}
	respBody["state"] = success
	respBody["message"] = "successfully to remove the friend request"
	respBody["data"] = map[string]string{"friend_req_id": friendReqId}
	s.sendRespMsg(w, http.StatusOK, headers, respBody)
}

// Create a friendship upon an accept of a friend request
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
	// user2Key is the acceptor
	if s.scGraph.GetRole(jwtUserKey, data.User2Key) != urepo.Owner {
		s.logger.Info("failed create friendship: unauthorized request")
		respBody["message"] = "unauthorized resource access"
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}
	results, err := s.scGraph.CreateFriend(r.Context(), friendReqId, data.User1Key, data.User2Key, data.AcceptedAt)
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
		s.logger.Info("failed remove friend edge: unsupported user_key type in JWT token: unauthorized request")
		respBody["message"] = "unauthorized resource access"
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}

	userId := r.URL.Query().Get("user_id")

	if s.scGraph.GetRole(jwtUserKey, userId) != urepo.Owner {
		s.logger.Info("failed remove friend edge: unauthorized request")
		respBody["message"] = "unauthorized resource access"
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}

	friendId := p.ByName("friend_id")
	if friendId == "" {
		s.logger.Info("failed to remove friend edge: friend_id is empty")
		respBody["message"] = "friend_id is empty"
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}

	toKey := r.URL.Query().Get("to_friend_id")
	if toKey == "" {
		s.logger.Info("failed to remove friend edge: To user friend_id is empty")
		respBody["message"] = "friend_id in query parameter is empty"
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}

	if err := s.scGraph.RemoveFriend(r.Context(), friendId, toKey); err != nil {
		s.logger.Errorf("failed to remove friend edge of %s due to %v", friendId, err)
		respBody["message"] = "server failed to remove resource"
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}

	respBody["state"] = success
	respBody["message"] = "successfully resource is removed"
	// TODO: Need to figure which data will send to front end
	s.sendRespMsg(w, http.StatusOK, headers, respBody)
}

func (s *server) ListFriendInfo(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	headers := map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed list friends: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespMsg(w, http.StatusUnauthorized, headers, map[string]interface{}{})
		return
	}
	userId := p.ByName("user_id")
	if s.scGraph.GetRole(jwtUserKey, userId) != urepo.Owner {
		s.logger.Info("failed list friends: unauthorized request")
		s.sendRespMsg(w, http.StatusUnauthorized, headers, map[string]interface{}{})
		return
	}
	pageNo, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		// s.logger.Warnf("page number not present in URL, therefore default page no = %d will be used", defaultPageNo)
		pageNo = defaultPageNo
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil {
		// s.logger.Warnf("page size not present in URL, therefore default page size = %d will be used", defaultPageSize)
		pageSize = defaultPageSize
	}

	respBody := map[string]interface{}{
		"state":         failed,
		"page":          pageNo,
		"page_size":     pageSize,
		"results_count": 0,
		"total_count":   0,
		"data":          map[string]string{},
	}
	results, err := s.scGraph.ListFriends(r.Context(), userId, (pageNo-1)*pageSize, pageSize)
	if err != nil {
		s.logger.Errorf("failed to list friends info due to %v", err)
		respBody["message"] = "server failed to list friends info"
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}
	totalCount, err := s.scGraph.CountFriends(r.Context(), userId)
	if err != nil {
		s.logger.Errorf("failed to count the friends. Err: %v", err)
		respBody["message"] = "server failed to list friends info"
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}
	resultsCount := len(results)
	respBody["state"] = success
	respBody["data"] = results
	respBody["results_count"] = resultsCount
	respBody["total_count"] = totalCount

	s.sendRespMsg(w, http.StatusOK, headers, respBody)
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

// TODO: This endpoint handler should be removed when the login logic handler implemented.
func (s *server) SetAuthToken(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := p.ByName("user_id")
	if userId == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed: userid is missing"))
		s.logger.Error("falied to Set-cookie: user_id is missing")
		return
	}
	aud := []string{authProvider, timeline}
	token, err := jwtPrvdr.GenJwtToken(authProvider, userId, aud)

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

// TODO: This endpoint handler should be removed when the login logic handler implemented.
func (s *server) SetCookie(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	subject := "482191" // TODO: change to user_key of authenticated user.
	aud := []string{authProvider, timeline}
	token, err := jwtPrvdr.GenJwtToken(authProvider, subject, aud)

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
