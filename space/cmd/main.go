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
	ev "github.com/NamalSanjaya/nexster/pkgs/models/event"
	erec "github.com/NamalSanjaya/nexster/pkgs/models/event_reaction"
	pb "github.com/NamalSanjaya/nexster/pkgs/models/posted_by"
	"github.com/NamalSanjaya/nexster/pkgs/models/user"
	rp "github.com/NamalSanjaya/nexster/space/pkg/repository"
	spsrv "github.com/NamalSanjaya/nexster/space/pkg/server"
	socigr "github.com/NamalSanjaya/nexster/space/pkg/social_graph"
)

type Configs struct {
	ArgDbCfg         argdb.Config        `yaml:"arangodb"`
	ContentClientCfg cl.HttpClientConfig `yaml:"content"`
}

const issuer string = "usrmgmt"
const asAud string = "space"

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
	logger := lg.New("Space")
	logger.EnableColor()

	router := httprouter.New()

	// arango db client
	argdbClient := argdb.NewDbClient(ctx, &configs.ArgDbCfg)

	// arango db collection clients
	argEventCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, ev.EventColl)
	argPostedByCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, pb.PostedByColl)
	argUserCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, user.UsersColl)
	argEventReactCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, erec.EventReactionColl)

	eventCtrler := ev.NewCtrler(argEventCollClient)
	postedByCtrler := pb.NewCtrler(argPostedByCollClient)
	userCtrler := user.NewCtrler(argUserCollClient)
	eventReactCtrler := erec.NewCtrler(argEventReactCollClient)

	// API clients
	contentApiClient := contapi.NewApiClient(&configs.ContentClientCfg)

	// repo
	repo := rp.NewRepo(argdbClient)

	sociGrphCtrler := socigr.NewGraph(eventCtrler, postedByCtrler, userCtrler, eventReactCtrler, contentApiClient, repo)
	srv := spsrv.New(sociGrphCtrler, logger)

	router.GET("/space/test", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("You just called to space/test endpoint...!"))
	})

	router.GET("/space/events/:eventKey/:reactType", srv.ListLoveReactUsersForEvent)
	router.GET("/space/events/:eventKey", srv.GetEventFromSpace)
	router.GET("/space/events", srv.ListUpcomingEventsFromSpace)
	router.GET("/space/my/events", srv.ListMyEventsFromSpace)

	router.POST("/space/events/:eventKey/reaction", srv.CreateEventReaction)
	router.POST("/space/events", srv.CreateEventInSpace)

	router.PUT("/space/events/reactions/:reactionKey/:reactionType/:state", srv.SetEventReactionState)

	router.DELETE("/space/events/:eventKey", srv.DeleteEventFromSpace)

	c := cors.New(cors.Options{
		AllowedOrigins:     []string{"http://localhost:3000", "https://exp-mora-nexster.azurewebsites.net"},
		AllowCredentials:   true,
		AllowedMethods:     []string{"GET", "POST", "PUT", "OPTIONS", "DELETE"},
		AllowedHeaders:     []string{"Authorization", "Content-Type"},
		OptionsPassthrough: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})

	jwtHandler := jwtAuth.NewHandler(issuer, asAud, router)
	handler := c.Handler(jwtHandler)
	log.Println("space_server - Listen 8003.....")
	log.Fatal(http.ListenAndServe(":8003", handler))
}
