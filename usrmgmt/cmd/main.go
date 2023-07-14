package main

import (
	"context"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	frnd "github.com/NamalSanjaya/nexster/pkgs/models/friend"
	freq "github.com/NamalSanjaya/nexster/pkgs/models/friend_request"
	usr "github.com/NamalSanjaya/nexster/pkgs/models/user"
	usrv "github.com/NamalSanjaya/nexster/usrmgmt/pkg/server"
	socigr "github.com/NamalSanjaya/nexster/usrmgmt/pkg/social_graph"
)

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

	router.GET("/usrmgmt/friends/:user_id", srv.ListFriendInfo)

	router.POST("/usrmgmt/friend_req", srv.HandleFriendReq)
	router.POST("/usrmgmt/friend_req/:friend_req_id", srv.CreateFriendLink)

	router.DELETE("/usrmgmt/friend_req/:friend_req_id", srv.RemovePendingFriendReq)
	router.DELETE("/usrmgmt/friend/:friend_id", srv.RemoveFriendship)

	log.Println("Listen....8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
