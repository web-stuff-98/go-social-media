package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"github.com/web-stuff-98/go-social-media/pkg/helpers"
	"github.com/web-stuff-98/go-social-media/pkg/validation"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
	"github.com/nfnt/resize"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func (h handler) Register(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}
	inbox := &models.Inbox{}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	var credentialsInput validation.Credentials
	if json.Unmarshal(body, &credentialsInput); err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	validate := validator.New()
	if err := validate.Struct(credentialsInput); err != nil {
		responseMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	user.Username = credentialsInput.Username
	user.Password = string(hash)

	inserted, err := h.Collections.UserCollection.InsertOne(r.Context(), user)
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	inbox.ID = inserted.InsertedID.(primitive.ObjectID)
	inbox.Messages = []models.PrivateMessage{}

	if _, err := h.Collections.InboxCollection.InsertOne(r.Context(), inbox); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	cookie, err := helpers.GenerateCookieAndSession(inserted.InsertedID.(primitive.ObjectID), h.Collections)
	http.SetCookie(w, &cookie)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h handler) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	var credentialsInput validation.Credentials
	if json.Unmarshal(body, &credentialsInput); err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	validate := validator.New()
	if err := validate.Struct(credentialsInput); err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	var user models.User
	if h.Collections.UserCollection.FindOne(r.Context(), bson.M{"username": bson.M{"$regex": credentialsInput.Username, "$options": "i"}}).Decode(&user); err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		} else {
			responseMessage(w, http.StatusNotFound, "Account does not exist")
		}
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentialsInput.Password)); err != nil {
		responseMessage(w, http.StatusUnauthorized, "Incorrect credentials")
		return
	}

	var pfp models.Pfp
	if err := h.Collections.PfpCollection.FindOne(r.Context(), bson.M{"_id": user.ID}).Decode(&pfp); err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	} else {
		user.Base64pfp = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(pfp.Binary.Data)
	}

	cookie, err := helpers.GenerateCookieAndSession(user.ID, h.Collections)
	http.SetCookie(w, &cookie)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	originalCookie, err := r.Cookie("refresh_token")
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "You have no token")
		return
	}
	token, err := jwt.ParseWithClaims(originalCookie.Value, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})
	rawSID := token.Claims.(*jwt.StandardClaims).Issuer
	sessionId, err := primitive.ObjectIDFromHex(rawSID)

	var session models.Session
	if h.Collections.SessionCollection.FindOne(r.Context(), bson.M{"_id": sessionId}).Decode(&session); err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		} else {
			responseMessage(w, http.StatusNotFound, "Account does not exist")
		}
		return
	}

	cookie, err := helpers.GenerateCookieAndSession(session.UID, h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var user models.User
	if h.Collections.UserCollection.FindOne(r.Context(), bson.M{"_id": session.UID}).Decode(&user); err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		} else {
			responseMessage(w, http.StatusNotFound, "Account does not exist")
		}
		return
	}

	var pfp models.Pfp
	if err := h.Collections.PfpCollection.FindOne(r.Context(), bson.M{"_id": user.ID}).Decode(&pfp); err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	} else {
		user.Base64pfp = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(pfp.Binary.Data)
	}

	http.SetCookie(w, &cookie)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	originalCookie, err := r.Cookie("refresh_token")
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	token, err := jwt.ParseWithClaims(originalCookie.Value, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})
	rawSID := token.Claims.(*jwt.StandardClaims).Issuer
	sessionId, err := primitive.ObjectIDFromHex(rawSID)

	var session models.Session
	if h.Collections.SessionCollection.FindOne(r.Context(), bson.M{"_id": sessionId}).Decode(&session); err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		} else {
			responseMessage(w, http.StatusNotFound, "Could not find your session")
		}
		return
	}

	if h.Collections.UserCollection.DeleteOne(r.Context(), bson.M{"_id": session.UID}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	clearedCookie := helpers.GetClearedCookie()
	http.SetCookie(w, &clearedCookie)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Deleted")
}

func (h handler) Logout(w http.ResponseWriter, r *http.Request) {
	_, session, _ := helpers.GetUserAndSessionFromRequest(r, h.Collections)

	clearedCookie := helpers.GetClearedCookie()
	http.SetCookie(w, &clearedCookie)

	if session != nil {
		res, err := h.Collections.SessionCollection.DeleteOne(r.Context(), bson.M{"_id": session.UID})
		if err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
		if res.DeletedCount == 0 {
			responseMessage(w, http.StatusBadRequest, "You are not logged in")
			return
		}
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Logged out")
}

func (h handler) UploadPfp(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	r.ParseMultipartForm(32 << 20) // what is << ? something to do with binary shift whatever that is. Is used here to define maximum memory in bytes.

	file, handler, err := r.FormFile("file")
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
	}
	defer file.Close()

	if handler.Size > 20*1024*1024 {
		responseMessage(w, http.StatusRequestEntityTooLarge, "File too large, max 20mb.")
		return
	}

	src, err := handler.Open()
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	var isJPEG, isPNG bool
	isJPEG = handler.Header.Get("Content-Type") == "image/jpeg"
	isPNG = handler.Header.Get("Content-Type") == "image/png"
	if !isJPEG && !isPNG {
		responseMessage(w, http.StatusBadRequest, "Only JPEG and PNG are supported")
		return
	}
	var img image.Image
	var decodeErr error
	if isJPEG {
		img, decodeErr = jpeg.Decode(src)
	} else {
		img, decodeErr = png.Decode(src)
	}
	if decodeErr != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	buf := &bytes.Buffer{}
	if img.Bounds().Dx() > img.Bounds().Dy() {
		img = resize.Resize(64, 0, img, resize.Lanczos3)
	} else {
		img = resize.Resize(0, 64, img, resize.Lanczos3)
	}
	if err := jpeg.Encode(buf, img, nil); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	pfpBytes := buf.Bytes()
	base64pfp := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(pfpBytes)
	user.Base64pfp = base64pfp

	res, err := h.Collections.PfpCollection.UpdateByID(r.Context(), user.ID, bson.M{"$set": bson.M{"binary": primitive.Binary{Data: buf.Bytes()}}})
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	if res.MatchedCount == 0 {
		_, err := h.Collections.PfpCollection.InsertOne(r.Context(), models.Pfp{
			ID:     user.ID,
			Binary: primitive.Binary{Data: buf.Bytes()},
		})
		if err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// Get all messages from conversations (except for own messages)
func (h handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	inbox := &models.Inbox{}
	if err := h.Collections.InboxCollection.FindOne(r.Context(), bson.M{"_id": user.ID}).Decode(&inbox); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(inbox)
}
