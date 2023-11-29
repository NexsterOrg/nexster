package socialgraph

import (
	"context"
	"log"
	"sort"
	"strconv"
	"strings"

	contapi "github.com/NamalSanjaya/nexster/pkgs/client/content_api"
	umath "github.com/NamalSanjaya/nexster/pkgs/utill/math"
	rp "github.com/NamalSanjaya/nexster/search/pkg/repository"
	"github.com/NamalSanjaya/nexster/search/utill/algo"
)

const minAcceptanceIndex float32 = float32(0.25)
const maxSearchResults int = 20

type socialGraph struct {
	contentClient contapi.Interface
	repo          rp.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewGraph(contentClientIntfce contapi.Interface, repoIntface rp.Interface) *socialGraph {
	return &socialGraph{
		contentClient: contentClientIntfce,
		repo:          repoIntface,
	}
}

func (gr *socialGraph) SearchAmongUsers(ctx context.Context, keyword string) ([]*map[string]string, error) {
	users, err := gr.repo.ListAllUsers(ctx)
	if err != nil {
		return []*map[string]string{}, nil
	}

	keyword = strings.ToLower(keyword)
	seletedUsers := []*map[string]string{}
	selectedCount := 0
	for _, user := range users {
		score := algo.JaccardIndexBasedScore(strings.ToLower((*user)["username"]), keyword)
		if score < minAcceptanceIndex {
			continue
		}
		(*user)["score"] = strconv.FormatFloat(float64(score), 'f', -1, 32)
		seletedUsers = append(seletedUsers, user)
		selectedCount++
	}

	sort.Slice(seletedUsers, func(i, j int) bool {
		valI, err := strconv.ParseFloat((*seletedUsers[i])["score"], 32)
		if err != nil {
			return false
		}
		valJ, err := strconv.ParseFloat(((*seletedUsers[j])["score"]), 32)
		if err != nil {
			return false
		}
		return valI > valJ
	})

	selectedLastIndex := umath.Min[int](selectedCount, maxSearchResults)
	seletedUsers = seletedUsers[:selectedLastIndex]

	for _, user := range seletedUsers {
		imgUrl, err := gr.contentClient.CreateImageUrl((*user)["image_url"], contapi.Viewer)
		if err != nil {
			log.Println("failed at user search: failed to create avatar url: ", err)
			continue
		}
		(*user)["image_url"] = imgUrl
		delete(*user, "score")
	}
	return seletedUsers, nil
}
