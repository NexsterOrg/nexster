package main

import (
	"context"
	"fmt"
	"log"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
	grepo "github.com/NamalSanjaya/nexster/timeline/pkg/repo/graph"
)

func main() {
	ctx := context.Background()
	argdbCfg := &argdb.Config{
		Hostname: "---",
		Database: "--",
		Username: "--",
		Password: "--",
		Port:     8529,
	}
	argdbClient := argdb.NewDbClient(ctx, argdbCfg)
	grpController := grepo.NewRepo(argdbClient)

	posts, err := grpController.GetPostsForTimeline(ctx, "users/482191", "2023-04-29T09:31:00.000Z", 3)
	if err != nil {
		log.Fatal(err)
	}

	for _, post := range posts {
		fmt.Printf("%+v\n", *post)
	}
	fmt.Println("--done--")
}
