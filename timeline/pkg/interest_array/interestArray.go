package interestarray

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	ig "github.com/NamalSanjaya/nexster/timeline/pkg/repository/interest_group"
	sv "github.com/NamalSanjaya/nexster/timeline/pkg/repository/stem_video"
	typs "github.com/NamalSanjaya/nexster/timeline/pkg/types"
)

type interestArrayCmd struct {
	stemVideo     sv.Interface
	interestGroup ig.Interface
	randGen       *rand.Rand
}

var _ Interface = (*interestArrayCmd)(nil)

func New(stemVideoIntfce sv.Interface, interestGroupIntfce ig.Interface) *interestArrayCmd {
	return &interestArrayCmd{
		stemVideo:     stemVideoIntfce,
		interestGroup: interestGroupIntfce,
		randGen:       rand.New(rand.NewSource(time.Now().UnixNano())),
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
		fmt.Println("Need to build the interest array")
		interestGroupIds := []string{"ig1", "ig2", "ig3", "ig4"} // todo: get from the db

		for _, grpId := range interestGroupIds {
			vIdsPerGrp, err := iac.interestGroup.ListVideoIdsForGroup(ctx, grpId)
			if err != nil {
				log.Println("err building the interest group", err)
				continue
			}
			videoIds = iac.combineSlicesRandomly(videoIds, vIdsPerGrp)
		}
		ln := len(videoIds)
		if ln == 0 {
			nextPg = -1
			return
		}
		// cache video Ids for the user feed
		err = iac.stemVideo.StoreVideoIdsForUserFeed(ctx, userKey, videoIds)
		if err != nil {
			log.Printf("failed to cache video Ids for userkey %s", userKey)
		}
		videoIds = videoIds[0:limit]
		// page reset
		nextPg = 2
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
			Type:        "video",
		})
		count++
	}
	err = nil
	return
}

/*
1. List video Ids from user feed.
2. If the list is empty --> build a new one
3. If creating ---> retry
4. If Ok
   4-1. pg, pgSize get the list of video ids.
   4-2. Loop over each and create the video content.
   4-3. Return values ListOfVideos, versionNo
*/

// Combine two slices randomly
func (iac *interestArrayCmd) combineSlicesRandomly(slice1, slice2 []string) []string {
	iac.shuffle(slice1)
	iac.shuffle(slice2)

	ln1 := len(slice1)
	ln2 := len(slice2)

	// Interleave the elements of both slices
	combined := make([]string, 0, ln1+ln2)
	for i := 0; i < ln1 || i < ln2; i++ {
		if i < ln1 {
			combined = append(combined, slice1[i])
		}
		if i < ln2 {
			combined = append(combined, slice2[i])
		}
	}

	return combined
}

// Shuffle a slice using a specific random number generator
func (iac *interestArrayCmd) shuffle(slice []string) {
	for i := range slice {
		j := iac.randGen.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}
