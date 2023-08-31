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
	fcrepo "github.com/NamalSanjaya/nexster/pkgs/models/faculty"
	frnd "github.com/NamalSanjaya/nexster/pkgs/models/friend"
	freq "github.com/NamalSanjaya/nexster/pkgs/models/friend_request"
	mrepo "github.com/NamalSanjaya/nexster/pkgs/models/media"
	rrepo "github.com/NamalSanjaya/nexster/pkgs/models/reaction"
	urepo "github.com/NamalSanjaya/nexster/pkgs/models/user"
	tsrv "github.com/NamalSanjaya/nexster/timeline/pkg/server"
	socigr "github.com/NamalSanjaya/nexster/timeline/pkg/social_graph"
)

const issuer string = "usrmgmt"
const asAud string = "timeline"

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
	argFacCollClient := argdb.NewCollClient(ctx, argdbCfg, fcrepo.FacultyColl)
	argFrndReqClient := argdb.NewCollClient(ctx, argdbCfg, freq.FriendReqColl)
	argFriendClient := argdb.NewCollClient(ctx, argdbCfg, frnd.FriendColl)

	mediaRepo := mrepo.NewRepo(argMedCollClient)
	userRepo := urepo.NewCtrler(argUsrCollClient)
	reactRepo := rrepo.NewRepo(argRactCollClient)
	facRepo := fcrepo.NewCtrler(argFacCollClient)
	frReqCtrler := freq.NewCtrler(argFrndReqClient)
	frndCtrler := frnd.NewCtrler(argFriendClient)

	sociGrphCtrler := socigr.NewRepo(mediaRepo, userRepo, reactRepo)
	srv := tsrv.New(sociGrphCtrler, logger)

	router.GET("/timeline/recent_posts/:userid", srv.ListRecentPostsForTimeline) // posts for public timeline
	router.GET("/timeline/my_posts/:userid", srv.ListPostsForOwnersTimeline)     // posts for private/owners timeline

	router.GET("/timeline/friend_sugs/v2/:faculty", srv.ListFriendSuggestionsV2)
	router.GET("/timeline/friend_sugs", srv.ListFriendSuggestions)

	router.GET("/timeline/media", srv.ListOwnersViewMedia)
	router.GET("/timeline/media/:user_id", srv.ListPublicMedia)
	router.GET("/timeline/r/media/:img_owner_id", srv.ListRoleBasedMedia) // "/r/*" --> for dynamic role based paths

	router.PUT("/timeline/reactions/:reaction_id", srv.UpdateMediaReactions)

	router.POST("/timeline/reactions", srv.CreateMediaReactions) // Create new reaction link

	c := cors.New(cors.Options{
		AllowedOrigins:     []string{"http://localhost:3000", "http://192.168.1.101:3000"},
		AllowCredentials:   true,
		AllowedMethods:     []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:     []string{"Authorization", "Content-Type"},
		OptionsPassthrough: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})

	jwtHandler := jwtAuth.NewHandler(issuer, asAud, router)
	handler := c.Handler(jwtHandler)
	log.Println("timeline_server - Listen 8001.....")
	log.Fatal(http.ListenAndServe(":8001", handler))
}
