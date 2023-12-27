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
	frnd "github.com/NamalSanjaya/nexster/pkgs/models/friend"
	freq "github.com/NamalSanjaya/nexster/pkgs/models/friend_request"
	rp "github.com/NamalSanjaya/nexster/search/pkg/repository"
	spsrv "github.com/NamalSanjaya/nexster/search/pkg/server"
	socigr "github.com/NamalSanjaya/nexster/search/pkg/social_graph"
)

type Configs struct {
	ArgDbCfg         argdb.Config        `yaml:"arangodb"`
	ContentClientCfg cl.HttpClientConfig `yaml:"content"`
}

const issuer string = "usrmgmt"
const asAud string = "search"

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
	logger := lg.New("Search")
	logger.EnableColor()

	router := httprouter.New()

	// arango db client
	argdbClient := argdb.NewDbClient(ctx, &configs.ArgDbCfg)

	// client per collection
	argFrndReqClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, freq.FriendReqColl)
	argFriendClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, frnd.FriendColl)

	// API clients
	contentApiClient := contapi.NewApiClient(&configs.ContentClientCfg)

	// repo
	repo := rp.NewRepo(argdbClient)
	frReqCtrler := freq.NewCtrler(argFrndReqClient)
	frndCtrler := frnd.NewCtrler(argFriendClient)

	sociGrphCtrler := socigr.NewGraph(contentApiClient, repo, frReqCtrler, frndCtrler)
	srv := spsrv.New(sociGrphCtrler, logger)

	router.GET("/search/test", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("You just called to search/test endpoint...!"))
	})

	router.GET("/search/users", srv.SearchForUser)

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
	log.Println("search_engine_server - Listen 8004.....")
	log.Fatal(http.ListenAndServe(":8004", handler))
}
