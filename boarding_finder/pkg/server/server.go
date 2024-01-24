package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/euroteltr/rbac"
	vdtor "github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	dtm "github.com/NamalSanjaya/nexster/boarding_finder/pkg/dto_mapper"
	rb "github.com/NamalSanjaya/nexster/boarding_finder/pkg/rbac"
	socigr "github.com/NamalSanjaya/nexster/boarding_finder/pkg/social_graph"
	"github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
	er "github.com/NamalSanjaya/nexster/pkgs/errors"
	uh "github.com/NamalSanjaya/nexster/pkgs/utill/http"
)

type server struct {
	logger    *lg.Logger
	scGraph   socigr.Interface
	validator *vdtor.Validate
	rbac      *rb.RbacGuard
}

var _ Interface = (*server)(nil)

func New(sgrInterface socigr.Interface, logger *lg.Logger, rbacGuard *rb.RbacGuard) *server {
	return &server{
		scGraph:   sgrInterface,
		logger:    logger,
		validator: vdtor.New(),
		rbac:      rbacGuard,
	}
}

// roles: reviewer, bdOwner
func (s *server) CreateAd(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{}
	userKey, statusCode, err := s.authorize(r.Context(), s.rbac.Perm.ManageBoardingAds, rbac.Create)
	if err != nil {
		s.logger.Infof("failed to create boarding ad: %v", err)
		uh.SendDefaultResp(w, statusCode, respBody)
		return
	}
	// read json body
	body, err := dtm.ReadJsonBody[dtm.CreateAdDto](r, s.validator)
	if err != nil {
		s.logger.Infof("failed to create boarding ad: %v", err)
		uh.SendDefaultResp(w, http.StatusBadRequest, respBody)
		return
	}
	adKey, ownedEdgeKey, err := s.scGraph.CreateAd(r.Context(), userKey, body)
	if er.IsNotFoundError(err) {
		s.logger.Infof("failed to create boarding ad: %v", err)
		uh.SendDefaultResp(w, http.StatusNotFound, respBody)
		return
	}
	if err != nil {
		s.logger.Errorf("failed to create boarding ad: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	respBody["adId"] = adKey
	respBody["ownedId"] = ownedEdgeKey
	uh.SendDefaultResp(w, http.StatusCreated, respBody)
}

// Request to create a boarding owner account. Request goes to pending status.
func (s *server) CreateBoardingOwner(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{}

	// read json body
	body, err := dtm.ReadJsonBody[dtm.CreateBoardingOwner](r, s.validator)
	if err != nil {
		s.logger.Infof("failed to create boarding owner: %v", err)
		uh.SendDefaultResp(w, http.StatusBadRequest, respBody)
		return
	}
	// 1. Check in redis cache whether main contact number is verified. IF YES proceed, ELSE return here, don't create account.
	// 2. Check otherContacts are verified or not and mark verified one. [ {phoneNo, verified}, {} ]
	bdOwnerKey, err := s.scGraph.CreateBoardingOwner(r.Context(), body)
	if er.IsConflictError(err) {
		s.logger.Errorf("failed to create boarding owner: %v", err)
		uh.SendDefaultResp(w, http.StatusConflict, respBody)
		return
	}
	if err != nil {
		s.logger.Errorf("failed to create boarding owner: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	respBody["id"] = bdOwnerKey
	uh.SendDefaultResp(w, http.StatusCreated, respBody)
}

func (s *server) authorize(ctx context.Context, perm *rbac.Permission, actions ...rbac.Action) (userKey string, statusCode int, err error) {
	userKey, err = jwt.GetUserKey(ctx)
	if err != nil {
		statusCode = http.StatusUnauthorized
		return
	}
	// validate the role
	roles, err := jwt.GetRoles(ctx)
	if err != nil {
		userKey = ""
		statusCode = http.StatusForbidden
		return
	}
	if !s.rbac.HasDesiredRole(roles, perm, actions...) {
		userKey = ""
		statusCode = http.StatusForbidden
		err = fmt.Errorf("user roles does not have sufficient permissions")
		return
	}
	return
}
