package main

import (
	"context"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"
	"github.com/rs/cors"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	jwtAuth "github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
	frnd "github.com/NamalSanjaya/nexster/pkgs/models/friend"
	freq "github.com/NamalSanjaya/nexster/pkgs/models/friend_request"
	usr "github.com/NamalSanjaya/nexster/pkgs/models/user"
	authprv "github.com/NamalSanjaya/nexster/usrmgmt/pkg/auth_provider"
	usrv "github.com/NamalSanjaya/nexster/usrmgmt/pkg/server"
	socigr "github.com/NamalSanjaya/nexster/usrmgmt/pkg/social_graph"
)

const issuer string = "usrmgmt"

func main() {
	ctx := context.Background()
	argdbCfg := &argdb.Config{
		Hostname: "",
		Database: "",
		Username: "",
		Password: "",
		Port:     8529,
	}
	logger := lg.New("UserMgmtSrv")
	logger.EnableColor()

	argFrndReqClient := argdb.NewCollClient(ctx, argdbCfg, freq.FriendReqColl)
	argFrndClient := argdb.NewCollClient(ctx, argdbCfg, frnd.FriendColl)
	argUsrClient := argdb.NewCollClient(ctx, argdbCfg, usr.UsersColl)

	frReqCtrler := freq.NewCtrler(argFrndReqClient)
	frndCtrler := frnd.NewCtrler(argFrndClient)
	usrCtrler := usr.NewCtrler(argUsrClient)

	grCtrler := socigr.NewGrphCtrler(frReqCtrler, frndCtrler, usrCtrler)
	srv := usrv.New(grCtrler, logger)

	router := httprouter.New()

	c := cors.New(cors.Options{
		AllowedOrigins:     []string{"http://localhost:3000", "http://192.168.1.101:3000"},
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

	router.GET("/usrmgmt/friends/:user_id", srv.ListFriendInfo)
	router.GET("/usrmgmt/set-token/:user_id", srv.SetAuthToken)

	router.GET("/usrmgmt/indexnos/:index_no", srv.GetUserKeyByIndexNo)
	router.GET("/usrmgmt/users/:user_id", srv.GetProfile)

	router.GET("/usrmgmt/friends/:user_id/count", srv.GetFriendsCount)

	router.GET("/usrmgmt/friend_req", srv.ListFriendReqs)
	router.POST("/usrmgmt/friend_req", srv.CreateNewFriendReq)
	router.POST("/usrmgmt/friend_req/:friend_req_id", srv.CreateFriendLink)
	router.GET("/usrmgmt/friend_req/count", srv.GetAllFriendReqsCount)

	// TODO: Need to complete the login handler function
	router.POST("/usrmgmt/login", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Write([]byte("your details was captured. logged in"))
	})

	router.DELETE("/usrmgmt/friend_req/:friend_req_id", srv.RemovePendingFriendReq)
	router.DELETE("/usrmgmt/friend/:friend_id", srv.RemoveFriendship)

	log.Println("usrmgmt_server - Listen 8000.....")
	log.Fatal(http.ListenAndServe(":8000", handler))
}
