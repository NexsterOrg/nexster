package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	vdtor "github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	jwtPrvdr "github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
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
	ctx := context.Background()
	results, err = s.scGraph.CreateFriendReq(ctx, data.From, data.To, data.Mode, data.State, data.ReqDate)
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
	friendReqId := p.ByName("friend_req_id")
	if friendReqId == "" {
		s.logger.Error("unable to remove friend request edge since friend_request_id is empty")
		s.sendRespMsg(w, http.StatusBadRequest, map[string]string{Date: ""}, map[string]interface{}{
			"state":   failed,
			"message": "friend_request_id is empty",
		})
		return
	}
	err := s.scGraph.RemoveFriendRequest(context.Background(), friendReqId)
	if err != nil {
		s.logger.Errorf("failed to remove friend request edge due to %v", err)
		s.sendRespMsg(w, http.StatusInternalServerError, map[string]string{Date: ""}, map[string]interface{}{
			"state":   failed,
			"message": "failed to remove friend request",
		})
		return
	}

	s.sendRespMsg(w, http.StatusOK, map[string]string{Date: ""}, map[string]interface{}{
		"state": success,
	})
}

func (s *server) CreateFriendLink(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	friendReqId := p.ByName("friend_req_id")
	if friendReqId == "" {
		s.logger.Error("unable to create friend edge since friend_request_id is empty")
		s.sendRespMsg(w, http.StatusBadRequest, map[string]string{Date: ""}, map[string]interface{}{
			"state":   failed,
			"message": "friend_request_id is empty",
			"data":    map[string]string{},
		})
		return
	}
	data, err := s.readFriendReqAccptJson(r)
	if err != nil {
		s.logger.Errorf("unable to create friend request edge since invalid request body due to %v", err)
		s.sendRespMsg(w, http.StatusBadRequest, map[string]string{Date: ""}, map[string]interface{}{
			"state":   failed,
			"message": "invalid request body",
			"data":    map[string]string{},
		})
		return
	}
	if err = vdtor.New().Struct(data); err != nil {
		s.logger.Errorf("unable to create friend request edge since some mandadary fields are missing in request body due to %v", err)
		s.sendRespMsg(w, http.StatusBadRequest, map[string]string{Date: ""}, map[string]interface{}{
			"state":   failed,
			"message": "mandadory fields are missing",
			"data":    map[string]string{},
		})
		return
	}
	results, err := s.scGraph.CreateFriend(context.Background(), friendReqId, data.User1Key, data.User2Key, data.AcceptedAt)
	if err != nil {
		s.logger.Errorf("unable to create friend request edge since server failed to create required resources due to %v", err)
		s.sendRespMsg(w, http.StatusInternalServerError, map[string]string{Date: ""}, map[string]interface{}{
			"state":   failed,
			"message": "server failed to create friend link",
			"data":    map[string]string{},
		})
		return
	}
	s.sendRespMsg(w, http.StatusOK, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}, map[string]interface{}{
		"state": success,
		"data":  results,
	})
}

func (s *server) RemoveFriendship(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	friendId := p.ByName("friend_id")
	if friendId == "" {
		s.logger.Errorf("failed to remove friend edge of since friend_id is empty", friendId)
		s.sendRespMsg(w, http.StatusBadRequest, map[string]string{
			ContentType: ApplicationJson_Utf8,
			Date:        "",
		}, map[string]interface{}{
			"state":   failed,
			"message": "friend_id is empty",
		})
		return
	}

	toKey := r.URL.Query().Get("to_friend_id")
	if toKey == "" {
		s.logger.Errorf("failed to remove friend edge of since to user friend_id is empty", friendId)
		s.sendRespMsg(w, http.StatusBadRequest, map[string]string{
			ContentType: ApplicationJson_Utf8,
			Date:        "",
		}, map[string]interface{}{
			"state":   failed,
			"message": "friend_id in query parameter is empty",
		})
		return
	}

	if err := s.scGraph.RemoveFriend(context.Background(), friendId, toKey); err != nil {
		s.logger.Errorf("failed to remove friend edge of %s due to %v", friendId, err)
		s.sendRespMsg(w, http.StatusInternalServerError, map[string]string{
			ContentType: ApplicationJson_Utf8,
			Date:        "",
		}, map[string]interface{}{
			"state":   failed,
			"message": "server failed to remove resource",
		})
		return
	}
	s.sendRespMsg(w, http.StatusOK, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}, map[string]interface{}{
		"state": success,
	})
}

func (s *server) ListFriendInfo(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := p.ByName("user_id")
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
	headers := map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}
	respBody := map[string]interface{}{
		"state":         failed,
		"page":          pageNo,
		"page_size":     pageSize,
		"results_count": 0,
		"total_count":   0,
		"data":          map[string]string{},
	}
	results, err := s.scGraph.ListFriends(context.Background(), userId, (pageNo-1)*pageSize, pageSize)
	if err != nil {
		s.logger.Errorf("failed to list friends info due to %v", err)
		respBody["message"] = "server failed to list friends info"
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}
	totalCount, err := s.scGraph.CountFriends(context.Background(), userId)
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

// TODO: This endpoint handler should be removed when the login logic handler implemented.
func (s *server) SetAuthToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	subject := "482191" // TODO: change to user_key of authenticated user.
	aud := []string{authProvider, timeline}
	token, err := jwtPrvdr.GenJwtToken(authProvider, subject, aud)

	if err != nil {
		// log the error
		s.logger.Errorf("falied to Set-cookie: %v", err)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Authorization header set in response")
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
