package http

import (
	"encoding/json"
	"log"
	"net/http"
)

// Set json response with default response parameters
func SendDefaultResp(w http.ResponseWriter, statusCode int, body map[string]interface{}) {
	w.Header().Add(ContentType, ApplicationJson_Utf8)
	w.Header().Add(Date, "")
	resp, err := json.Marshal(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("failed to marshal json body: ", err)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(resp)
}
