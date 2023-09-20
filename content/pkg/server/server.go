package server

import (
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"

	blclient "github.com/NamalSanjaya/nexster/content/pkg/client/blob"
)

type server struct {
	blobClient blclient.Interface
}

func New(blClient blclient.Interface) *server {
	return &server{
		blobClient: blClient,
	}
}

func (s *server) ServeImages(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	blobName := p.ByName("imgId")
	imgReader, contentType, err := s.blobClient.ImageReader(r.Context(), blobName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer imgReader.Close()

	w.Header().Add("Content-Type", contentType)
	w.Header().Add("Cache-Control", "max-age=3600, private")
	w.Header().Add("Date", "")

	io.Copy(w, imgReader)
}
