package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	"github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
	urepo "github.com/NamalSanjaya/nexster/pkgs/models/user"
	socigr "github.com/NamalSanjaya/nexster/timeline/pkg/social_graph"
)

const (
	failed            string = "failed"
	success           string = "success"
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
 * TODO: Need to include logged user's relation with reactions
 * Retrieve list of most recent posts before the given time threshold.
 * Query Parameters : last_post_at, max_post_count
 * Need Owner permission
 */
func (s *server) ListRecentPostsForTimeline(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var userId, lastPostAt, postCountStr string
	emptyArr, _ := json.Marshal([]int{})
	userId = p.ByName("userid")

	// Check the role & permissions
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		// failed to convert into string, badrequest
		s.logger.Warn("failed list recent posts: unsupported user_key type in JWT token")
		s.setResponseHeaders(w, http.StatusBadRequest, map[string]string{Date: ""})
		w.Write(emptyArr)
		return
	}
	// Check Owner permissions
	if s.scGraph.GetRole(jwtUserKey, userId) != urepo.Owner {
		// Unauthorized request, what to do
		s.logger.Warn("failed list recent posts: unauthorized request")
		s.setResponseHeaders(w, http.StatusUnauthorized, map[string]string{Date: ""})
		w.Write(emptyArr)
		return
	}

	if lastPostAt = r.URL.Query().Get("last_post_at"); lastPostAt == "" {
		lastPostAt = time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02T15:04:05.000Z") // standard format
	}
	postCountStr = r.URL.Query().Get("max_post_count")
	postCount, err := strconv.Atoi(postCountStr)
	if err != nil {
		postCount = defaultPostCount
	}
	visibility := "public"
	content, err := s.scGraph.ListRecentPosts(r.Context(), userId, lastPostAt, visibility, postCount)
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

// List posts for private timeline, Need Owner Permission
func (s *server) ListPostsForOwnersTimeline(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var userId, lastPostAt, postCountStr string
	emptyArr, _ := json.Marshal([]int{})
	userId = p.ByName("userid")

	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Warn("failed list owners posts: unsupported user_key type in JWT token")
		s.setResponseHeaders(w, http.StatusBadRequest, map[string]string{Date: ""})
		w.Write(emptyArr)
		return
	}
	// Check Owner permissions
	if s.scGraph.GetRole(jwtUserKey, userId) != urepo.Owner {
		s.logger.Warn("failed list owners posts: unauthorized request")
		s.setResponseHeaders(w, http.StatusUnauthorized, map[string]string{Date: ""})
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
	content, err := s.scGraph.ListOwnersPosts(r.Context(), userId, lastPostAt, postCount)
	if err != nil {
		s.logger.Errorf("failed list owners posts due to %w", err)
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
		s.logger.Errorf("failed convert the list of owners posts into json for due to %w", err)
	}
	w.Write(body)
}

// Need Owner Permission
func (s *server) ListFriendSuggestionsForTimeline(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var userId, startedAt, noOfSugStr string
	headers := map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}
	respBody := map[string]interface{}{
		"state":         failed,
		"started_at":    "",
		"newest_at":     "",
		"results_count": 0,
		"data":          []map[string]string{},
	}
	if userId = r.URL.Query().Get("userid"); userId == "" {
		s.logger.Errorf("failed list friend suggestions since userid query parameter is empty")
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}
	// Check the role & permissions
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Warn("failed list friend suggestions: unsupported user_key type in JWT token")
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}

	// Check Owner permissions
	if s.scGraph.GetRole(jwtUserKey, userId) != urepo.Owner {
		s.logger.Warn("failed list friend suggestions: unauthorized request")
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}

	if startedAt = r.URL.Query().Get("started_at"); startedAt == "" {
		startedAt = defaultDate
	}
	respBody["started_at"] = startedAt
	noOfSugStr = r.URL.Query().Get("max_sugs")
	noOfSugs, err := strconv.Atoi(noOfSugStr)
	if err != nil {
		noOfSugs = defaultFriendSugs
	}

	content, err := s.scGraph.ListFriendSuggestions(r.Context(), userId, startedAt, noOfSugs)
	if err != nil {
		s.logger.Errorf("failed list friend suggestions due to %w", err)
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}

	ln := len(content)
	if ln != 0 {
		// no content
		respBody["newest_at"] = (*content[ln-1])["friendship_started"]
		respBody["results_count"] = ln
		respBody["data"] = content
	}
	respBody["state"] = success
	s.sendRespMsg(w, http.StatusOK, headers, respBody)
}

// Logic
// rector_id == jwtUserKey. rector should be the owner.
func (s *server) UpdateMediaReactions(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var fromUserKey, toMediaKey, reactionKey string
	headers := map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}
	respBody := map[string]interface{}{
		"state": failed,
		"data":  map[string]string{"key": ""},
	}

	reactionKey = p.ByName("reaction_id")

	if fromUserKey = r.URL.Query().Get("reactor_id"); fromUserKey == "" {
		s.logger.Errorf("failed update media reaction since reactor_id query parameter is empty")
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}

	// Check the role & permissions
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Warn("failed update media reaction: unsupported user_key type in JWT token")
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}
	// Rector_Id == Authenticated_User
	if s.scGraph.GetRole(jwtUserKey, fromUserKey) != urepo.Owner {
		s.logger.Warn("failed update media reaction: unauthorized request")
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}

	if toMediaKey = r.URL.Query().Get("media_id"); toMediaKey == "" {
		s.logger.Errorf("failed update media reaction since media_id query parameter is empty")
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}
	// Need to check this place
	// if reactionKey = r.URL.Query().Get("reaction_id"); reactionKey == "" {
	// 	s.logger.Errorf("failed update media reaction since reaction_id query parameter is empty")
	// 	s.setResponseHeaders(w, http.StatusBadRequest, map[string]string{Date: ""})
	// 	return
	// }

	data, err := s.readReactionJson(r)
	if err != nil {
		s.logger.Errorf("Unable to read the request body since request body in wrong format")
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}
	updatedKey, err := s.scGraph.UpdateMediaReaction(r.Context(), fromUserKey, toMediaKey, reactionKey, data)
	if err != nil {
		s.logger.Errorf("Failed to Update reaction with id %s for media id %s due %v.", fromUserKey, toMediaKey, err)
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}
	respBody["state"] = success
	respBody["data"] = map[string]string{"key": updatedKey}
	s.sendRespMsg(w, http.StatusOK, headers, respBody)
}

// return http code should be 201.
func (s *server) CreateMediaReactions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var fromUserKey, toMediaKey string
	headers := map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}
	respBody := map[string]interface{}{
		"state": failed,
		"data":  map[string]string{"key": ""},
	}

	if fromUserKey = r.URL.Query().Get("reactor_id"); fromUserKey == "" {
		s.logger.Errorf("failed to create media reaction since reactor_id query parameter is empty")
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}

	// Check the role & permissions
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Warn("failed create media reaction: unsupported user_key type in JWT token")
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}
	// Rector_Id == Authenticated_User
	if s.scGraph.GetRole(jwtUserKey, fromUserKey) != urepo.Owner {
		s.logger.Warn("failed create media reaction: unauthorized request")
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}

	if toMediaKey = r.URL.Query().Get("media_id"); toMediaKey == "" {
		s.logger.Errorf("failed to create media reaction since media_id query parameter is empty")
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}

	data, err := s.readReactionJson(r)
	if err != nil {
		s.logger.Errorf("unable to read the request body since request body in wrong format")
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}
	// TODO: if the resource is not newly created, status code should be 200.
	createdKey, err := s.scGraph.CreateMediaReaction(r.Context(), fromUserKey, toMediaKey, data)
	if err != nil {
		s.logger.Errorf("Failed to create reaction with id %s for media id %s due %v.", fromUserKey, toMediaKey, err)
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}
	respBody["state"] = success
	respBody["data"] = map[string]string{"key": createdKey}
	s.sendRespMsg(w, http.StatusCreated, headers, respBody)
}

func (s *server) setResponseHeaders(w http.ResponseWriter, statusCode int, headers map[string]string) {
	for key, val := range headers {
		w.Header().Add(key, val)
	}
	w.WriteHeader(statusCode)
}

func (s *server) readReactionJson(r *http.Request) (map[string]interface{}, error) {
	var data map[string]interface{}
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
