package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// message body information
const ContentType string = "Content-Type"
const ContentLength string = "Content-Length"

const Date string = "Date"

// message body information - ContentType
const ApplicationJson_Utf8 string = "application/json; charset=utf-8"

type Interface interface {
	CreateNewFriendReq(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	RemovePendingFriendReq(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	RemoveFriendship(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	CreateFriendLink(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ListFriendInfo(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	GetProfile(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	GetFriendsCount(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	GetUserKeyByIndexNo(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ListFriendReqs(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	GetAllFriendReqsCount(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	EditBasicProfileInfo(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	DeleteUser(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	ResetPassword(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	GetAccessToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	EmailAccountCreationLink(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	CreateUserAccount(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	ValidatePasswordResetLink(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	GetAllUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) // all user 
}

type FriendRequest struct {
	Key     string `json:"friendreq_id"` // BUG
	From    string `json:"requestor"`
	To      string `json:"friend" validate:"required"`
	Mode    string `json:"mode" validate:"required"`
	State   string `json:"state" validate:"required"`
	ReqDate string `json:"req_date"`
}

type FriendReqAcceptance struct {
	User1Key string `json:"reqstor_id" validate:"required"`
	// User2Key   string `json:"acceptor_id" validate:"required"`
	// AcceptedAt string `json:"accepted_at" validate:"required"`
}

type ServerConfig struct {
	SecretHmacKey             string   `yaml:"secretHmacKey"`
	FrontendDomain            string   `yaml:"frontendDomain"`
	FeAccountRegPath          string   `yaml:"feAccountRegPath"`
	FePasswdResetLinkPath     string   `yaml:"fePasswdResetLinkPath"`
	ProjectDir                string   `yaml:"projectDir"`
	PublicKeyPath             string   `yaml:"publicKeyPath"`
	PrivateKeyPath            string   `yaml:"privateKeyPath"`
	AllowedOrigins            []string `yaml:"allowedOrigins"`
	AllowedEmailDomain        string   `yaml:"allowedEmailDomain"`
	DefaultUserRoles          []string `yaml:"defaultUserRoles"`
	RegLinkExpireTime         int      `yaml:"regLinkExpireTime"`
	PasswdResetLinkExpireTime int      `yaml:"passwdResetLinkExpireTime"`

}
