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
	frReqCtrler := freq.NewCtrler(argFrndReqClient)
	frndCtrler := frnd.NewCtrler(argFrndClient)

	grCtrler := socigr.NewGrphCtrler(frReqCtrler, frndCtrler)
	srv := usrv.New(grCtrler, logger)

	router := httprouter.New()

	router.POST("/usrmgmt/friend_req", srv.HandleFriendReq)
	router.POST("/usrmgmt/friend_req/:friend_req_id", srv.CreateFriendLink)

	router.DELETE("/usrmgmt/friend_req/:friend_req_id", srv.RemovePendingFriendReq)

	log.Println("Listen....8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
