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
	smsapi "github.com/NamalSanjaya/nexster/pkgs/client/sms_api"
	er "github.com/NamalSanjaya/nexster/pkgs/errors"
	uh "github.com/NamalSanjaya/nexster/pkgs/utill/http"
	mt "github.com/NamalSanjaya/nexster/pkgs/utill/math"
	strg "github.com/NamalSanjaya/nexster/pkgs/utill/string"
)

const otpMsg string = "Nexster code: %d. Valid for 5 minutes."
const fromNx string = "Nexster"

type server struct {
	logger    *lg.Logger
	scGraph   socigr.Interface
	validator *vdtor.Validate
	rbac      *rb.RbacGuard
	smsClient smsapi.Interface
}

var _ Interface = (*server)(nil)

func New(sgrInterface socigr.Interface, logger *lg.Logger, rbacGuard *rb.RbacGuard, smsIntfce smsapi.Interface) *server {
	return &server{
		scGraph:   sgrInterface,
		logger:    logger,
		validator: vdtor.New(),
		rbac:      rbacGuard,
		smsClient: smsIntfce,
	}
}

/**
TODO: Ad can refer to owner's address info depeding of the locationSameAsOwner.
* Change the code logic according to that.
*/

// roles: reviewer, bdOwner,
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

// roles:  bdOwner, student, reviewer
func (s *server) GetAdForMainView(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{}
	_, statusCode, err := s.authorize(r.Context(), s.rbac.Perm.ManageBoardingAds, rbac.Read)
	if err != nil {
		s.logger.Infof("failed to get ad: %v", err)
		uh.SendDefaultResp(w, statusCode, respBody)
		return
	}
	result, err := s.scGraph.GetAdForMainView(r.Context(), p.ByName("adKey"))
	if er.IsNotFoundError(err) {
		s.logger.Infof("failed to get ad: %v", err)
		uh.SendDefaultResp(w, http.StatusNotFound, respBody)
		return
	}
	if er.IsConflictError(err) {
		s.logger.Infof("failed to get ad: %v", err)
		uh.SendDefaultResp(w, http.StatusConflict, respBody)
		return
	}
	if err != nil {
		s.logger.Infof("failed to get ad: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	uh.SendDefaultRespAny(w, http.StatusOK, result)
}

// roles: reviewer (Need to write a api to create reviewers - usrmgmt api). reviewer --> member
func (s *server) ChangeStatusOfAd(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{}
	_, statusCode, err := s.authorize(r.Context(), s.rbac.Perm.ManageBoardingAds, s.rbac.Action.Accept, s.rbac.Action.Reject)
	if err != nil {
		s.logger.Infof("failed to change status of ad: %v", err)
		uh.SendDefaultResp(w, statusCode, respBody)
		return
	}
	// read json body
	body, err := dtm.ReadJsonBody[dtm.AdStatus](r, s.validator)
	if err != nil {
		s.logger.Infof("failed to change status of ad: %v", err)
		uh.SendDefaultResp(w, http.StatusBadRequest, respBody)
		return
	}

	err = s.scGraph.ChangeAdStatus(r.Context(), p.ByName("adKey"), body.Status)
	if er.IsNotFoundError(err) {
		s.logger.Infof("failed to change status of ad: %v", err)
		uh.SendDefaultResp(w, http.StatusNotFound, respBody)
		return
	}
	if err != nil {
		s.logger.Infof("failed to change status of ad: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	uh.SendDefaultResp(w, http.StatusNoContent, respBody)
}

// roles; bdOwner, reviewer, student
func (s *server) ListAdsForMainView(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{}
	_, statusCode, err := s.authorize(r.Context(), s.rbac.Perm.ManageBoardingAds, rbac.Read)
	if err != nil {
		s.logger.Infof("failed to list ads: %v", err)
		uh.SendDefaultResp(w, statusCode, respBody)
		return
	}
	data := dtm.ConvertQueryParams(r)
	result, resultsCount, err := s.scGraph.ListAdsWithFilters(r.Context(), data)
	if err != nil {
		s.logger.Infof("failed to list ads: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	respBody["pg"] = data.Pg
	respBody["pgSize"] = data.PgSize
	respBody["resultsCount"] = resultsCount
	respBody["data"] = result
	uh.SendDefaultResp(w, http.StatusOK, respBody)
}

func (s *server) SendOTP(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{}

	// read json body
	body, err := dtm.ReadJsonBody[dtm.Otp](r, s.validator)
	if err != nil {
		s.logger.Infof("failed to send otp: %v", err)
		uh.SendDefaultResp(w, http.StatusBadRequest, respBody)
		return
	}

	phoneNo, err := strg.ConvertToValidMobileNo(body.PhoneNo)
	if err != nil {
		s.logger.Info("failed to send otp: invalid phoneNo")
		uh.SendDefaultResp(w, http.StatusBadRequest, respBody)
		return
	}
	if err = s.smsClient.SendSms(r.Context(), fromNx, fmt.Sprintf(otpMsg, mt.GenRandomNumber()), phoneNo); err != nil {
		s.logger.Infof("failed to send otp: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	uh.SendDefaultResp(w, http.StatusNoContent, respBody)
}
