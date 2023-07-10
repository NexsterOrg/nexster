package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	socigr "github.com/NamalSanjaya/nexster/timeline/pkg/social_graph"
)

const (
	defaultPostCount int = 10
)

type Server struct {
	scGraph socigr.Interface
	logger  *lg.Logger
}

func New(sgrInterface socigr.Interface, logger *lg.Logger) *Server {
	return &Server{
		scGraph: sgrInterface,
		logger:  logger,
	}
}

/*
 * Retrieve list of most recent posts before the given time threshold.
 * Query Parameters : userid, type, last_post_at, max_post_count
 */
func (s *Server) ListRecentPostsForTimeline(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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
		lastPostAt = "2023-04-29T09:31:00.000Z"
	}
	postCountStr = r.URL.Query().Get("max_post_count")
	postCount, err := strconv.Atoi(postCountStr)
	if err != nil {
		postCount = defaultPostCount
	}
	// TODO
	// Need to create arango db id for the user using userid. This following line is a temporary work
	userId = fmt.Sprintf("users/%s", userId)

	content, err := s.scGraph.ListRecentPosts(ctx, userId, lastPostAt, visibility, postCount)
	if err != nil {
		s.logger.Errorf("failed list recent posts due to %v", err)
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

func (s *Server) setResponseHeaders(w http.ResponseWriter, statusCode int, headers map[string]string) {
	for key, val := range headers {
		w.Header().Add(key, val)
	}
	w.WriteHeader(statusCode)
}
