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
	tp "github.com/NamalSanjaya/nexster/timeline/pkg/types"
)

const (
	failed            string = "failed"
	success           string = "success"
	defaultPostCount  int    = 10
	defaultFriendSugs int    = 10
	defaultDate       string = "2023-01-01T01:00:00.000Z"
	defaultPageNo     int    = 1
	defaultPageSize   int    = 10
	defaultBirthday   string = "2002-01-01"
	defaultGender     string = "male"
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
		s.logger.Warn("failed list recent posts: unsupported user_key type in JWT token: unauthorized request")
		s.setResponseHeaders(w, http.StatusUnauthorized, map[string]string{Date: ""})
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
		s.logger.Warn("failed list owners posts: unsupported user_key type in JWT token: unauthorized request")
		s.setResponseHeaders(w, http.StatusUnauthorized, map[string]string{Date: ""})
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
func (s *server) ListFriendSuggestions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{
		"state":         failed,
		"results_count": 0,
		"data":          []map[string]string{},
	}
	// Check the role & permissions
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed list friend suggestions: unsupported user_key type in JWT token: unauthorized request")
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
	respBody["page"] = pageNo
	respBody["page_size"] = pageSize
	content, err := s.scGraph.ListFriendSuggestions(r.Context(), jwtUserKey, (pageNo-1)*pageSize, pageSize)
	if err != nil {
		s.logger.Errorf("failed list friend suggestions due to %w", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	respBody["results_count"] = len(content)
	respBody["data"] = content
	respBody["state"] = success
	s.sendRespDefault(w, http.StatusOK, respBody)
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

	// Check the role & permissions
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Warn("failed update media reaction: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}
	reactionKey = p.ByName("reaction_id")

	if fromUserKey = r.URL.Query().Get("reactor_id"); fromUserKey == "" {
		s.logger.Errorf("failed update media reaction since reactor_id query parameter is empty")
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
	// Check the role & permissions
	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Warn("failed create media reaction: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}
	if fromUserKey = r.URL.Query().Get("reactor_id"); fromUserKey == "" {
		s.logger.Errorf("failed to create media reaction since reactor_id query parameter is empty")
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

// permission: owner
func (s *server) ListOwnersViewMedia(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var userKey string
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
		s.logger.Warn("failed to list owner view media: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}
	if userKey = r.URL.Query().Get("user_id"); userKey == "" {
		s.logger.Info("failed to list owner view media: user_id query parameter is empty")
		s.sendRespMsg(w, http.StatusBadRequest, headers, respBody)
		return
	}
	// Check the role & permissions
	if s.scGraph.GetRole(jwtUserKey, userKey) != urepo.Owner {
		s.logger.Warn("failed to list owner view media: unauthorized request")
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
	respBody["results_count"] = 0
	medias, err := s.scGraph.ListAllMedia(r.Context(), userKey, (pageNo-1)*pageSize, pageSize)
	if err != nil {
		s.logger.Errorf("failed to list owner view media: %v: id=%s", err, userKey)
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}
	resCount := len(medias)
	if resCount == 0 {
		respBody["data"] = []map[string]string{}
	} else {
		respBody["data"] = medias
	}
	respBody["state"] = success
	respBody["results_count"] = resCount
	s.sendRespMsg(w, http.StatusOK, headers, respBody)
}

func (s *server) ListPublicMedia(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	headers := map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}
	respBody := map[string]interface{}{
		"state": failed,
		"data":  map[string]string{},
	}
	userKey := p.ByName("user_id")

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
	respBody["results_count"] = 0

	medias, err := s.scGraph.ListPublicMedia(r.Context(), userKey, (pageNo-1)*pageSize, pageSize)
	if err != nil {
		s.logger.Errorf("failed to list public media: %v: id=%s", err, userKey)
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}
	resCount := len(medias)
	if resCount == 0 {
		respBody["data"] = []map[string]string{}
	} else {
		respBody["data"] = medias
	}
	respBody["state"] = success
	respBody["results_count"] = resCount
	s.sendRespMsg(w, http.StatusOK, headers, respBody)

}

func (s *server) ListRoleBasedMedia(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	headers := map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date:        "",
	}
	respBody := map[string]interface{}{
		"state":         failed,
		"data":          []map[string]string{},
		"results_count": 0,
	}
	imgOwnerKey := p.ByName("img_owner_id")
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

	jwtUserKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Warn("failed to list media: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespMsg(w, http.StatusUnauthorized, headers, respBody)
		return
	}

	var medias []*map[string]string
	if s.scGraph.GetRole(jwtUserKey, imgOwnerKey) != urepo.Owner {
		// list public visible images
		medias, err = s.scGraph.ListPublicMedia(r.Context(), imgOwnerKey, (pageNo-1)*pageSize, pageSize)
	} else {
		// list both private and public visible images
		medias, err = s.scGraph.ListAllMedia(r.Context(), imgOwnerKey, (pageNo-1)*pageSize, pageSize)
	}

	if err != nil {
		s.logger.Errorf("failed to list media: %v: imgOwnerKey=%s, authUserKey=%s", err, imgOwnerKey, jwtUserKey)
		s.sendRespMsg(w, http.StatusInternalServerError, headers, respBody)
		return
	}

	resCount := len(medias)
	if resCount == 0 {
		respBody["data"] = []map[string]string{}
	} else {
		respBody["data"] = medias
	}
	respBody["state"] = success
	respBody["results_count"] = resCount
	s.sendRespMsg(w, http.StatusOK, headers, respBody)
}

func (s *server) ListFriendSuggestionsV2(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{
		"state":         failed,
		"data":          []map[string]string{},
		"results_count": 0,
	}
	faculty := p.ByName("faculty")
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

	userKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Warn("failed to list friend suggs: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}
	birthday := r.URL.Query().Get("birthday")
	if birthday == "" {
		birthday = defaultBirthday
	}
	gender := r.URL.Query().Get("gender")
	if gender == "" {
		gender = defaultGender
	}
	friends, err := s.scGraph.ListFriendSuggsV2(r.Context(), userKey, birthday, faculty, gender, (pageNo-1)*pageSize, pageSize)
	if err != nil {
		s.logger.Errorf("failed to list friend suggs-v2: %v: userKey=%s", err, userKey)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	// Attach Friend State
	resultCount := 0
	for _, each := range friends {
		state, reqId, err := s.scGraph.AttachFriendState(r.Context(), userKey, (*each)["key"])
		if err != nil {
			s.logger.Errorf("error found during attaching friend state: %v: userKey=%s, friendKey=%s\n", err, userKey, (*each)["key"])
			continue
		}
		(*each)["friend_state"] = state
		(*each)["friend_req_id"] = reqId
		resultCount++
	}
	respBody["state"] = success
	respBody["results_count"] = resultCount
	respBody["data"] = friends
	s.sendRespDefault(w, http.StatusOK, respBody)
}

// permission : owner
func (s *server) CreateImagePost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{
		"state": failed,
		"data":  map[string]string{},
	}
	userKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed to create image post: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}
	body, err := tp.ReadJsonBody[tp.Post](r)
	if err != nil {
		s.logger.Infof("failed to create image post: unable to read request body: %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, respBody)
		return
	}
	mediaKey, mediaOwnerKey, err := s.scGraph.CreateImagePost(r.Context(), userKey, body)
	if err != nil {
		s.logger.Infof("failed to create image post: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	s.sendRespDefault(w, http.StatusCreated, map[string]interface{}{
		"state":         success,
		"mediaKey":      mediaKey,
		"mediaOwnerKey": mediaOwnerKey,
	})
}

// permission : owner
func (s *server) DeleteImagePost(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{
		"state": failed,
	}
	userKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed to create image post: unsupported user_key type in JWT token: unauthorized request")
		s.sendRespDefault(w, http.StatusUnauthorized, respBody)
		return
	}
	mediaKey := p.ByName("mediaKey")
	if err := s.scGraph.DeleteImagePost(r.Context(), userKey, mediaKey); err != nil {
		s.logger.Errorf("failed to delete image post: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, respBody)
		return
	}
	s.sendRespDefault(w, http.StatusOK, map[string]interface{}{
		"state": success,
	})
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

// similar to `sendRespMsg` but only have predefined headers
func (s *server) sendRespDefault(w http.ResponseWriter, statusCode int, body map[string]interface{}) {
	w.Header().Add(ContentType, ApplicationJson_Utf8)
	w.Header().Add(Date, "")
	w.WriteHeader(statusCode)
	resp, _ := json.Marshal(body)
	w.Write(resp)
}
