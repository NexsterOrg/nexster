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

	"github.com/NamalSanjaya/nexster/content/pkg/client/blob"
	avtrrepo "github.com/NamalSanjaya/nexster/content/pkg/repository/avatar"
	srv "github.com/NamalSanjaya/nexster/content/pkg/server"
	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	azb "github.com/NamalSanjaya/nexster/pkgs/azure/blob_storage"
	avtr "github.com/NamalSanjaya/nexster/pkgs/models/avatar"
)

type Configs struct {
	Server       srv.ServerConfig        `yaml:"server"`
	AzBlobClient azb.AzBlobClientConfigs `yaml:"azure"`
	ArgDbCfg     argdb.Config            `yaml:"arangodb"`
}

func main() {
	yamlFile, err := os.ReadFile("../configs/config.yaml")
	if err != nil {
		log.Panicf("Error reading YAML file: %v", err)
	}
	var configs Configs
	if err := yaml.Unmarshal(yamlFile, &configs); err != nil {
		log.Panicf("Error unmarshaling YAML: %v", err)
	}
	azBlobClient := azb.NewAzBlobClient(&configs.AzBlobClient)
	imgBlobClient := blob.NewBlobClient(configs.AzBlobClient.Containers["images"], azBlobClient)

	ctx := context.Background()
	argAvatarCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, avtr.AvatarColl)
	avatarCtrler := avtr.NewCtrler(argAvatarCollClient)
	avatarRepo := avtrrepo.NewRepo(avatarCtrler)

	logger := lg.New("Content")
	logger.EnableColor()

	csrv := srv.New(&configs.Server, logger, imgBlobClient, avatarRepo)

	router := httprouter.New()
	router.GET("/content/hmac/image/:namespace/:imgId", csrv.CreateImgUrl)
	router.GET("/content/images/:namespace/:imgId", csrv.ServeImages) // this path is use to create the image by CreateImgUrl

	c := cors.New(cors.Options{
		AllowedOrigins:     []string{"http://localhost:3000", "http://192.168.1.101:3000"},
		AllowCredentials:   true,
		AllowedMethods:     []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:     []string{"Authorization", "Content-Type"},
		OptionsPassthrough: true,
		Debug:              false,
	})
	handler := c.Handler(router)
	log.Println("content_server - Listen 8002.....")
	log.Fatal(http.ListenAndServe(":8002", handler))
}
