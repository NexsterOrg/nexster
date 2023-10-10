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
	pb "github.com/NamalSanjaya/nexster/pkgs/models/posted_by"
	"github.com/NamalSanjaya/nexster/pkgs/models/user"
	spsrv "github.com/NamalSanjaya/nexster/space/pkg/server"
	socigr "github.com/NamalSanjaya/nexster/space/pkg/social_graph"
)

type Configs struct {
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
	logger := lg.New("Space")
	logger.EnableColor()

	router := httprouter.New()

	// arango db collection clients
	argEventCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, ev.EventColl)
	argPostedByCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, pb.PostedByColl)
	argUserCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, user.UsersColl)

	eventCtrler := ev.NewCtrler(argEventCollClient)
	postedByCtrler := pb.NewCtrler(argPostedByCollClient)
	userCtrler := user.NewCtrler(argUserCollClient)

	// API clients
	contentApiClient := contapi.NewApiClient(&configs.ContentClientCfg)

	sociGrphCtrler := socigr.NewGraph(eventCtrler, postedByCtrler, userCtrler, contentApiClient)
	srv := spsrv.New(sociGrphCtrler, logger)

	router.GET("/space/events", srv.ListLatestEventsFromSpace)

	router.POST("/space/events", srv.CreateEventInSpace)

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
	log.Println("space_server - Listen 8003.....")
	log.Fatal(http.ListenAndServe(":8003", handler))
}
