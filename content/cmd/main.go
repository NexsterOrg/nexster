package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"

	"github.com/NamalSanjaya/nexster/content/pkg/client/blob"
	srv "github.com/NamalSanjaya/nexster/content/pkg/server"
	azb "github.com/NamalSanjaya/nexster/pkgs/azure/blob_storage"
)

const (
	storageAccount string = ""
	imageContainer string = ""
)

func main() {
	router := httprouter.New()
	azBlobClient := azb.NewAzBlobClient(storageAccount)
	imgBlobClient := blob.NewBlobClient(imageContainer, azBlobClient)

	csrv := srv.New(imgBlobClient)

	router.GET("/images/:imgId", csrv.ServeImages)

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
