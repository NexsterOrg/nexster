package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	socigr "github.com/NamalSanjaya/nexster/timeline/pkg/social_graph"
)

const (
	defaultPostCount  int    = 10
	defaultFriendSugs int    = 10
	defaultDate       string = "2023-01-01T01:00:00.000Z"
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

/*
 * Retrieve list of most recent posts before the given time threshold.
 * Query Parameters : userid, type, last_post_at, max_post_count
 */
func (s *server) ListRecentPostsForTimeline(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var userId, visibility, lastPostAt, postCountStr string
	ctx := context.Background()
	emptyArr, _ := json.Marshal([]int{})
	if userId = r.URL.Query().Get("userid"); userId == "" {
		s.logger.Errorf("failed list recent post since userid query parameter is empty")
		s.setResponseHeaders(w, http.StatusBadRequest, map[string]string{Date: ""})
		w.Write(emptyArr)
		return
	}
	if visibility = r.URL.Query().Get("type"); visibility == "" {
		s.logger.Errorf("failed list recent post since type query parameter is empty")
		s.setResponseHeaders(w, http.StatusBadRequest, map[string]string{Date: ""})
		w.Write(emptyArr)
		return
	}
	if lastPostAt = r.URL.Query().Get("last_post_at"); lastPostAt == "" {
		lastPostAt = time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02T15:04:05.000Z")
	}
	postCountStr = r.URL.Query().Get("max_post_count")
	postCount, err := strconv.Atoi(postCountStr)
	if err != nil {
		postCount = defaultPostCount
	}

	content, err := s.scGraph.ListRecentPosts(ctx, userId, lastPostAt, visibility, postCount)
	if err != nil {
		s.logger.Errorf("failed list recent posts due to %w", err)
		s.setResponseHeaders(w, http.StatusInternalServerError, map[string]string{Date: ""})
		w.Write(emptyArr)
		return
	}
	s.setResponseHeaders(w, http.StatusOK, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	})
	if content == nil {
		w.Write(emptyArr)
		return
	}
	body, err := json.Marshal(content)
	if err != nil {
		s.logger.Errorf("failed convert the list of posts into json for due to %w", err)
	}
	w.Write(body)
}

func (s *server) ListFriendSuggestionsForTimeline(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var userId, startedAt, noOfSugStr string
	ctx := context.Background()
	emptyArr, _ := json.Marshal([]int{})
	if userId = r.URL.Query().Get("userid"); userId == "" {
		s.logger.Errorf("failed list friend suggestions since userid query parameter is empty")
		s.setResponseHeaders(w, http.StatusBadRequest, map[string]string{Date: ""})
		w.Write(emptyArr)
		return
	}
	if startedAt = r.URL.Query().Get("started_at"); startedAt == "" {
		startedAt = defaultDate
	}
	noOfSugStr = r.URL.Query().Get("max_sugs")
	noOfSugs, err := strconv.Atoi(noOfSugStr)
	if err != nil {
		noOfSugs = defaultFriendSugs
	}

	content, err := s.scGraph.ListFriendSuggestions(ctx, userId, startedAt, noOfSugs)
	if err != nil {
		s.logger.Errorf("failed list friend suggestions due to %w", err)
		s.setResponseHeaders(w, http.StatusInternalServerError, map[string]string{Date: ""})
		w.Write(emptyArr)
		return
	}

	s.setResponseHeaders(w, http.StatusOK, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	})
	if content == nil {
		w.Write(emptyArr)
		return
	}
	body, err := json.Marshal(content)
	if err != nil {
		s.logger.Errorf("failed convert the list of friend suggestions into json for due to %w", err)
	}
	w.Write(body)
}

func (s *server) UpdateMediaReactions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var fromUserKey, toMediaKey, reactionKey string
	if fromUserKey = r.URL.Query().Get("reactor_id"); fromUserKey == "" {
		s.logger.Errorf("failed update media reaction since reactor_id query parameter is empty")
		s.setResponseHeaders(w, http.StatusBadRequest, map[string]string{Date: ""})
		return
	}

	if toMediaKey = r.URL.Query().Get("media_id"); toMediaKey == "" {
		s.logger.Errorf("failed update media reaction since media_id query parameter is empty")
		s.setResponseHeaders(w, http.StatusBadRequest, map[string]string{Date: ""})
		return
	}

	if reactionKey = r.URL.Query().Get("reaction_id"); reactionKey == "" {
		s.logger.Errorf("failed update media reaction since reaction_id query parameter is empty")
		s.setResponseHeaders(w, http.StatusBadRequest, map[string]string{Date: ""})
		return
	}

	data, err := s.readReactionJson(r)
	if err != nil {
		s.logger.Errorf("Unable to read the request body since request body in wrong format")
		s.setResponseHeaders(w, http.StatusBadRequest, map[string]string{Date: ""})
		return
	}
	ctx := context.Background()
	err = s.scGraph.UpdateMediaReaction(ctx, fromUserKey, toMediaKey, reactionKey, data)
	if err != nil {
		s.logger.Errorf("Failed to Update reaction with id %s for media id %s due %v\n", fromUserKey, toMediaKey, err)
		s.setResponseHeaders(w, http.StatusInternalServerError, map[string]string{Date: ""})
		return
	}
	s.setResponseHeaders(w, http.StatusOK, map[string]string{Date: ""})
}

func (s *server) setResponseHeaders(w http.ResponseWriter, statusCode int, headers map[string]string) {
	for key, val := range headers {
		w.Header().Add(key, val)
	}
	w.WriteHeader(statusCode)
}

func (s *server) readReactionJson(r *http.Request) (map[string]interface{}, error) {
	var data map[string]interface{}
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
