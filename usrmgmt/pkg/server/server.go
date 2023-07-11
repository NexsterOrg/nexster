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
			"err_msg": "request body is in wrong format",
		})
		w.Write(resp)
		return
	}
	if err = vdtor.New().Struct(data); err != nil {
		s.logger.Errorf("required fields are not in friend req json content, Error: %v", err)
		s.setResponseHeaders(w, http.StatusBadRequest, map[string]string{Date: ""})
		resp, _ := json.Marshal(map[string]string{
			"state":   "failed",
			"err_msg": "required fields are missing in request body",
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
			"err_msg": "failed to create required resources",
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

func (s *server) setResponseHeaders(w http.ResponseWriter, statusCode int, headers map[string]string) {
	for key, val := range headers {
		w.Header().Add(key, val)
	}
	w.WriteHeader(statusCode)
}
