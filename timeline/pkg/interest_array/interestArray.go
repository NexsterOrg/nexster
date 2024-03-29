package interestarray

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	intrs "github.com/NamalSanjaya/nexster/pkgs/models/interests"
	gr "github.com/NamalSanjaya/nexster/timeline/pkg/repository/graph_repo"
	ig "github.com/NamalSanjaya/nexster/timeline/pkg/repository/interest_group"
	sv "github.com/NamalSanjaya/nexster/timeline/pkg/repository/stem_video"
	typs "github.com/NamalSanjaya/nexster/timeline/pkg/types"
)

type interestArrayCmd struct {
	stemVideo       sv.Interface
	interestGroup   ig.Interface
	randGen         *rand.Rand
	grphrepo        gr.Interface
	interestsCtrler intrs.Interface
}

var _ Interface = (*interestArrayCmd)(nil)

func New(stemVideoIntfce sv.Interface, interestGroupIntfce ig.Interface, graphRepoIntfce gr.Interface, interestIntfce intrs.Interface) *interestArrayCmd {
	return &interestArrayCmd{
		stemVideo:       stemVideoIntfce,
		interestGroup:   interestGroupIntfce,
		randGen:         rand.New(rand.NewSource(time.Now().UnixNano())),
		grphrepo:        graphRepoIntfce,
		interestsCtrler: interestIntfce,
	}
}

// If nextPg = -1 then all data is consumed.
func (iac *interestArrayCmd) ListVideoIdsForFeed(ctx context.Context, userKey string, curPage, offset, limit int) (videos []*typs.StemVideoResp, count, nextPg int, err error) {
	videos = []*typs.StemVideoResp{}
	nextPg = curPage

	isExist, err := iac.stemVideo.IsUserVideoFeedExist(ctx, userKey)
	if err != nil {
		return
	}

	endIndex := offset + limit - 1
	if offset < 0 || endIndex < 0 {
		err = fmt.Errorf("invalid inputs")
		return
	}

	videoIds := []string{}
	if isExist {
		videoIds, err = iac.stemVideo.ListVideoIdsForUser(ctx, userKey, offset, endIndex)
		if err != nil {
			return
		}
		if len(videoIds) == 0 {
			// all pages consumed. page reset
			nextPg = -1
			return
		}
		nextPg = curPage + 1
	} else {
		// Build the interest array
		ytVideos, err2 := iac.interestsCtrler.ListVideosForInterest(ctx, userKey)
		if err2 != nil {
			err = err2
			return
		}

		iac.fisherYatesShuffle(ytVideos)

		for _, ytVideo := range ytVideos {
			if err2 = iac.stemVideo.StoreVideo(ctx, ytVideo.VId, ytVideo.Title, ytVideo.PubAt); err2 != nil {
				continue
			}
			if count < limit {
				videos = append(videos, &typs.StemVideoResp{
					Id:          ytVideo.VId,
					Title:       ytVideo.Title,
					PublishedAt: ytVideo.PubAt,
					Type:        stemVideoType,
				})
				count++
			}
			videoIds = append(videoIds, ytVideo.VId)
		}

		if count == 0 {
			nextPg = -1
			return
		}
		// cache video Ids for the user feed
		err = iac.stemVideo.StoreVideoIdsForUserFeed(ctx, userKey, videoIds)
		if err != nil {
			log.Printf("failed to cache video Ids for userkey %s: %v", userKey, err)
		}
		nextPg = 2
		err = nil
		return
	}

	for _, vId := range videoIds {
		stemVideo, err := iac.stemVideo.GetContent(ctx, vId)
		if err != nil {
			log.Println(err)
			continue
		}
		videos = append(videos, &typs.StemVideoResp{
			Id:          vId,
			Title:       stemVideo.Title,
			PublishedAt: stemVideo.PublishedAt,
			Type:        stemVideoType,
		})
		count++
	}
	err = nil
	return
}

// Perform Fisher-Yates shuffle algorithm
func (iac *interestArrayCmd) fisherYatesShuffle(slice []*intrs.YoutubeVideo) {
	for i := len(slice) - 1; i > 0; i-- {
		j := iac.randGen.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}
