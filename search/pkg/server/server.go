package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	"github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
	uh "github.com/NamalSanjaya/nexster/pkgs/utill/http"
	socigr "github.com/NamalSanjaya/nexster/search/pkg/social_graph"
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

func (s *server) SearchForUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{
		"state":         uh.Failed,
		"results_count": 0,
		"data":          []*map[string]string{},
	}

	userKey, ok := r.Context().Value(jwt.JwtUserKey).(string)
	if !ok {
		s.logger.Info("failed at search for user: unsupported user_key type in JWT token: unauthorized request")
		uh.SendDefaultResp(w, http.StatusUnauthorized, respBody)
		return
	}
	keyword := r.URL.Query().Get("q")

	results, err := s.scGraph.SearchAmongUsers(r.Context(), keyword)
	if err != nil {
		s.logger.Errorf("failed at search for user: failed to get searched results: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}

	// Attach Friend State
	resultCount := 0
	for _, each := range results {
		state, reqId, err := s.scGraph.AttachFriendState(r.Context(), userKey, (*each)["key"])
		if err != nil {
			// TODO:
			// Remove this one from results
			s.logger.Errorf("failed at search for user: error found during attaching friend state: %v: userKey=%s, friendKey=%s\n", err, userKey, (*each)["key"])
			continue
		}
		(*each)["friend_state"] = state
		(*each)["friend_req_id"] = reqId
		resultCount++
	}

	uh.SendDefaultResp(w, http.StatusOK, map[string]interface{}{
		"state":         uh.Success,
		"results_count": resultCount,
		"data":          results,
	})
}
