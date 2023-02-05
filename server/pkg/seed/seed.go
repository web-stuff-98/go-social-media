package seed

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"log"
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

func SeedDB(colls *db.Collections, numUsers int, numPosts int, numRooms int, uids map[primitive.ObjectID]struct{}, pids map[primitive.ObjectID]struct{}, rids map[primitive.ObjectID]struct{}) (err error) {
	log.Println("Generating seed...")

	// Generate users
	for i := 0; i < numUsers; i++ {
		uid, err := generateUser(i, colls)
		if err != nil {
			log.Fatalln(err)
		}
		uids[uid] = struct{}{}
	}

	lipsum := loremipsum.NewWithSeed(rand.Int63())

	// Generate posts
	for i := 0; i < numPosts; i++ {
		uid := randomKey(uids)
		pid, err := generatePost(colls, lipsum, uid)
		if err != nil {
			log.Fatalln(err)
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
			log.Fatalln(err)
		}
		rids[rid] = struct{}{}
	}

	log.Println("Seed generated...")

	return nil
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
	var notifications models.Notifications
	inbox.ID = inserted.InsertedID.(primitive.ObjectID)
	inbox.Messages = []models.PrivateMessage{}
	inbox.MessagesSentTo = []primitive.ObjectID{}
	notifications.ID = inserted.InsertedID.(primitive.ObjectID)
	notifications.Notifications = []models.Notification{}
	if _, err := colls.InboxCollection.InsertOne(context.TODO(), inbox); err != nil {
		return primitive.NilObjectID, err
	}
	if _, err := colls.NotificationsCollection.InsertOne(context.TODO(), inbox); err != nil {
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
	title := strings.Title(sentence(12, 18))
	description := strings.Title(sentence(12, 23))
	tags := []string{}

	/*rb := helpers.DownloadURL("https://loripsum.net/api/link/ul/ol/dl/bq/code/headers/decorate/long")
	bodyBytes, err := ioutil.ReadAll(rb)*/
	bodyTags := []string{}
	bodyTags = append(bodyTags, encapsulateIntoHTMLTag("h1", strings.Title(sentence(8, 22))))
	bodyTags = append(bodyTags, encapsulateIntoHTMLTag("h2", strings.Title(sentence(8, 22))))
	bodyTags = append(bodyTags, paragraphs(lipsum))
	if rand.Float32() > 0.5 {
		bodyTags = append(bodyTags, encapsulateIntoHTMLTag("h3", strings.Title(sentence(5, 15))))
	}
	if rand.Float32() > 0.5 {
		bodyTags = append(bodyTags, list())
	}
	bodyTags = append(bodyTags, paragraphs(lipsum))
	if rand.Float32() > 0.5 {
		bodyTags = append(bodyTags, encapsulateIntoHTMLTag("h3", strings.Title(sentence(5, 15))))
	}
	if rand.Float32() > 0.5 {
		bodyTags = append(bodyTags, list())
	}
	bodyTags = append(bodyTags, paragraphs(lipsum))
	if rand.Float32() > 0.5 {
		bodyTags = append(bodyTags, list())
		bodyTags = append(bodyTags, encapsulateIntoHTMLTag("p", sentence(10, 35)))
	}
	var body string
	for _, v := range bodyTags {
		body += v
	}

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
	imgThumb = resize.Resize(600, 0, img, resize.Lanczos2)
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

	timeRandNegative := 1
	if rand.Float32() > 0.5 {
		timeRandNegative = -1
	}
	timeRandOffset := time.Hour * time.Duration(rand.Float32()*24*float32(timeRandNegative))
	createdAt := time.Now().Add(timeRandOffset)
	updatedAt := time.Now()
	if rand.Float32() > 0.8 {
		timeRandOffsetB := time.Hour * time.Duration(rand.Float32()*24)
		updatedAt = createdAt.Add(timeRandOffsetB)
	}

	post := models.Post{
		Title:       title,
		Description: description,
		//Body:          string(bodyBytes),
		Body:          body,
		Tags:          tags,
		Author:        uid,
		ImagePending:  false,
		Slug:          slug,
		CreatedAt:     primitive.NewDateTimeFromTime(createdAt),
		UpdatedAt:     primitive.NewDateTimeFromTime(updatedAt),
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
		log.Fatalln(err)
	}

	if colls.PostCollection.UpdateByID(context.TODO(), pid, bson.M{"$set": bson.M{"sort_vote_count": positiveVotes - negativeVotes}}); err != nil {
		log.Fatalln(err)
	}

	return nil
}

func generateRoom(colls *db.Collections, lipsum *loremipsum.LoremIpsum, uid primitive.ObjectID, i int) (primitive.ObjectID, error) {
	name := "ExampleRoom" + strconv.Itoa(i+1)

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
		Private:      false,
	}

	inserted, err := colls.RoomCollection.InsertOne(context.TODO(), room)
	if err != nil {
		return primitive.NilObjectID, err
	}

	if colls.RoomPrivateDataCollection.InsertOne(context.TODO(), models.RoomPrivateData{
		ID:      inserted.InsertedID.(primitive.ObjectID),
		Members: []primitive.ObjectID{},
		Banned:  []primitive.ObjectID{},
	}); err != nil {
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
		if rand.Float32() > 0.75 {
			updatedAt = createdAt.Add(time.Hour * 2 * time.Duration(rand.Float32()))
		}
		if rand.Float32() > 0.2 {
			if len(comments.Comments) > 0 {
				randIndex := len(comments.Comments) - 1
				if randIndex == 0 || randIndex == -1 {
					randIndex = 1
				}
				parentId = comments.Comments[rand.Intn(randIndex)].ID.Hex()
			}
		}
		content := sentence(7, 50)
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
		log.Fatalln(err)
	}

	return nil
}

func sentence(minWords int, maxWords int) string {
	wordCount := (rand.Intn(maxWords-minWords) + minWords) + 8
	lipsum := loremipsum.NewWithSeed(int64(rand.Intn(1000)))
	return strings.ReplaceAll(strings.ToLower(lipsum.Words(wordCount)), "lorem ipsum dolor sit amet consectetur adipiscing elit", "")
}

func list() string {
	items := []string{}
	numItems := rand.Intn(4) + 3
	for i := 0; i < numItems; i++ {
		items = append(items, encapsulateIntoHTMLTag("li", sentence(5, 30)))
	}
	listType := "ol"
	if rand.Float32() > 0.5 {
		listType = "ul"
	}
	return encapsulateIntoHTMLTag(listType, strings.Join(items, ""))
}

func paragraphs(lipsum *loremipsum.LoremIpsum) string {
	numParagraphs := rand.Intn(4-1) + 1
	paragraphs := []string{}
	i := 0
	lastWasSpecialTag := false
	for {
		i++
		if rand.Float32() > 0.5 && !lastWasSpecialTag {
			lastWasSpecialTag = true
			tagType := "b"
			if rand.Float32() > 0.75 {
				tagType = "mark"
			}
			paragraphs = append(paragraphs, encapsulateIntoHTMLTag("p", sentence(8, 22)+encapsulateIntoHTMLTag(tagType, sentence(5, 10)+sentence(8, 22))))
		} else {
			paragraphs = append(paragraphs, lipsum.Paragraph())
		}
		lastWasSpecialTag = false
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
