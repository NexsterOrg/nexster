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

	authprv "github.com/NamalSanjaya/nexster/boarding_finder/pkg/auth_provider"
	"github.com/NamalSanjaya/nexster/boarding_finder/pkg/rbac"
	rp "github.com/NamalSanjaya/nexster/boarding_finder/pkg/repository"
	bdfsrv "github.com/NamalSanjaya/nexster/boarding_finder/pkg/server"
	socigr "github.com/NamalSanjaya/nexster/boarding_finder/pkg/social_graph"
	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	jwtAuth "github.com/NamalSanjaya/nexster/pkgs/auth/jwt"
	cl "github.com/NamalSanjaya/nexster/pkgs/client"
	contapi "github.com/NamalSanjaya/nexster/pkgs/client/content_api"
	bao "github.com/NamalSanjaya/nexster/pkgs/models/boardingAdOwned"
	bdo "github.com/NamalSanjaya/nexster/pkgs/models/boardingOwner"
	bad "github.com/NamalSanjaya/nexster/pkgs/models/boarding_ads"
	ustr "github.com/NamalSanjaya/nexster/pkgs/utill/string"
)

type Configs struct {
	Server           bdfsrv.ServerConfig `yaml:"server"`
	ArgDbCfg         argdb.Config        `yaml:"arangodb"`
	ContentClientCfg cl.HttpClientConfig `yaml:"content"`
}

const issuer string = "usrmgmt"
const asAud string = "bdFinder"

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
	logger := lg.New("bdFinder")
	logger.EnableColor()

	router := httprouter.New()

	// arango db client
	argdbClient := argdb.NewDbClient(ctx, &configs.ArgDbCfg)

	// arango db collection clients
	argBdAdClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, bad.BdAdsColl)
	argBdAdOwnedClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, bao.BdAdOwnedColl)
	argBdOwnerClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, bdo.BdOwnerColl)

	bdAdCtrler := bad.NewCtrler(argBdAdClient)
	bdAdOwnedCtrler := bao.NewCtrler(argBdAdOwnedClient)
	bdOwnerCtrler := bdo.NewCtrler(argBdOwnerClient)

	// repo
	repo := rp.NewRepo(argdbClient)

	// API clients
	contentApiClient := contapi.NewApiClient(&configs.ContentClientCfg)

	sociGrphCtrler := socigr.NewGraph(bdAdCtrler, bdAdOwnedCtrler, bdOwnerCtrler, repo, contentApiClient)
	srv := bdfsrv.New(sociGrphCtrler, logger, rbac.NewRbacGuard())

	router.GET("/bdfinder/test", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("You just called to boarding-finder/test endpoint...!"))
	})

	router.GET("/bdfinder/ads/:adKey", srv.GetAdForMainView)
	router.GET("/bdfinder/ads", srv.ListAdsForMainView)

	router.POST("/bdfinder/ads", srv.CreateAd)

	router.PUT("/bdfinder/ads/:adKey/status", srv.ChangeStatusOfAd)

	// non-protect paths
	router.POST(authprv.BdOwnerAccCreatePath, srv.CreateBoardingOwner)

	c := cors.New(cors.Options{
		AllowedOrigins:     configs.Server.AllowedOrigins,
		AllowCredentials:   true,
		AllowedMethods:     []string{"GET", "POST", "PUT", "OPTIONS", "DELETE"},
		AllowedHeaders:     []string{"Authorization", "Content-Type"},
		OptionsPassthrough: true,
		Debug:              false,
	})

	jwtHandler := jwtAuth.NewHandler(issuer, asAud, router, ustr.MkCompletePath(configs.Server.ProjectDir, configs.Server.PublicKeyPath))
	authProviderHandler := authprv.NewAuthProvider(jwtHandler)
	handler := c.Handler(authProviderHandler)
	log.Println("bd-finder_server - Listen 8005.....")
	log.Fatal(http.ListenAndServe(":8005", handler))
}
