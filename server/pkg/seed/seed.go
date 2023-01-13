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
	"strconv"
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

	log.Println("Generating seed...")

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

	// Generate post comments
	for pid := range pids {
		generateComments(colls, pid, uids)
	}

	// Generate post votes
	for pid := range pids {
		generatePostVotes(colls, pid, uids)
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

	log.Println("Seed generated...")

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
	if colls.PostCommentsCollection.InsertOne(context.TODO(), models.PostComments{
		ID:       inserted.InsertedID.(primitive.ObjectID),
		Comments: []models.PostComment{},
		Votes:    []models.PostCommentVote{},
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

	bias := rand.Float32() * (rand.Float32() * 2)

	votes := models.PostVotes{ID: pid, Votes: []models.PostVote{}}
	for i := 0; i < len(uidsArray); i++ {
		if rand.Float32() > 0.8 {
			// at random, dont vote
			return nil
		}
		pos := rand.Float32() < bias
		vote := &models.PostVote{
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
	name := "ExampleRoom" + strconv.Itoa(i)

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
		ID:           primitive.NewObjectID(),
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

	if colls.RoomMessagesCollection.InsertOne(context.TODO(), models.RoomMessages{
		ID:       inserted.InsertedID.(primitive.ObjectID),
		Messages: []models.RoomMessage{},
	}); err != nil {
		return primitive.NilObjectID, err
	}

	return inserted.InsertedID.(primitive.ObjectID), nil
}

func generateComments(colls *db.Collections, pid primitive.ObjectID, uids map[primitive.ObjectID]struct{}) error {
	uidsArray := []primitive.ObjectID{}
	for id := range uids {
		uidsArray = append(uidsArray, id)
	}

	comments := &models.PostComments{}
	var max int = rand.Intn(300-50) + 50
	var min int = 0
	numCmts := rand.Intn(max-min) + min
	for i := 0; i < numCmts; i++ {
		timeRandNegative := 1
		if rand.Float32() > 0.5 {
			timeRandNegative = -1
		}
		timeRandOffset := time.Hour * time.Duration(rand.Float32()*3*float32(timeRandNegative))
		createdAt := time.Now().Add(timeRandOffset)
		updatedAt := createdAt
		parentId := ""
		if rand.Float32() > 0.8 {
			updatedAt = createdAt.Add(time.Hour * 2 * time.Duration(rand.Float32()))
		}
		if rand.Float32() > 0.333 {
			if len(comments.Comments) > 0 {
				randIndex := len(comments.Comments) - 1
				if randIndex == 0 || randIndex == -1 {
					randIndex = 1
				}
				parentId = comments.Comments[rand.Intn(randIndex)].ID.Hex()
			}
		}
		lipsum := loremipsum.NewWithSeed(int64(rand.Intn(1000)))
		content := sentence(3, 60, lipsum)
		if len(content) > 200 {
			content = content[:200]
		}
		cmt := &models.PostComment{
			ID:        primitive.NewObjectID(),
			Author:    uidsArray[rand.Intn(len(uidsArray)-1)],
			Content:   content,
			CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
			UpdatedAt: primitive.NewDateTimeFromTime(updatedAt),
			ParentID:  parentId,
		}
		comments.Comments = append(comments.Comments, *cmt)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(uidsArray), func(i, j int) { uidsArray[i], uidsArray[j] = uidsArray[j], uidsArray[i] })

	for _, c := range comments.Comments {
		numVoters := rand.Intn(len(uidsArray))
		for i := 0; i < numVoters; i++ {
			if rand.Float32() > 0.5 {
				// at random, dont vote
				break
			}
			vote := &models.PostCommentVote{
				Uid:       uidsArray[i],
				IsUpvote:  rand.Float32() > 0.4,
				CommentID: c.ID,
			}
			comments.Votes = append(comments.Votes, *vote)
		}
	}

	if _, err := colls.PostCommentsCollection.UpdateByID(context.TODO(), pid, bson.M{"$set": bson.M{"comments": comments.Comments, "votes": comments.Votes}}); err != nil {
		return err
	}

	return nil
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
