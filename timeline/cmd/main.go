package main

import (
	"context"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	mrepo "github.com/NamalSanjaya/nexster/pkgs/models/media"
	rrepo "github.com/NamalSanjaya/nexster/pkgs/models/reaction"
	urepo "github.com/NamalSanjaya/nexster/pkgs/models/user"
	tsrv "github.com/NamalSanjaya/nexster/timeline/pkg/server"
	socigr "github.com/NamalSanjaya/nexster/timeline/pkg/social_graph"
)

func main() {
	ctx := context.Background()
	argdbCfg := &argdb.Config{
		Hostname: "--",
		Database: "--",
		Username: "--",
		Password: "--",
		Port:     8529,
	}
	logger := lg.New("Timeline")
	logger.EnableColor()

	router := httprouter.New()
	argRactCollClient := argdb.NewCollClient(ctx, argdbCfg, rrepo.ReactionColl)
	argMedCollClient := argdb.NewCollClient(ctx, argdbCfg, mrepo.MediaColl)
	argUsrCollClient := argdb.NewCollClient(ctx, argdbCfg, urepo.UsersColl)

	mediaRepo := mrepo.NewRepo(argMedCollClient)
	userRepo := urepo.NewCtrler(argUsrCollClient)
	reactRepo := rrepo.NewRepo(argRactCollClient)

	sociGrphCtrler := socigr.NewRepo(mediaRepo, userRepo, reactRepo)
	srv := tsrv.New(sociGrphCtrler, logger)

	router.GET("/timeline/recent_posts/:userid", srv.ListRecentPostsForTimeline) // posts for public timeline
	router.GET("/timeline/my_posts/:userid", srv.ListPostsForOwnersTimeline)     // posts for private/owners timeline
	router.GET("/timeline/friend_sugs", srv.ListFriendSuggestionsForTimeline)

	router.PUT("/timeline/reactions", srv.UpdateMediaReactions)

	log.Println("Listen....8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
