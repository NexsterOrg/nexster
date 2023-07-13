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
	argdbClient := argdb.NewDbClient(ctx, argdbCfg)
	argCollClient := argdb.NewCollClient(ctx, argdbCfg, rrepo.ReactionColl)

	mediaRepo := mrepo.NewRepo(argdbClient)
	userRepo := urepo.NewRepo(argdbClient)
	reactRepo := rrepo.NewRepo(argCollClient)

	sociGrphCtrler := socigr.NewRepo(mediaRepo, userRepo, reactRepo)
	srv := tsrv.New(sociGrphCtrler, logger)

	router.GET("/timeline/recent_posts", srv.ListRecentPostsForTimeline)
	router.GET("/timeline/friend_sugs", srv.ListFriendSuggestionsForTimeline)

	router.PUT("/timeline/reactions", srv.UpdateMediaReactions)

	log.Println("Listen....8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
