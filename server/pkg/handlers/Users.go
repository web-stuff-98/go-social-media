package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (h handler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rawId, _ := vars["id"]
	id, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var user models.User
	if err := h.Collections.UserCollection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	recvChan := make(chan bool)
	h.SocketServer.GetUserOnlineStatus <- socketserver.GetUserOnlineStatus{
		RecvChan: recvChan,
		Uid:      id,
	}
	isOnline := <-recvChan
	user.IsOnline = isOnline

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h handler) GetPfp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rawId, _ := vars["id"]
	id, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var pfp models.Pfp
	if err := h.Collections.PfpCollection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&pfp); err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(pfp.Binary.Data)))
	if _, err := w.Write(pfp.Binary.Data); err != nil {
		log.Println("Unable to write image to response")
	}
}
