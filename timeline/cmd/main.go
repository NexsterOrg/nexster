package main

import (
	"context"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
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
	logger := lg.New("nexster-timeline")
	logger.EnableColor()

	router := httprouter.New()
	argdbClient := argdb.NewDbClient(ctx, argdbCfg)
	sociGrphCtrler := socigr.NewRepo(argdbClient)
	srv := tsrv.New(sociGrphCtrler, logger)

	router.GET("/timeline/fetch_posts", srv.ListRecentPostsForTimeline)

	log.Println("Listen....8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
