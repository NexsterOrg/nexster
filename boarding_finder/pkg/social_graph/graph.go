package socialgraph

import (
	"context"
	"fmt"

	dtm "github.com/NamalSanjaya/nexster/boarding_finder/pkg/dto_mapper"
	rp "github.com/NamalSanjaya/nexster/boarding_finder/pkg/repository"
	er "github.com/NamalSanjaya/nexster/pkgs/errors"
	bao "github.com/NamalSanjaya/nexster/pkgs/models/boardingAdOwned"
	bdo "github.com/NamalSanjaya/nexster/pkgs/models/boardingOwner"
	bda "github.com/NamalSanjaya/nexster/pkgs/models/boarding_ads"
	pwd "github.com/NamalSanjaya/nexster/pkgs/utill/password"
)

type socialGraph struct {
	bdAdsCtrler     bda.Interface
	bdAdOwnedCtrler bao.Interface
	bdOwnerCtrler   bdo.Interface
	repo            rp.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewGraph(bdAdsIntfce bda.Interface, bdAdOwnedIntfce bao.Interface, bdOwnerIntfce bdo.Interface, repoIntfce rp.Interface) *socialGraph {
	return &socialGraph{
		bdAdsCtrler:     bdAdsIntfce,
		bdAdOwnedCtrler: bdAdOwnedIntfce,
		bdOwnerCtrler:   bdOwnerIntfce,
		repo:            repoIntfce,
	}
}

func (gr *socialGraph) CreateAd(ctx context.Context, bdOwnerKey string, data *dtm.CreateAdDto) (adNodeKey, ownedEdgeKey string, err error) {
	isExisted, err := gr.bdOwnerCtrler.Exist(ctx, bdOwnerKey)
	if err != nil {
		return
	}
	if !isExisted {
		err = er.NewNotFoundError("owner does not exist")
		return
	}
	adNodeKey, err = gr.bdAdsCtrler.Create(ctx, &bda.BoardingAds{
		Title:        data.Title,
		Description:  data.Description,
		Bills:        data.Bills,
		ImageUrls:    data.ImageUrls,
		Rent:         data.Rent,
		Address:      data.Address,
		Beds:         data.Beds,
		Baths:        data.Baths,
		Gender:       data.Gender,
		Distance:     data.Distance,
		DistanceUnit: data.DistanceUnit,
		Status:       bda.Pending,
	})
	if err != nil {
		return
	}
	ownedEdgeKey, err = gr.bdAdOwnedCtrler.CreateDocument(ctx, bda.MkBdAdsDocId(adNodeKey), bdo.MkDocId(bdOwnerKey))
	return
}

// Create a boarding owner node with pending status
func (gr *socialGraph) CreateBoardingOwner(ctx context.Context, data *dtm.CreateBoardingOwner) (bdOwnerKey string, err error) {
	exist, err := gr.repo.ExistAndUniqueForMainContact(ctx, data.MainContact)
	if err != nil {
		return
	}
	if exist {
		err = er.NewConflictError("boarding user already exists")
		return
	}
	newPasswdHash, err := pwd.HashPassword(data.Password)
	if err != nil {
		err = fmt.Errorf("failed to hash password: %v", err)
		return
	}

	bdOwnerKey, err = gr.bdOwnerCtrler.Create(ctx, &bdo.BoardingOwner{
		Name:          data.Name,
		MainContact:   data.MainContact,
		OtherContacts: data.OtherContacts,
		Email:         data.Email,
		Password:      newPasswdHash,
		ImageUrl:      data.ImageUrl,
		Address:       data.Address,
		Location:      data.Location,
		Status:        bdo.Pending,
	})
	return
}
