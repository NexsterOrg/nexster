package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	vdtor "github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	socigr "github.com/NamalSanjaya/nexster/usrmgmt/pkg/social_graph"
)

const (
	failed  string = "failed"
	success string = "success"
)

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
	data, err := s.readFriendReqJson(r)
	if err != nil {
		s.logger.Errorf("failed to read json content in friend req, Error: %v", err)
		s.setResponseHeaders(w, http.StatusBadRequest, map[string]string{Date: ""})
		resp, _ := json.Marshal(map[string]string{
			"state":   "failed",
			"message": "request body is in wrong format",
		})
		w.Write(resp)
		return
	}
	if err = vdtor.New().Struct(data); err != nil {
		s.logger.Errorf("required fields are not in friend req json content, Error: %v", err)
		s.setResponseHeaders(w, http.StatusBadRequest, map[string]string{Date: ""})
		resp, _ := json.Marshal(map[string]string{
			"state":   "failed",
			"message": "required fields are missing in request body",
		})
		w.Write(resp)
		return
	}
	ctx := context.Background()
	err = s.scGraph.CreateFriendReq(ctx, data.From, data.To, data.Mode, data.State, data.ReqDate)
	if err != nil {
		s.logger.Errorf("failed to create friend req edge in db, Error: %v", err)
		s.setResponseHeaders(w, http.StatusInternalServerError, map[string]string{Date: ""})
		resp, _ := json.Marshal(map[string]string{
			"state":   "failed",
			"message": "failed to create required resources",
		})
		w.Write(resp)
		return
	}
	s.setResponseHeaders(w, http.StatusOK, map[string]string{Date: ""})
	resp, _ := json.Marshal(map[string]string{
		"state": "success",
	})
	w.Write(resp)
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

func (s *server) readFriendReqJson(r *http.Request) (*FriendRequest, error) {
	data := &FriendRequest{}
	b, err := ioutil.ReadAll(r.Body)
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
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return data, err
	}
	if err = json.Unmarshal(b, &data); err != nil {
		return data, err
	}
	return data, nil
}

func (s *server) setResponseHeaders(w http.ResponseWriter, statusCode int, headers map[string]string) {
	for key, val := range headers {
		w.Header().Add(key, val)
	}
	w.WriteHeader(statusCode)
}

func (s *server) sendRespMsg(w http.ResponseWriter, statusCode int, headers map[string]string, body map[string]interface{}) {
	for key, val := range headers {
		w.Header().Add(key, val)
	}
	w.WriteHeader(statusCode)
	resp, _ := json.Marshal(body)
	w.Write(resp)
}
