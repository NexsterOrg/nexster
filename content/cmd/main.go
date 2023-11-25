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
	mdrepo "github.com/NamalSanjaya/nexster/content/pkg/repository/media"
	srv "github.com/NamalSanjaya/nexster/content/pkg/server"
	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	azb "github.com/NamalSanjaya/nexster/pkgs/azure/blob_storage"
	avtr "github.com/NamalSanjaya/nexster/pkgs/models/avatar"
	md "github.com/NamalSanjaya/nexster/pkgs/models/media"
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

	argmediaCollClient := argdb.NewCollClient(ctx, &configs.ArgDbCfg, md.MediaColl)
	mediaCtrler := md.NewRepo(argmediaCollClient)
	mediaRepo := mdrepo.NewRepository(mediaCtrler)

	logger := lg.New("Content")
	logger.EnableColor()

	csrv := srv.New(&configs.Server, logger, imgBlobClient, avatarRepo, mediaRepo)

	router := httprouter.New()
	router.GET("/content/hmac/image/:namespace/:imgId", csrv.CreateImgUrl)
	router.GET("/content/images/:namespace/:imgId", csrv.ServeImages) // this path is use to create the image by CreateImgUrl

	router.POST("/content/images/:namespace", csrv.UploadImage)

	router.PUT("/content/images/:namespace/:imgId", csrv.ReplaceImage) // TODO: Need to check

	router.DELETE("/content/images/:namespace/:imgId", csrv.DeleteImage)

	c := cors.New(cors.Options{
		AllowedOrigins:     []string{"http://localhost:3000", "http://192.168.1.101:3000"},
		AllowCredentials:   true,
		AllowedMethods:     []string{"GET", "POST", "PUT", "OPTIONS", "DELETE"},
		AllowedHeaders:     []string{"Authorization", "Content-Type"},
		OptionsPassthrough: true,
		Debug:              false,
	})
	handler := c.Handler(router)
	log.Println("content_server - Listen 8002.....")
	log.Fatal(http.ListenAndServe(":8002", handler))
}

/** TODO:
1. All routes are not protected.
2. Image serve route should be open route. (can't be protected since it need to access by browser)
3. Image URL create route, upload image, Replace image route should be protected.
issue: https://github.com/NamalSanjaya/nexster-rnd/issues/18
*/
