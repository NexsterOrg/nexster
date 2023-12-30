package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"
	"github.com/rs/cors"
	"gopkg.in/yaml.v3"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	jwtAuth "github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
	cl "github.com/NamalSanjaya/nexster/pkgs/client"
	contapi "github.com/NamalSanjaya/nexster/pkgs/client/content_api"
	avtr "github.com/NamalSanjaya/nexster/pkgs/models/avatar"
	fac "github.com/NamalSanjaya/nexster/pkgs/models/faculty"
	frnd "github.com/NamalSanjaya/nexster/pkgs/models/friend"
	freq "github.com/NamalSanjaya/nexster/pkgs/models/friend_request"
	hgen "github.com/NamalSanjaya/nexster/pkgs/models/hasGender"
	stdt "github.com/NamalSanjaya/nexster/pkgs/models/student"
	usr "github.com/NamalSanjaya/nexster/pkgs/models/user"
	umail "github.com/NamalSanjaya/nexster/pkgs/utill/mail"
	authprv "github.com/NamalSanjaya/nexster/usrmgmt/pkg/auth_provider"
	usrv "github.com/NamalSanjaya/nexster/usrmgmt/pkg/server"
	socigr "github.com/NamalSanjaya/nexster/usrmgmt/pkg/social_graph"
)

type Configs struct {
	Server           usrv.ServerConfig   `yaml:"server"`
	ArgDbCfg         argdb.Config        `yaml:"arangodb"`
	ContentClientCfg cl.HttpClientConfig `yaml:"content"`
	MailCfg          umail.MailConfig    `yaml:"mail"`
}

const issuer string = "usrmgmt"

func main() {
	ctx := context.Background()

	yamlFile, err := os.ReadFile("../configs/config.yaml")
	if err != nil {
		log.Panicf("Error reading YAML file: %v", err)
	}
	var configs Configs
	if err := yaml.Unmarshal(yamlFile, &configs); err != nil {
		log.Panicf("Error unmarshaling YAML: %v", err)
	}
	logger := lg.New("UserMgmtSrv")
	logger.EnableColor()

	argFrndReqClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, freq.FriendReqColl)
	argFrndClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, frnd.FriendColl)
	argUsrClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, usr.UsersColl)
	argAvtrClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, avtr.AvatarColl)
	argStudentClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, stdt.StudnetColl)
	argFacultyClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, fac.FacultyColl)
	argHasGenderClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, hgen.HasGenderColl)

	frReqCtrler := freq.NewCtrler(argFrndReqClient)
	frndCtrler := frnd.NewCtrler(argFrndClient)
	usrCtrler := usr.NewCtrler(argUsrClient)
	avtrCtrler := avtr.NewCtrler(argAvtrClient)
	stdtCtrler := stdt.NewCtrler(argStudentClient)
	facCtrler := fac.NewCtrler(argFacultyClient)
	hasGenCtrler := hgen.NewCtrler(argHasGenderClient)

	// API clients
	contentApiClient := contapi.NewApiClient(&configs.ContentClientCfg)

	// mail client
	mailClient := umail.New(&configs.MailCfg)

	grCtrler := socigr.NewGrphCtrler(frReqCtrler, frndCtrler, usrCtrler, contentApiClient, avtrCtrler, stdtCtrler, facCtrler, hasGenCtrler)
	srv := usrv.New(&configs.Server, grCtrler, logger, mailClient)

	router := httprouter.New()

	c := cors.New(cors.Options{
		AllowedOrigins:     configs.Server.AllowedOrigins,
		AllowCredentials:   true,
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Authorization", "Content-Type"},
		OptionsPassthrough: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})
	jwtHandler := jwtAuth.NewHandler(issuer, issuer, router) // Issuer also become an audience in usrmgmt. Since it is the one issues tokens.
	authProviderHandler := authprv.NewAuthProvider(jwtHandler)

	handler := c.Handler(authProviderHandler)

	// test api
	router.GET("/usrmgmt/test", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("You are in usrmgmt/test page...!"))
	})

	router.GET("/usrmgmt/all/friends", srv.ListFriendInfo)
	router.GET("/usrmgmt/friends/:user_id/count", srv.GetFriendsCount)
	router.GET("/usrmgmt/set-token/:user_id", srv.SetAuthToken)

	router.GET("/usrmgmt/indexnos/:index_no", srv.GetUserKeyByIndexNo)
	router.GET("/usrmgmt/users/:user_id", srv.GetProfile)

	router.GET("/usrmgmt/friend_req", srv.ListFriendReqs)
	router.POST("/usrmgmt/friend_req", srv.CreateNewFriendReq)
	router.POST("/usrmgmt/friend_req/:friend_req_id", srv.CreateFriendLink)
	router.GET("/usrmgmt/friend_req/count", srv.GetAllFriendReqsCount)

	// TODO: Need to complete the login handler function
	router.POST("/usrmgmt/login", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Write([]byte("your details was captured. logged in"))
	})

	router.POST(authprv.AccessTokenPath, srv.GetAccessToken)
	router.POST(authprv.AccountCreationLinkPath, srv.EmailAccountCreationLink)
	router.POST(authprv.AccCreationLinkValidatePath, srv.ValidateLinkCreationParams)
	router.POST(authprv.AccCreatePath, srv.CreateUserAccount)

	router.PUT("/usrmgmt/profile/edit", srv.EditBasicProfileInfo)
	router.PUT("/usrmgmt/profile/password", srv.ResetPassword)

	router.DELETE("/usrmgmt/friend_req/:friend_req_id", srv.RemovePendingFriendReq)
	router.DELETE("/usrmgmt/friend/:friend_id", srv.RemoveFriendship)
	router.DELETE("/usrmgmt/profile", srv.DeleteUser)

	log.Println("usrmgmt_server - Listen 8000.....")
	log.Fatal(http.ListenAndServe(":8000", handler))
}
