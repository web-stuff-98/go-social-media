package helpers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func createCookie(token string, expiry time.Time) http.Cookie {
	var cookie http.Cookie
	cookie.Name = "refresh_token"
	cookie.Value = token
	cookie.Expires = expiry
	cookie.Secure = os.Getenv("PRODUCTION") == "true"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteDefaultMode
	cookie.Path = "/"
	return cookie
}

func GetClearedCookie() http.Cookie {
	var cookie http.Cookie
	cookie.Name = "refresh_token"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-time.Second)
	cookie.Secure = os.Getenv("PRODUCTION") == "true"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteDefaultMode
	cookie.Path = "/"
	return cookie
}

// Used by login, register and refresh to set the cookie. Cookie is refresh token with encrypted sid
func GenerateCookieAndSession(uid primitive.ObjectID, collections db.Collections) (http.Cookie, error) {
	collections.SessionCollection.DeleteMany(context.TODO(), bson.M{"_uid": uid})
	expiry := primitive.NewDateTimeFromTime(time.Now().Add(time.Minute * 20))
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in generate cookie helper function, why is this happening?")
		}
	}()
	inserted, err := collections.SessionCollection.InsertOne(context.TODO(), models.Session{
		ExpiresAt: expiry,
		UID:       uid,
	})
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    inserted.InsertedID.(primitive.ObjectID).Hex(),
		ExpiresAt: expiry.Time().Unix(),
	})
	token, err := claims.SignedString([]byte(os.Getenv("SECRET")))
	cookie := createCookie(token, expiry.Time())
	return cookie, err
}

func GetUserAndSessionFromRequest(r *http.Request, collections db.Collections) (*models.User, *models.Session, error) {
	originalCookie, err := r.Cookie("refresh_token")
	if err != nil {
		return nil, nil, err
	}
	token, err := jwt.ParseWithClaims(originalCookie.Value, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})
	rawSID := token.Claims.(*jwt.StandardClaims).Issuer
	sessionID, err := primitive.ObjectIDFromHex(rawSID)
	if sessionID == primitive.NilObjectID {
		return nil, nil, fmt.Errorf("NIL OBJECT ID")
	}
	if err != nil {
		return nil, nil, err
	}
	var session models.Session
	var user models.User
	if collections.SessionCollection.FindOne(context.TODO(), bson.M{"_id": sessionID}).Decode(&session); err != nil {
		return nil, nil, err
	}
	if collections.UserCollection.FindOne(context.TODO(), bson.M{"_id": session.UID}).Decode(&user); err != nil {
		return nil, nil, err
	}
	if user.ID == primitive.NilObjectID {
		return nil, nil, fmt.Errorf("NIL UID")
	}
	return &user, &session, nil
}

func DownloadImageURL(inputURL string) io.ReadCloser {
	_, err := url.Parse(inputURL)
	if err != nil {
		log.Fatal("Failed to parse image url")
	}
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.URL.Opaque = req.URL.Path
			return nil
		},
	}
	resp, err := client.Get(inputURL)
	if err != nil {
		log.Fatal(err)
	}
	return resp.Body
}
func DownloadRandomImage(pfp bool) io.ReadCloser {
	if !pfp {
		return DownloadImageURL("https://picsum.photos/1100/500")
	} else {
		return DownloadImageURL("https://100k-faces.glitch.me/random-image")
	}
}

func RemoveDuplicates(strings []string) []string {
	// Create a map to hold the unique strings
	uniqueStrings := make(map[string]bool)
	// Create a new slice to hold the unique strings
	var unique []string

	// Loop through the input slice of strings
	for _, str := range strings {
		// If the string is not in the map, add it to the map
		// and append it to the unique slice
		if !uniqueStrings[str] {
			uniqueStrings[str] = true
			unique = append(unique, str)
		}
	}
	return unique
}
