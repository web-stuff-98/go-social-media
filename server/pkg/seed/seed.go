package seed

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"math"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"github.com/web-stuff-98/go-social-media/pkg/helpers"

	"github.com/lucsky/cuid"
	"github.com/nfnt/resize"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/loremipsum.v1"
)

func SeedDB(colls *db.Collections, numUsers int, numPosts int) (uids map[primitive.ObjectID]struct{}, pids map[primitive.ObjectID]struct{}, err error) {
	uids = make(map[primitive.ObjectID]struct{})
	pids = make(map[primitive.ObjectID]struct{})

	// Generate users
	for i := 0; i < numUsers; i++ {
		uid, err := generateUser(i, colls)
		if err != nil {
			return nil, nil, err
		}
		uids[uid] = struct{}{}
	}

	lipsum := loremipsum.New()

	// Generate posts
	for i := 0; i < numPosts; i++ {
		uid := randomKey(uids)
		pid, err := generatePost(colls, lipsum, uid)
		if err != nil {
			return nil, nil, err
		}
		pids[pid] = struct{}{}
	}

	return uids, pids, err
}

func generateUser(i int, colls *db.Collections) (uid primitive.ObjectID, err error) {
	r := helpers.DownloadRandomImage(true)
	var img image.Image
	var decodeErr error
	defer r.Close()
	img, decodeErr = jpeg.Decode(r)
	if decodeErr != nil {
		return primitive.NilObjectID, decodeErr
	}
	img = resize.Resize(64, 0, img, resize.Lanczos2)
	buf := &bytes.Buffer{}
	if err := jpeg.Encode(buf, img, nil); err != nil {
		return primitive.NilObjectID, err
	}
	inserted, err := colls.UserCollection.InsertOne(context.TODO(), models.User{
		Username: fmt.Sprintf("TestAcc%d", i+1),
		Password: "$2a$12$VyvB4n4y8eq6mX8of9A3OOv/FRSzxSe54sk6ptifiT82RMtGpPI4a",
	})
	if err != nil {
		return primitive.NilObjectID, err
	}
	if colls.PfpCollection.InsertOne(context.TODO(), models.Pfp{
		ID:     inserted.InsertedID.(primitive.ObjectID),
		Binary: primitive.Binary{Data: buf.Bytes()},
	}); err != nil {
		return primitive.NilObjectID, err
	}
	buf = nil
	return inserted.InsertedID.(primitive.ObjectID), nil
}

func generatePost(colls *db.Collections, lipsum *loremipsum.LoremIpsum, uid primitive.ObjectID) (primitive.ObjectID, error) {
	minWordsInTitle := 5
	minWordsInDescription := 8
	maxWordsInTitle := int(math.Max(float64(minWordsInTitle+1), float64(rand.Intn(20))))
	maxWordsInDescription := int(math.Max(float64(minWordsInDescription+1), float64(rand.Intn(20))))

	log.Println(maxWordsInTitle)
	log.Println(maxWordsInDescription)

	wordsInTitle := rand.Intn(maxWordsInTitle-minWordsInTitle) + minWordsInTitle
	wordsInDescription := rand.Intn(maxWordsInDescription-minWordsInDescription) + minWordsInDescription

	title := strings.Title(lipsum.Words(wordsInTitle))
	description := strings.Title(lipsum.Words(wordsInDescription))
	body := ""
	tags := []string{}

	var bodyTags = []string{}

	// Add heading to body
	bodyTags = append(bodyTags, encapsulateIntoHTMLTag("h1", sentence(6, 22, lipsum)))
	bodyTags = append(bodyTags, "<br/>")
	// Add subheading to body
	bodyTags = append(bodyTags, encapsulateIntoHTMLTag("h2", sentence(8, 26, lipsum)))
	bodyTags = append(bodyTags, "<br/>")
	// Add paragraphs to body
	bodyTags = append(bodyTags, paragraphs(lipsum))
	bodyTags = append(bodyTags, "<br/>")
	// Join body into string
	body = strings.Join(bodyTags, "")

	// Add 5 - 8 tags
	i := 0
	numTags := rand.Intn(8-5) + 5
	for {
		i++
		tags = append(tags, lipsum.Word())
		if i == numTags {
			break
		}
	}

	r := helpers.DownloadRandomImage(false)
	var img image.Image
	var imgThumb image.Image
	var imgBlur image.Image
	var decodeErr error
	defer r.Close()
	img, decodeErr = jpeg.Decode(r)
	if decodeErr != nil {
		return primitive.NilObjectID, decodeErr
	}
	img = resize.Resize(1000, 0, img, resize.Lanczos2)
	imgBlur = resize.Resize(10, 0, img, resize.Lanczos2)
	imgThumb = resize.Resize(500, 0, img, resize.Lanczos2)
	buf := &bytes.Buffer{}
	blurBuf := &bytes.Buffer{}
	thumbBuf := &bytes.Buffer{}
	if err := jpeg.Encode(buf, img, nil); err != nil {
		return primitive.NilObjectID, err
	}
	if err := jpeg.Encode(blurBuf, imgBlur, nil); err != nil {
		return primitive.NilObjectID, err
	}
	if err := jpeg.Encode(thumbBuf, imgThumb, nil); err != nil {
		return primitive.NilObjectID, err
	}

	re := regexp.MustCompile("[^a-z0-9]+")
	slug := re.ReplaceAllString(strings.ToLower(title), "-")
	slug = strings.TrimFunc(slug, func(r rune) bool {
		runes := []rune("-")
		return r == runes[0]
	}) + "-" + cuid.Slug()

	post := models.Post{
		Title:        title,
		Description:  description,
		Body:         body,
		Tags:         tags,
		Author:       uid,
		ImagePending: false,
		Slug:         slug,
		CreatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		ImgBlur:      "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(blurBuf.Bytes()),
	}

	inserted, err := colls.PostCollection.InsertOne(context.TODO(), post)
	if err != nil {
		return primitive.NilObjectID, err
	}

	if colls.PostThumbCollection.InsertOne(context.TODO(), models.PostThumb{
		ID:     inserted.InsertedID.(primitive.ObjectID),
		Binary: primitive.Binary{Data: thumbBuf.Bytes()},
	}); err != nil {
		return primitive.NilObjectID, err
	}
	if colls.PostImageCollection.InsertOne(context.TODO(), models.PostImage{
		ID:     inserted.InsertedID.(primitive.ObjectID),
		Binary: primitive.Binary{Data: buf.Bytes()},
	}); err != nil {
		return primitive.NilObjectID, err
	}

	return inserted.InsertedID.(primitive.ObjectID), nil
}

func sentence(minWords int, maxWords int, lipsum *loremipsum.LoremIpsum) string {
	wordCount := rand.Intn(maxWords-minWords) + minWords
	return lipsum.Words(wordCount)
}

func paragraphs(lipsum *loremipsum.LoremIpsum) string {
	numParagraphs := rand.Intn(4-1) + 1
	paragraphs := []string{}
	i := 0
	for {
		i++
		paragraphs = append(paragraphs, encapsulateIntoHTMLTag("p", lipsum.Paragraph()))
		paragraphs = append(paragraphs, "<br/>")
		if i == numParagraphs {
			break
		}
	}
	return strings.Join(paragraphs, "")
}

func encapsulateIntoHTMLTag(tag string, content string) string {
	return "<" + tag + ">" + content + "</" + tag + ">"
}

func randomKey(m map[primitive.ObjectID]struct{}) primitive.ObjectID {
	keys := make([]primitive.ObjectID, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys[rand.Intn(len(keys))]
}
