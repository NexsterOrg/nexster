package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	vdtor "github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	"github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
	"github.com/NamalSanjaya/nexster/pkgs/errors"
	uh "github.com/NamalSanjaya/nexster/pkgs/utill/http"
	socigr "github.com/NamalSanjaya/nexster/space/pkg/social_graph"
	tp "github.com/NamalSanjaya/nexster/space/pkg/types"
)

type server struct {
	logger  *lg.Logger
	scGraph socigr.Interface
}

var _ Interface = (*server)(nil)

func New(sgrInterface socigr.Interface, logger *lg.Logger) *server {
	return &server{
		scGraph: sgrInterface,
		logger:  logger,
	}
}

func (s *server) CreateEventInSpace(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{
		"state": uh.Failed,
		"data":  map[string]string{},
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed to create event: unsupported user_key type in JWT token: unauthorized request")
		uh.SendDefaultResp(w, http.StatusUnauthorized, respBody)
		return
	}
	data, err := s.readJsonEventBody(r)
	if err != nil {
		s.logger.Infof("failed to create event: invalid request body: %v", err)
		uh.SendDefaultResp(w, http.StatusBadRequest, respBody)
		return
	}
	if err = vdtor.New().Struct(data); err != nil {
		s.logger.Infof("failed to create event: some mandadary fields are missing in request body: %v", err)
		uh.SendDefaultResp(w, http.StatusBadRequest, respBody)
		return
	}

	eventKey, postedByKey, err := s.scGraph.CreateEvent(r.Context(), jwtUserKey, data)
	if err != nil {
		s.logger.Errorf("failed to create event: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	respBody["state"] = uh.Success
	respBody["data"] = map[string]string{
		"eventKey":    eventKey,
		"postedByKey": postedByKey,
	}
	uh.SendDefaultResp(w, http.StatusCreated, respBody)
}

// Viewer permission
func (s *server) ListUpcomingEventsFromSpace(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{
		"state": uh.Failed,
		"data":  []map[string]string{},
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed to list upcoming events: unsupported user_key type in JWT token: unauthorized request")
		uh.SendDefaultResp(w, http.StatusUnauthorized, respBody)
		return
	}
	pageNo, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		pageNo = uh.DefaultPageNo
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil {
		pageSize = uh.DefaultPageSize
	}
	events, err := s.scGraph.ListUpcomingEvents(r.Context(), jwtUserKey, (pageNo-1)*pageSize, pageSize)
	if err != nil {
		s.logger.Errorf("failed to list latest events: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	uh.SendDefaultResp(w, http.StatusOK, map[string]interface{}{
		"state":        uh.Success,
		"page":         pageNo,
		"pageSize":     pageSize,
		"resultsCount": len(events),
		"data":         events,
	})
}

func (s *server) GetEventFromSpace(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{
		"state": uh.Failed,
		"data":  map[string]string{},
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed to get event info: unsupported user_key type in JWT token: unauthorized request")
		uh.SendDefaultResp(w, http.StatusUnauthorized, respBody)
		return
	}
	eventKey := p.ByName("eventKey")
	event, err := s.scGraph.GetEvent(r.Context(), jwtUserKey, eventKey)
	if errors.IsNotFoundError(err) {
		s.logger.Infof("failed to get event info: event is not found: eventKey=%s", eventKey)
		uh.SendDefaultResp(w, http.StatusNotFound, respBody)
		return
	}
	if err != nil {
		s.logger.Errorf("failed to get event info: eventKey=%s: %v", eventKey, err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	uh.SendDefaultResp(w, http.StatusOK, map[string]interface{}{
		"state": uh.Success,
		"data":  event,
	})
}

func (s *server) readJsonEventBody(r *http.Request) (*tp.Event, error) {
	data := &tp.Event{}
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

// viewer permission
func (s *server) ListLoveReactUsersForEvent(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{
		"state": uh.Failed,
		"data":  []map[string]string{},
	}
	reactType := p.ByName("reactType")
	eventKey := p.ByName("eventKey")
	pageNo, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		pageNo = uh.DefaultPageNo
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil {
		pageSize = uh.DefaultPageSize
	}
	eventLovers, err := s.scGraph.ListEventReactUsersForType(r.Context(), eventKey, reactType, (pageNo-1)*pageSize, pageSize)
	if err != nil {
		s.logger.Errorf("failed to list %s react users: eventKey=%s, %v", reactType, eventKey, err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	uh.SendDefaultResp(w, http.StatusOK, map[string]interface{}{
		"state":        uh.Success,
		"page":         pageNo,
		"pageSize":     pageSize,
		"resultsCount": len(eventLovers),
		"data":         eventLovers,
	})
}

func (s *server) CreateEventReaction(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{
		"state": uh.Failed,
		"data":  map[string]string{},
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed to create event reaction: unsupported user_key type in JWT token: unauthorized request")
		uh.SendDefaultResp(w, http.StatusUnauthorized, respBody)
		return
	}
	body, err := tp.ReadJsonBody[tp.EventReaction](r)
	if err != nil {
		s.logger.Infof("failed to create event reaction: request body is invalid: %v", err)
		uh.SendDefaultResp(w, http.StatusBadRequest, respBody)
		return
	}
	newEventReactKey, err := s.scGraph.CreateEventReactionEdge(r.Context(), jwtUserKey, p.ByName("eventKey"), body)
	if err != nil {
		s.logger.Infof("failed to create event reaction: %v", err)
		status := http.StatusInternalServerError
		if errors.IsNotConflictError(err) {
			status = http.StatusConflict
		} else if errors.IsNotFoundError(err) {
			status = http.StatusNotFound
		}
		uh.SendDefaultResp(w, status, respBody)
		return
	}
	uh.SendDefaultResp(w, http.StatusCreated, map[string]interface{}{
		"state": uh.Success,
		"data":  map[string]string{"key": newEventReactKey},
	})
}

func (s *server) SetEventReactionState(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{"state": uh.Failed}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed to set event reaction: unsupported user_key type in JWT token: unauthorized request")
		uh.SendDefaultResp(w, http.StatusUnauthorized, respBody)
		return
	}
	state, err := strconv.ParseBool(p.ByName("state"))
	if err != nil {
		s.logger.Infof("failed to set event reaction: failed to convert state to boolean: %s: %v", p.ByName("state"), err)
		uh.SendDefaultResp(w, http.StatusBadRequest, respBody)
		return
	}
	reactionType := p.ByName("reactionType")
	if reactionType != "love" && reactionType != "going" {
		s.logger.Info("failed to set event reaction: unsupported reaction type given")
		uh.SendDefaultResp(w, http.StatusBadRequest, respBody)
		return
	}
	if err = s.scGraph.SetEventReactionState(r.Context(), jwtUserKey, p.ByName("reactionKey"),
		map[string]bool{reactionType: state}); err != nil {
		statusCode := http.StatusInternalServerError
		if errors.IsNotFoundError(err) {
			statusCode = http.StatusNotFound

		} else if errors.IsUnAuthError(err) {
			statusCode = http.StatusUnauthorized
		}
		s.logger.Infof("failed to set event reaction: userKey=%s, %v", jwtUserKey, err)
		uh.SendDefaultResp(w, statusCode, respBody)
		return
	}
	respBody["state"] = uh.Success
	uh.SendDefaultResp(w, http.StatusOK, respBody)
}

// Owner permission
func (s *server) ListMyEventsFromSpace(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{
		"state": uh.Failed,
		"data":  []map[string]string{},
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed to list my events: unsupported user_key type in JWT token: unauthorized request")
		uh.SendDefaultResp(w, http.StatusUnauthorized, respBody)
		return
	}
	pageNo, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		pageNo = uh.DefaultPageNo
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil {
		pageSize = uh.DefaultPageSize
	}
	events, err := s.scGraph.ListMyEvents(r.Context(), jwtUserKey, (pageNo-1)*pageSize, pageSize)
	if err != nil {
		s.logger.Errorf("failed to list my events: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	uh.SendDefaultResp(w, http.StatusOK, map[string]interface{}{
		"state":        uh.Success,
		"page":         pageNo,
		"pageSize":     pageSize,
		"resultsCount": len(events),
		"data":         events,
	})
}

func (s *server) DeleteEventFromSpace(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{
		"state": uh.Failed,
	}
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed to remove event: unsupported user_key type in JWT token: unauthorized request")
		uh.SendDefaultResp(w, http.StatusUnauthorized, respBody)
		return
	}
	eventKey := p.ByName("eventKey")
	if err := s.scGraph.DeleteEvent(r.Context(), jwtUserKey, eventKey); err != nil {
		s.logger.Errorf("failed to remove event: %v: eventKey=%s", err, eventKey)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	uh.SendDefaultResp(w, http.StatusOK, map[string]interface{}{"state": uh.Success})
}
