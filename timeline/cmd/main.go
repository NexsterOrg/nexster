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
	fcrepo "github.com/NamalSanjaya/nexster/pkgs/models/faculty"
	frnd "github.com/NamalSanjaya/nexster/pkgs/models/friend"
	freq "github.com/NamalSanjaya/nexster/pkgs/models/friend_request"
	mrepo "github.com/NamalSanjaya/nexster/pkgs/models/media"
	mo "github.com/NamalSanjaya/nexster/pkgs/models/media_owner"
	rrepo "github.com/NamalSanjaya/nexster/pkgs/models/reaction"
	urepo "github.com/NamalSanjaya/nexster/pkgs/models/user"
	ustr "github.com/NamalSanjaya/nexster/pkgs/utill/string"
	tsrv "github.com/NamalSanjaya/nexster/timeline/pkg/server"
	socigr "github.com/NamalSanjaya/nexster/timeline/pkg/social_graph"
)

type Configs struct {
	Server           tsrv.ServerConfig   `yaml:"server"`
	ArgDbCfg         argdb.Config        `yaml:"arangodb"`
	ContentClientCfg cl.HttpClientConfig `yaml:"content"`
}

const issuer string = "usrmgmt"
const asAud string = "timeline"

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
	logger := lg.New("Timeline")
	logger.EnableColor()

	router := httprouter.New()

	// arango db collection clients
	argRactCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, rrepo.ReactionColl)
	argMedCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, mrepo.MediaColl)
	argUsrCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, urepo.UsersColl)
	argFacCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, fcrepo.FacultyColl)
	argFrndReqClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, freq.FriendReqColl)
	argFriendClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, frnd.FriendColl)
	argMediaOwnerClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, mo.MediaOwnerColl)

	mediaRepo := mrepo.NewRepo(argMedCollClient)
	userRepo := urepo.NewCtrler(argUsrCollClient)
	reactRepo := rrepo.NewRepo(argRactCollClient)
	facRepo := fcrepo.NewCtrler(argFacCollClient)
	frReqCtrler := freq.NewCtrler(argFrndReqClient)
	frndCtrler := frnd.NewCtrler(argFriendClient)
	mdOwnerCtrler := mo.NewCtrler(argMediaOwnerClient)

	// API clients
	contentApiClient := contapi.NewApiClient(&configs.ContentClientCfg)

	sociGrphCtrler := socigr.NewRepo(mediaRepo, userRepo, reactRepo, facRepo, frReqCtrler, frndCtrler, mdOwnerCtrler, contentApiClient)
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
	router.POST("/timeline/posts/image", srv.CreateImagePost)    // Create new post

	router.DELETE("/timeline/posts/image/:mediaKey", srv.DeleteImagePost) // Delete post

	c := cors.New(cors.Options{
		AllowedOrigins:     configs.Server.AllowedOrigins,
		AllowCredentials:   true,
		AllowedMethods:     []string{"GET", "POST", "PUT", "OPTIONS", "DELETE"},
		AllowedHeaders:     []string{"Authorization", "Content-Type"},
		OptionsPassthrough: true,
		Debug:              false,
	})

	jwtHandler := jwtAuth.NewHandler(issuer, asAud, router, ustr.MkCompletePath(configs.Server.ProjectDir, configs.Server.PublicKeyPath))
	handler := c.Handler(jwtHandler)
	log.Println("timeline_server - Listen 8001.....")
	log.Fatal(http.ListenAndServe(":8001", handler))
}
