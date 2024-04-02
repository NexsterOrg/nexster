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
	"github.com/NamalSanjaya/nexster/pkgs/cache/redis"
	cl "github.com/NamalSanjaya/nexster/pkgs/client"
	contapi "github.com/NamalSanjaya/nexster/pkgs/client/content_api"
	fcrepo "github.com/NamalSanjaya/nexster/pkgs/models/faculty"
	frnd "github.com/NamalSanjaya/nexster/pkgs/models/friend"
	freq "github.com/NamalSanjaya/nexster/pkgs/models/friend_request"
	intrs "github.com/NamalSanjaya/nexster/pkgs/models/interests"
	intrsIn "github.com/NamalSanjaya/nexster/pkgs/models/interestsIn"
	mrepo "github.com/NamalSanjaya/nexster/pkgs/models/media"
	mo "github.com/NamalSanjaya/nexster/pkgs/models/media_owner"
	rrepo "github.com/NamalSanjaya/nexster/pkgs/models/reaction"
	urepo "github.com/NamalSanjaya/nexster/pkgs/models/user"
	concr "github.com/NamalSanjaya/nexster/pkgs/utill/concurrency"
	ustr "github.com/NamalSanjaya/nexster/pkgs/utill/string"
	ytapi "github.com/NamalSanjaya/nexster/timeline/pkg/client/youtube_api"
	ia "github.com/NamalSanjaya/nexster/timeline/pkg/interest_array"
	grrepo "github.com/NamalSanjaya/nexster/timeline/pkg/repository/graph_repo"
	ig "github.com/NamalSanjaya/nexster/timeline/pkg/repository/interest_group"
	sv "github.com/NamalSanjaya/nexster/timeline/pkg/repository/stem_video"
	tsrv "github.com/NamalSanjaya/nexster/timeline/pkg/server"
	socigr "github.com/NamalSanjaya/nexster/timeline/pkg/social_graph"
)

type Configs struct {
	Server           tsrv.ServerConfig   `yaml:"server"`
	ArgDbCfg         argdb.Config        `yaml:"arangodb"`
	ContentClientCfg cl.HttpClientConfig `yaml:"content"`
	RedisCfg         redis.Config        `yaml:"redis"`
	StemVideoFeed    sv.StemVideoConfig  `yaml:"stemVideoFeed"`
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

	// redis cache instances
	redisClient := redis.NewClient(ctx, &configs.RedisCfg)
	interestGroupRepo := ig.New(redisClient)
	stemVideoRepo := sv.New(&configs.StemVideoFeed, redisClient)

	// arango direct db client
	argdbClient := argdb.NewDbClient(ctx, &configs.ArgDbCfg)

	// arango db collection clients
	argRactCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, rrepo.ReactionColl)
	argMedCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, mrepo.MediaColl)
	argUsrCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, urepo.UsersColl)
	argFacCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, fcrepo.FacultyColl)
	argFrndReqClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, freq.FriendReqColl)
	argFriendClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, frnd.FriendColl)
	argMediaOwnerClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, mo.MediaOwnerColl)
	argInteretsClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, intrs.InterestsColl)
	argInteretsInClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, intrsIn.InterestsInColl)

	mediaRepo := mrepo.NewRepo(argMedCollClient)
	userRepo := urepo.NewCtrler(argUsrCollClient)
	reactRepo := rrepo.NewRepo(argRactCollClient)
	facRepo := fcrepo.NewCtrler(argFacCollClient)
	frReqCtrler := freq.NewCtrler(argFrndReqClient)
	frndCtrler := frnd.NewCtrler(argFriendClient)
	mdOwnerCtrler := mo.NewCtrler(argMediaOwnerClient)
	interestsCtrler := intrs.NewCtrler(argInteretsClient)
	interestsInCtrler := intrsIn.NewCtrler(argInteretsInClient)

	// graph repository
	grphRepo := grrepo.NewRepo(argdbClient)

	// interest array repository
	interestArrCmder := ia.New(stemVideoRepo, interestGroupRepo, grphRepo, interestsCtrler)

	// API clients
	contentApiClient := contapi.NewApiClient(&configs.ContentClientCfg)

	// assign youtube clients
	ytClients := []*ytapi.YoutubeApi{}
	for _, apiKey := range configs.Server.APIKeys {
		ytClients = append(ytClients, ytapi.NewClient(ctx, apiKey))

	}

	sociGrphCtrler := socigr.NewRepo(mediaRepo, userRepo, reactRepo, facRepo, frReqCtrler, frndCtrler, mdOwnerCtrler, contentApiClient, interestsCtrler, interestsInCtrler)
	srv := tsrv.New(&configs.Server, sociGrphCtrler, logger, interestArrCmder, ytClients)

	// Schdule YoutubeFetcher
	// TODO: Since we are shutting down the servers, recurring won't work properly. [HIGH]
	go concr.SchduleRecurringTaskInDays(ctx, "Youtube Fetcher", configs.Server.YoutubeFetcherRecurringInDays, func() {
		srv.YoutubeAPIFetcher(ctx)
	})

	// Only to create interestsIn edges for existing users. Once it is done, this need to be removed.
	srv.CreateInterestEdgesForExistingUsers(ctx)

	router.GET("/timeline/stem/videos", srv.VideoFeedForTimeline)
	router.GET("/timeline/posts/anytype", srv.ListAllTypePostForTimeline)
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
