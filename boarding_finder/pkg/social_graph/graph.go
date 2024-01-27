package socialgraph

import (
	"context"
	"fmt"
	"log"

	dtm "github.com/NamalSanjaya/nexster/boarding_finder/pkg/dto_mapper"
	rp "github.com/NamalSanjaya/nexster/boarding_finder/pkg/repository"
	contapi "github.com/NamalSanjaya/nexster/pkgs/client/content_api"
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
	contentClient   contapi.Interface
}

var _ Interface = (*socialGraph)(nil)

func NewGraph(bdAdsIntfce bda.Interface, bdAdOwnedIntfce bao.Interface, bdOwnerIntfce bdo.Interface, repoIntfce rp.Interface,
	contentApiClient contapi.Interface) *socialGraph {
	return &socialGraph{
		bdAdsCtrler:     bdAdsIntfce,
		bdAdOwnedCtrler: bdAdOwnedIntfce,
		bdOwnerCtrler:   bdOwnerIntfce,
		repo:            repoIntfce,
		contentClient:   contentApiClient,
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
		Title:               data.Title,
		Description:         data.Description,
		Bills:               data.Bills,
		ImageUrls:           data.ImageUrls,
		Rent:                data.Rent,
		Address:             data.Address,
		Beds:                data.Beds,
		Baths:               data.Baths,
		Gender:              data.Gender,
		Distance:            data.Distance,
		DistanceUnit:        data.DistanceUnit,
		Status:              bda.Pending,
		LocationSameAsOwner: false, // TODO: Update with user input value
	})
	if err != nil {
		return
	}
	ownedEdgeKey, err = gr.bdAdOwnedCtrler.CreateDocument(ctx, bda.MkBdAdsDocId(adNodeKey), bdo.MkDocId(bdOwnerKey))
	return
}

// Create a boarding owner node with pending status
func (gr *socialGraph) CreateBoardingOwner(ctx context.Context, data *dtm.CreateBoardingOwner, roles []string) (bdOwnerKey string, err error) {
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
		Roles:         roles,
	})
	return
}

func (gr *socialGraph) GetAdForMainView(ctx context.Context, adKey string) (adWithOwner *dtm.AdsWithOwner, err error) {
	adWithOwner = &dtm.AdsWithOwner{}
	result, err := gr.bdAdsCtrler.GetAdWithOwner(ctx, bda.MkBdAdsDocId(adKey))
	if err != nil {
		return
	}
	// check ad status
	if result.From.Status != bda.Accepted {
		return
	}
	// check owner data
	if result.To.Status != bdo.Active {
		return
	}
	adWithOwner = dtm.ConvertAdWithOwnerData(result)
	if result.From.LocationSameAsOwner {
		adWithOwner.Ad.Address = ""
		adWithOwner.Ad.Distance = 0
		adWithOwner.Ad.DistanceUnit = ""
	} else {
		adWithOwner.Owner.Address = ""
		adWithOwner.Owner.Location = ""
	}
	return
}

func (gr *socialGraph) ChangeAdStatus(ctx context.Context, adKey, status string) error {
	return gr.bdAdsCtrler.Update(ctx, adKey, map[string]interface{}{
		"status": status,
	})
}

// TODO: Need to improve this function
func (gr *socialGraph) ListAdsWithFilters(ctx context.Context, data *dtm.ListFilterQueryParams) (ads []*dtm.AdForList, adsCount int, err error) {
	ads = []*dtm.AdForList{}
	results, err := gr.bdAdsCtrler.ListAdsWithFilters(ctx, data.MinRent, data.MaxRent, data.MaxDistance, data.MinBeds, data.MaxBeds, data.MinBaths, data.MaxBaths, (data.Pg-1)*data.PgSize, data.PgSize, data.SortBy, data.Genders, data.BillTypes)
	if err != nil {
		return
	}

	for _, ad := range results {
		imgs := ad.ImageUrls
		if len(imgs) == 0 {
			// Ads without at least one image is not allowed.
			continue
		}

		coverImgUrl, err2 := gr.contentClient.CreateImageUrl(imgs[0], contapi.Viewer)
		if err2 != nil {
			log.Println("failed to create cover image url for ad: ", err2)
			continue
		}
		ads = append(ads, &dtm.AdForList{
			Key:       ad.Key,
			Title:     ad.Title,
			ImageUrl:  coverImgUrl,
			Rent:      ad.Rent,
			Beds:      ad.Beds,
			Baths:     ad.Baths,
			Gender:    ad.Gender,
			Distance:  ad.Distance,
			CreatedAt: ad.CreatedAt,
		})
		adsCount++
	}
	return
}

// check the boarding owner existence for a given phone number
func (gr *socialGraph) IsBoardingOwnerExist(ctx context.Context, phoneNo string) (bool, error) {
	return gr.repo.ExistAndUniqueForMainContact(ctx, phoneNo)
}

func (gr *socialGraph) IsAdOwner(ctx context.Context, adKey, userKey string) (bool, error) {
	exist, err := gr.repo.IsUniqueEdgeExist(ctx, bda.MkBdAdsDocId(adKey), bdo.MkDocId(userKey))
	if err != nil {
		return false, fmt.Errorf("failed to check ad ownership existence: %v", err)
	}
	return exist, nil
}

func (gr *socialGraph) DeleteAd(ctx context.Context, adKey, userKey string) (err error) {
	if err = gr.bdAdsCtrler.Delete(ctx, adKey); err != nil {
		return
	}
	if err = gr.repo.DelEdgeFromTo(ctx, bda.MkBdAdsDocId(adKey), bdo.MkDocId(userKey)); err != nil {
		return
	}
	return
}
