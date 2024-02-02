package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	otpMap    map[string]*OtpInfo
}

var _ Interface = (*server)(nil)

func New(sgrInterface socigr.Interface, logger *lg.Logger, rbacGuard *rb.RbacGuard, smsIntfce smsapi.Interface) *server {
	return &server{
		scGraph:   sgrInterface,
		logger:    logger,
		validator: vdtor.New(),
		rbac:      rbacGuard,
		smsClient: smsIntfce,
		otpMap:    map[string]*OtpInfo{},
	}
}

/**
TODO: Ad can refer to owner's address info depeding of the locationSameAsOwner.
* Change the code logic according to that.
*/

// roles: reviewer, bdOwner,
func (s *server) CreateAd(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{}
	userKey, _, statusCode, err := s.authorize(r.Context(), s.rbac.Perm.ManageBoardingAds, rbac.Create)
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
	// Check whether main contact has been validate or not
	otpInfo, ok := s.otpMap[body.MainContact]
	if !ok {
		s.logger.Info("failed to create boarding owner: otp not found")
		uh.SendDefaultResp(w, http.StatusUnauthorized, respBody)
		return
	}
	if !otpInfo.Verified {
		s.logger.Info("failed to create boarding owner: otp not verified")
		uh.SendDefaultResp(w, http.StatusUnauthorized, respBody)
		return
	}
	// NOTE: Due to cost with verifying each other contact number, we are not allowed to created other contacts.
	// REMOVE THIS IN FUTURE, ONCE WE CAN AFFORD TO VERIFY OTHER CONTACTS.
	body.OtherContacts = []string{}

	bdOwnerKey, err := s.scGraph.CreateBoardingOwner(r.Context(), body, []string{rb.BdOwner})
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

func (s *server) authorize(ctx context.Context, perm *rbac.Permission, actions ...rbac.Action) (userKey string, roles []string, statusCode int, err error) {
	roles = []string{}
	userKey, err = jwt.GetUserKey(ctx)
	if err != nil {
		statusCode = http.StatusUnauthorized
		return
	}
	// validate the role
	roles, err = jwt.GetRoles(ctx)
	if err != nil {
		userKey = ""
		roles = []string{}
		statusCode = http.StatusForbidden
		return
	}
	if !s.rbac.HasPrivileagesForDesiredRoles(roles, perm, actions...) {
		userKey = ""
		roles = []string{}
		statusCode = http.StatusForbidden
		err = fmt.Errorf("user roles does not have sufficient permissions")
		return
	}
	return
}

// roles:  bdOwner, student, reviewer
func (s *server) GetAdForMainView(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{}
	_, _, statusCode, err := s.authorize(r.Context(), s.rbac.Perm.ManageBoardingAds, rbac.Read)
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
	_, _, statusCode, err := s.authorize(r.Context(), s.rbac.Perm.ManageBoardingAds, s.rbac.Action.Accept, s.rbac.Action.Reject)
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
	_, _, statusCode, err := s.authorize(r.Context(), s.rbac.Perm.ManageBoardingAds, rbac.Read)
	if err != nil {
		s.logger.Infof("failed to list ads: %v", err)
		uh.SendDefaultResp(w, statusCode, respBody)
		return
	}
	data := dtm.ConvertQueryParams(r)
	result, resultsCount, total, err := s.scGraph.ListAdsWithFilters(r.Context(), data)
	if err != nil {
		s.logger.Infof("failed to list ads: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	respBody["pg"] = data.Pg
	respBody["pgSize"] = data.PgSize
	respBody["resultsCount"] = resultsCount
	respBody["total"] = total
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
	// check whether there is an account for given phone number
	isExist, err := s.scGraph.IsBoardingOwnerExist(r.Context(), body.PhoneNo)
	if er.IsConflictError(err) {
		s.logger.Infof("failed to send otp: multiple users for %s: %v", body.PhoneNo, err)
		uh.SendDefaultResp(w, http.StatusConflict, respBody)
		return
	}
	if err != nil {
		s.logger.Infof("failed to send otp: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	if isExist {
		s.logger.Infof("failed to send otp: already exist: %v", err)
		uh.SendDefaultResp(w, http.StatusConflict, respBody)
		return
	}
	// check whether given phone no has already been verified or not.
	if otpInfo, ok := s.otpMap[body.PhoneNo]; ok && otpInfo.Verified {
		uh.SendDefaultResp(w, http.StatusNoContent, respBody)
		return
	}

	otp := mt.GenRandomNumber()
	if err = s.smsClient.SendSms(r.Context(), fromNx, fmt.Sprintf(otpMsg, otp), phoneNo); err != nil {
		s.logger.Infof("failed to send otp: %v", err)
		uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
		return
	}
	expAt := time.Now().Add(5 * time.Minute).Unix()
	s.otpMap[body.PhoneNo] = &OtpInfo{
		Otp:      otp,
		ExpAt:    expAt,
		Verified: false,
	}
	respBody["expAt"] = expAt
	uh.SendDefaultResp(w, http.StatusOK, respBody)
}

func (s *server) VerifyOTP(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	respBody := map[string]interface{}{}

	// read json body
	body, err := dtm.ReadJsonBody[dtm.UserInputOtp](r, s.validator)
	if err != nil {
		s.logger.Infof("failed to verify otp: %v", err)
		uh.SendDefaultResp(w, http.StatusBadRequest, respBody)
		return
	}
	otpInfo, ok := s.otpMap[body.PhoneNo]
	// No record about the sent OTP
	if !ok {
		uh.SendDefaultResp(w, http.StatusUnauthorized, respBody)
		return
	}
	// Already verified
	if otpInfo.Verified {
		uh.SendDefaultResp(w, http.StatusNoContent, respBody)
		return
	}
	// Check expiration & equality of otp
	if time.Now().Unix() > otpInfo.ExpAt || otpInfo.Otp != body.Otp {
		uh.SendDefaultResp(w, http.StatusUnauthorized, respBody)
		return
	}
	otpInfo.Verified = true
	uh.SendDefaultResp(w, http.StatusOK, respBody)
}

func (s *server) DeleteAd(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	respBody := map[string]interface{}{}
	userKey, roles, statusCode, err := s.authorize(r.Context(), s.rbac.Perm.ManageBoardingAds, rbac.Delete)
	if err != nil {
		s.logger.Infof("failed to delete ad: %v", err)
		uh.SendDefaultResp(w, statusCode, respBody)
		return
	}
	adKey := p.ByName("adKey")
	allowDel := true
	if s.rbac.HasDesiredRole(rb.BdOwner, roles) {
		// check whether given ad is own by this owner
		allowDel, err = s.scGraph.IsAdOwner(r.Context(), adKey, userKey)
		if err != nil {
			s.logger.Infof("failed to delete ad: %v", err)
			uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
			return
		}
	}
	if allowDel {
		// delete the ad
		err = s.scGraph.DeleteAd(r.Context(), adKey, userKey)
		if er.IsNotFoundError(err) {
			s.logger.Infof("failed to delete ad: %v", err)
			uh.SendDefaultResp(w, http.StatusNotFound, respBody)
			return
		}
		if err != nil {
			s.logger.Infof("failed to delete ad: %v", err)
			uh.SendDefaultResp(w, http.StatusInternalServerError, respBody)
			return
		}
		uh.SendDefaultResp(w, http.StatusOK, respBody)
		return
	}
	s.logger.Info("failed to delete ad: not allow to delete")
	uh.SendDefaultResp(w, http.StatusBadRequest, respBody)
}
