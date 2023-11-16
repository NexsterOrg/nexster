package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	blclient "github.com/NamalSanjaya/nexster/content/pkg/client/blob"
	avtr "github.com/NamalSanjaya/nexster/content/pkg/repository/avatar"
	mdr "github.com/NamalSanjaya/nexster/content/pkg/repository/media"
	"github.com/NamalSanjaya/nexster/pkgs/crypto/hmac"
	chttp "github.com/NamalSanjaya/nexster/pkgs/utill/http" // custom http library
	"github.com/NamalSanjaya/nexster/pkgs/utill/uuid"
)

const (
	avatarNS       string = "avatar"
	postNS         string = "post"
	eventPostersNs string = "event-posters"
	publicView     string = "public"
	privateView    string = "private"
)

// query parameters
const (
	permission string = "perm"
	timestamp  string = "ts"
	imageHMac  string = "imgMac"
)

type server struct {
	config     *ServerConfig
	blobClient blclient.Interface
	avatarRepo avtr.Interface
	mediaRepo  mdr.Interface
	logger     *lg.Logger
}

var _ Interface = (*server)(nil)

func New(cfg *ServerConfig, logger *lg.Logger, blClient blclient.Interface, avatarIntfce avtr.Interface, mediaIntfce mdr.Interface) *server {
	return &server{
		config:     cfg,
		logger:     logger,
		blobClient: blClient,
		avatarRepo: avatarIntfce,
		mediaRepo:  mediaIntfce,
	}
}

// imageId + perm + ts
func (s *server) ServeImages(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	namespace := p.ByName("namespace")
	blobName := p.ByName("imgId") // image Id (eg: 18733627.png)

	perm := r.URL.Query().Get(permission) // owner | viewer
	ts := r.URL.Query().Get(timestamp)    // ts - timestamp
	imgMac := r.URL.Query().Get(imageHMac)

	// check whether content is modified or not
	if !hmac.ValidateHMAC(s.config.SecretImgKey, imgMac, blobName, perm, ts) {
		// return unAuthorized
		s.logger.Infof("failed to server image: unauthorized access: hmac valdiation failed for blob %s in namespace: %s", blobName, namespace)
		s.sendRespDefault(w, http.StatusUnauthorized, map[string]interface{}{})
		return
	}

	var view string
	var err error
	if namespace == avatarNS {
		view, err = s.avatarRepo.GetView(r.Context(), getImgKey(blobName))
	} else if namespace == postNS {
		view, err = s.mediaRepo.GetView(r.Context(), getImgKey(blobName))
	} else if namespace == eventPostersNs {
		// event posters have public view in the current system.
		view = publicView
	} else {
		// TODO: space for futher namespaces
		return
	}

	if err != nil {
		s.logger.Errorf("failed to server image: failed to get the view from %s repo: %v", namespace, err)
		s.sendRespDefault(w, http.StatusInternalServerError, map[string]interface{}{})
		return
	}

	var permOk bool
	if view == publicView && (perm == "viewer" || perm == "owner") {
		permOk = true
	} else if view == privateView && perm == "owner" {
		permOk = true
	}

	if !permOk {
		s.logger.Infof("failed to server image: unauthorized access: insufficient permission to access blob %s in namespace: %s", blobName, namespace)
		s.sendRespDefault(w, http.StatusUnauthorized, map[string]interface{}{})
		return
	}

	imgReader, contentType, err := s.blobClient.ImageReader(r.Context(), getBlobFullName(namespace, blobName))
	if err != nil {
		s.logger.Errorf("failed to server image: failed to read image from blob stroage: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, map[string]interface{}{})
		return
	}
	defer imgReader.Close()

	w.Header().Add("Content-Type", contentType)
	w.Header().Add("Cache-Control", "max-age=3600, private")
	w.Header().Add("Date", "")

	io.Copy(w, imgReader)
}

func (s *server) CreateImgUrl(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	namespace := p.ByName("namespace") // eg: avatar, post, poster
	imgId := p.ByName("imgId")         // eg: 1348502.png
	rndTimestamp := strconv.Itoa(rand.Intn(10000))
	imgUrl := fmt.Sprintf("%s/content/images/%s/%s?%s=%s&%s=%s&%s=%s", s.config.Url, namespace, imgId,
		permission, r.URL.Query().Get(permission),
		timestamp, rndTimestamp,
		imageHMac, hmac.CalculateHMAC(s.config.SecretImgKey,
			p.ByName("imgId"),
			r.URL.Query().Get("perm"),
			rndTimestamp,
		),
	)

	s.sendRespDefault(w, http.StatusOK, map[string]interface{}{"url": imgUrl})
}

func (s *server) UploadImage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	defaultBody := map[string]interface{}{
		"state": chttp.Failed,
		"data":  map[string]string{},
	}
	namespace := p.ByName("namespace")
	imgType := r.URL.Query().Get("type")
	if imgType == "" {
		s.logger.Info("failed to upload image: type query parameter is empty")
		s.sendRespDefault(w, http.StatusBadRequest, defaultBody)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Errorf("failed to upload image: failed to read request body: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, defaultBody)
		return
	}

	imageBytes, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		s.logger.Errorf("failed to upload image: decode base64 image to []byte: %v", err)
		s.sendRespDefault(w, http.StatusBadRequest, defaultBody)
		return
	}
	imgFullname, err := s.blobClient.UploadImage(r.Context(), imgType, imageBytes, &blclient.UploadImageOptions{
		BlobName: getBlobFullName(namespace, uuid.GenUUID4()),
	})
	if err != nil {
		s.logger.Errorf("failed to upload image: failed to upload image: %v", err)
		s.sendRespDefault(w, http.StatusInternalServerError, defaultBody)
		return
	}
	s.sendRespDefault(w, http.StatusCreated, map[string]interface{}{
		"state": chttp.Success,
		"data":  map[string]string{"imageName": imgFullname},
	})
}

func (s *server) sendRespDefault(w http.ResponseWriter, statusCode int, body map[string]interface{}) {
	w.Header().Add(ContentType, ApplicationJson_Utf8)
	w.Header().Add(Date, "")
	w.WriteHeader(statusCode)
	resp, _ := json.Marshal(body)
	w.Write(resp)
}

func getBlobFullName(namespace, blobName string) string {
	return fmt.Sprintf("%s/%s", namespace, blobName)
}

func getImgKey(input string) string {
	parts := strings.Split(input, ".")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

func mkBlobFullName(namespace, imgId, imgType string) string {
	return fmt.Sprintf("%s/%s.%s", namespace, imgId, imgType)
}
