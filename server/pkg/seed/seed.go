package seed

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/loremipsum.v1"
)

func SeedDB(colls *db.Collections, numUsers int, numPosts int, numRooms int) (uids map[primitive.ObjectID]struct{}, pids map[primitive.ObjectID]struct{}, rids map[primitive.ObjectID]struct{}, err error) {
	uids = make(map[primitive.ObjectID]struct{})
	pids = make(map[primitive.ObjectID]struct{})
	rids = make(map[primitive.ObjectID]struct{})

	// Generate users
	for i := 0; i < numUsers; i++ {
		uid, err := generateUser(i, colls)
		if err != nil {
			return nil, nil, nil, err
		}
		uids[uid] = struct{}{}
	}

	lipsum := loremipsum.New()

	// Generate posts
	for i := 0; i < numPosts; i++ {
		uid := randomKey(uids)
		pid, err := generatePost(colls, lipsum, uid)
		if err != nil {
			return nil, nil, nil, err
		}
		pids[pid] = struct{}{}
	}

	// Generate rooms
	for i := 0; i < numPosts; i++ {
		uid := randomKey(uids)
		rid, err := generateRoom(colls, lipsum, uid, i)
		if err != nil {
			return nil, nil, nil, err
		}
		rids[rid] = struct{}{}
	}

	// Generate post votes
	for pid := range pids {
		generatePostVotes(colls, pid, uids)
	}

	return uids, pids, rids, err
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
	var inbox models.Inbox
	inbox.ID = inserted.InsertedID.(primitive.ObjectID)
	inbox.Messages = []models.PrivateMessage{}
	inbox.MessagesSentTo = []primitive.ObjectID{}
	if _, err := colls.InboxCollection.InsertOne(context.TODO(), inbox); err != nil {
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

	tags = helpers.RemoveDuplicates(tags)

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
		Title:         title,
		Description:   description,
		Body:          body,
		Tags:          tags,
		Author:        uid,
		ImagePending:  false,
		Slug:          slug,
		CreatedAt:     primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:     primitive.NewDateTimeFromTime(time.Now()),
		ImgBlur:       "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(blurBuf.Bytes()),
		SortVoteCount: 0,
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
	if colls.PostVoteCollection.InsertOne(context.TODO(), models.PostVotes{
		ID:    inserted.InsertedID.(primitive.ObjectID),
		Votes: []models.PostVote{},
	}); err != nil {
		return primitive.NilObjectID, err
	}

	return inserted.InsertedID.(primitive.ObjectID), nil
}

func generatePostVotes(colls *db.Collections, pid primitive.ObjectID, uids map[primitive.ObjectID]struct{}) error {
	uidsArray := []primitive.ObjectID{}
	for id := range uids {
		uidsArray = append(uidsArray, id)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(uidsArray), func(i, j int) { uidsArray[i], uidsArray[j] = uidsArray[j], uidsArray[i] })

	positiveVotes := 0
	negativeVotes := 0

	votes := models.PostVotes{ID: pid, Votes: []models.PostVote{}}
	for i := 0; i < len(uidsArray); i++ {
		pos := rand.Float32() > 0.5
		vote := &models.PostVote{
			ID:       primitive.NewObjectID(),
			Uid:      uidsArray[i],
			IsUpvote: pos,
		}
		if pos {
			positiveVotes++
		} else {
			negativeVotes++
		}
		votes.Votes = append(votes.Votes, *vote)
	}

	_, err := colls.PostVoteCollection.UpdateByID(context.TODO(), pid, bson.M{"$set": bson.M{"votes": votes.Votes}})
	if err != nil {
		return err
	}

	if colls.PostCollection.UpdateByID(context.TODO(), pid, bson.M{"$set": bson.M{"sort_vote_count": positiveVotes - negativeVotes}}); err != nil {
		return err
	}

	return nil
}

func generateRoom(colls *db.Collections, lipsum *loremipsum.LoremIpsum, uid primitive.ObjectID, i int) (primitive.ObjectID, error) {
	name := "ExampleRoom" + string(rune(i))

	r := helpers.DownloadRandomImage(false)
	var img image.Image
	var imgBlur image.Image
	var decodeErr error
	defer r.Close()
	img, decodeErr = jpeg.Decode(r)
	if decodeErr != nil {
		return primitive.NilObjectID, decodeErr
	}
	img = resize.Resize(400, 0, img, resize.Lanczos2)
	imgBlur = resize.Resize(10, 0, img, resize.Lanczos2)
	buf := &bytes.Buffer{}
	blurBuf := &bytes.Buffer{}
	if err := jpeg.Encode(buf, img, nil); err != nil {
		return primitive.NilObjectID, err
	}
	if err := jpeg.Encode(blurBuf, imgBlur, nil); err != nil {
		return primitive.NilObjectID, err
	}

	room := models.Room{
		Name:         name,
		Author:       uid,
		ImagePending: false,
		CreatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		ImgBlur:      "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(blurBuf.Bytes()),
	}

	inserted, err := colls.RoomCollection.InsertOne(context.TODO(), room)
	if err != nil {
		return primitive.NilObjectID, err
	}

	if colls.RoomImageCollection.InsertOne(context.TODO(), models.RoomImage{
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
