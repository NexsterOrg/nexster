package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

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
		"state": uh.Failed,
		"data":  []*map[string]string{},
	}
	keyword := r.URL.Query().Get("q")
	results, err := s.scGraph.SearchAmongUsers(r.Context(), keyword)
	if err != nil {
		s.logger.Errorf("failed to search for user: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	uh.SendDefaultResp(w, http.StatusOK, map[string]interface{}{
		"state": uh.Success,
		"data":  results,
	})
}
