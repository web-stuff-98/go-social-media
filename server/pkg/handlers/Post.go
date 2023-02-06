package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"github.com/web-stuff-98/go-social-media/pkg/helpers"
	"github.com/web-stuff-98/go-social-media/pkg/socketmodels"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"
	"github.com/web-stuff-98/go-social-media/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/lucsky/cuid"
	"github.com/nfnt/resize"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h handler) VoteOnPost(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawId := mux.Vars(r)["id"]
	postId, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	var voteInput validation.Vote
	if json.Unmarshal(body, &voteInput); err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	validate := validator.New()
	if err := validate.Struct(voteInput); err != nil {
		responseMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	votes := &models.PostVotes{}
	if err := h.Collections.PostVoteCollection.FindOne(r.Context(), bson.M{"_id": postId}).Decode(&votes); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	removeVote := false
	removeVoteIsUpvote := false
	for _, pv := range votes.Votes {
		if pv.Uid == user.ID {
			removeVoteIsUpvote = pv.IsUpvote
			removeVote = true
		}
	}

	var positiveVotes int = 0
	var negativeVotes int = 0
	for _, v := range votes.Votes {
		if user != nil {
			if user.ID != v.Uid {
				if v.IsUpvote == true {
					positiveVotes++
				} else {
					negativeVotes++
				}
			}
		} else {
			if v.IsUpvote == true {
				positiveVotes++
			} else {
				negativeVotes++
			}
		}
	}

	if !removeVote {
		if voteInput.IsUpvote {
			positiveVotes++
		} else {
			negativeVotes++
		}
		if _, err := h.Collections.PostVoteCollection.UpdateByID(r.Context(), postId, bson.M{
			"$addToSet": bson.M{
				"votes": bson.M{
					"uid":       user.ID,
					"is_upvote": voteInput.IsUpvote,
				},
			},
		}); err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	} else {
		if removeVoteIsUpvote {
			positiveVotes--
		} else {
			negativeVotes--
		}
		if _, err := h.Collections.PostVoteCollection.UpdateByID(r.Context(), postId, bson.M{
			"$pull": bson.M{
				"votes": bson.M{
					"uid": user.ID,
				},
			},
		}); err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	}

	if _, err := h.Collections.PostCollection.UpdateByID(r.Context(), postId, bson.M{"$set": bson.M{"sort_vote_count": positiveVotes - negativeVotes}}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
	}

	outBytes, err := json.Marshal(socketmodels.OutMessage{
		Type: "POST_VOTE",
		Data: `{"ID":"` + postId.Hex() + `","is_upvote":` + strconv.FormatBool(voteInput.IsUpvote) + `,"remove":` + strconv.FormatBool(removeVote) + `}`,
	})

	h.SocketServer.SendDataToSubscriptionExclusive <- socketserver.ExclusiveSubscriptionDataMessage{
		Name:    "post_card=" + postId.Hex(),
		Data:    outBytes,
		Exclude: map[primitive.ObjectID]bool{user.ID: true},
	}

	h.SocketServer.SendDataToSubscriptionExclusive <- socketserver.ExclusiveSubscriptionDataMessage{
		Name:    "post_page=" + postId.Hex(),
		Data:    outBytes,
		Exclude: map[primitive.ObjectID]bool{user.ID: true},
	}

	responseMessage(w, http.StatusOK, "Voted")
}

func (h handler) VoteOnPostComment(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawPostId := mux.Vars(r)["postId"]
	postId, err := primitive.ObjectIDFromHex(rawPostId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	rawCmtId := mux.Vars(r)["commentId"]
	commentId, err := primitive.ObjectIDFromHex(rawCmtId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	var voteInput validation.Vote
	if json.Unmarshal(body, &voteInput); err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	validate := validator.New()
	if err := validate.Struct(voteInput); err != nil {
		responseMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	comments := &models.PostComments{}
	if err := h.Collections.PostCommentsCollection.FindOne(r.Context(), bson.M{"_id": postId}).Decode(&comments); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	removeVote := false
	removeVoteIsUpvote := false
	for _, cv := range comments.Votes {
		if cv.CommentID == commentId && cv.Uid == user.ID {
			removeVoteIsUpvote = cv.IsUpvote
			removeVote = true
		}
	}

	var positiveVotes int = 0
	var negativeVotes int = 0
	for _, v := range comments.Votes {
		if v.CommentID == commentId {
			if user != nil {
				if user.ID != v.Uid {
					if v.IsUpvote == true {
						positiveVotes++
					} else {
						negativeVotes++
					}
				}
			} else {
				if v.IsUpvote == true {
					positiveVotes++
				} else {
					negativeVotes++
				}
			}
		}
	}

	if !removeVote {
		if voteInput.IsUpvote {
			positiveVotes++
		} else {
			negativeVotes++
		}
		if _, err := h.Collections.PostCommentsCollection.UpdateByID(r.Context(), postId, bson.M{
			"$addToSet": bson.M{
				"votes": bson.M{
					"uid":        user.ID,
					"is_upvote":  voteInput.IsUpvote,
					"comment_id": commentId.Hex(),
				},
			},
		}); err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	} else {
		if removeVoteIsUpvote {
			positiveVotes--
		} else {
			negativeVotes--
		}
		if _, err := h.Collections.PostCommentsCollection.UpdateByID(r.Context(), postId, bson.M{
			"$pull": bson.M{
				"votes": bson.M{
					"uid":        user.ID,
					"comment_id": commentId.Hex(),
				},
			},
		}); err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	}

	outBytes, err := json.Marshal(socketmodels.OutMessage{
		Type: "POST_COMMENT_VOTE",
		Data: `{"ID":"` + commentId.Hex() + `","is_upvote":` + strconv.FormatBool(voteInput.IsUpvote) + `,"remove":` + strconv.FormatBool(removeVote) + `}`,
	})

	h.SocketServer.SendDataToSubscriptionExclusive <- socketserver.ExclusiveSubscriptionDataMessage{
		Name:    "post_page=" + postId.Hex(),
		Data:    outBytes,
		Exclude: map[primitive.ObjectID]bool{user.ID: true},
	}

	responseMessage(w, http.StatusOK, "Voted")
}

func (h handler) CommentOnPost(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	parentId := ""

	if r.URL.Query().Has("parent_id") {
		parentId = r.URL.Query().Get("parent_id")
	}

	var commentInput validation.PostComment
	if json.Unmarshal(body, &commentInput); err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	validate := validator.New()
	if err := validate.Struct(commentInput); err != nil {
		responseMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	rawPostId := mux.Vars(r)["postId"]
	postId, err := primitive.ObjectIDFromHex(rawPostId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var post models.Post
	if h.Collections.PostCollection.FindOne(r.Context(), bson.M{"_id": postId}).Decode(&post); err != nil {
		responseMessage(w, http.StatusNotFound, "Post not found")
		return
	}

	comment := models.PostComment{
		ID:        primitive.NewObjectIDFromTimestamp(time.Now()),
		ParentID:  parentId,
		Author:    user.ID,
		Content:   commentInput.Content,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	commentsRes, err := h.Collections.PostCommentsCollection.UpdateByID(r.Context(), post.ID, bson.M{"$push": bson.M{"comments": comment}})
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	if commentsRes.MatchedCount == 0 {
		_, err := h.Collections.PostCommentsCollection.InsertOne(r.Context(), models.PostComments{
			ID:       post.ID,
			Comments: []models.PostComment{comment},
		})
		if err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	}

	jsonBytes, err := json.Marshal(comment)
	outBytes, err := json.Marshal(socketmodels.OutChangeMessage{
		Type:   "CHANGE",
		Method: "INSERT",
		Entity: "POST_COMMENT",
		Data:   string(jsonBytes),
	})

	h.SocketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
		Name: "post_page=" + post.ID.Hex(),
		Data: outBytes,
	}

	// Create notification for recipient if they aren't already on the post page
	/*if parentId != "" {
		var cmts models.PostComments
		if err := h.Collections.PostCommentsCollection.FindOne(r.Context(), bson.M{"_id": postId}).Decode(&cmts); err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		} else {
			replyingToSelf := false
			var cmt models.PostComment
			for _, pc := range cmts.Comments {
				if pc.ID.Hex() == parentId {
					cmt = pc
					if cmt.Author == user.ID {
						replyingToSelf = true
					}
					break
				}
			}
			if !replyingToSelf {
				for _, oi := range h.SocketServer.Subscriptions["post_page="+postId.Hex()] {
					found := false
					if oi == cmt.Author {
						found = true
						break
					}
					if !found {
						h.Collections.NotificationsCollection.UpdateByID(r.Context(), cmt.Author, bson.M{"$push": bson.M{"notifications": models.Notification{
							Type: "REPLY:" + postId.Hex() + ":" + user.ID.Hex(),
						}}})
					}
				}
			}
		}
	}*/

	responseMessage(w, http.StatusCreated, "Comment added")
}

func (h handler) DeleteCommentOnPost(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawPostId := mux.Vars(r)["postId"]
	rawId := mux.Vars(r)["id"]
	id, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	postId, err := primitive.ObjectIDFromHex(rawPostId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	res, err := h.Collections.PostCommentsCollection.UpdateByID(r.Context(), postId, bson.M{
		"$pull": bson.M{
			"comments": bson.M{
				"_id":       id,
				"author_id": user.ID,
			},
			"votes": bson.M{
				"comment_id": id,
			},
		},
	})
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	if res.ModifiedCount == 0 {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	if _, err := h.Collections.PostCommentsCollection.UpdateByID(r.Context(), postId, bson.M{
		"$pull": bson.M{
			"comments": bson.M{
				"parent_id": id,
			},
		},
	}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	outBytes, err := json.Marshal(socketmodels.OutChangeMessage{
		Type:   "CHANGE",
		Method: "DELETE",
		Entity: "POST_COMMENT",
		Data:   `{"ID":"` + rawId + `"}`,
	})

	h.SocketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
		Name: "post_page=" + rawPostId,
		Data: outBytes,
	}

	responseMessage(w, http.StatusOK, "Comment deleted")
}

func (h handler) UpdatePostComment(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	var commentInput validation.PostComment
	if json.Unmarshal(body, &commentInput); err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	validate := validator.New()
	if err := validate.Struct(commentInput); err != nil {
		responseMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	rawPostId := mux.Vars(r)["postId"]
	rawId := mux.Vars(r)["id"]
	id, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	postId, err := primitive.ObjectIDFromHex(rawPostId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	res, err := h.Collections.PostCommentsCollection.UpdateOne(r.Context(), bson.M{
		"_id":                postId,
		"comments._id":       id,
		"comments.author_id": user.ID,
	}, bson.M{
		"$set": bson.M{
			"comments.$.content": commentInput.Content,
		},
	})
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	if res.ModifiedCount == 0 {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	outBytes, err := json.Marshal(socketmodels.OutChangeMessage{
		Type:   "CHANGE",
		Method: "UPDATE",
		Entity: "POST_COMMENT",
		Data:   `{"ID":"` + rawId + `","content":"` + commentInput.Content + `","updated_at":"` + time.Now().Format(time.RFC3339) + `"}`,
	})

	h.SocketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
		Name: "post_page=" + rawPostId,
		Data: outBytes,
	}

	responseMessage(w, http.StatusOK, "Comment updated")
}

func (h handler) GetNewestPosts(w http.ResponseWriter, r *http.Request) {
	findOptions := options.Find()
	findOptions.SetLimit(15)
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

	var posts []models.Post
	cursor, err := h.Collections.PostCollection.Find(r.Context(), bson.M{"image_pending": false}, findOptions)
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	for cursor.Next(r.Context()) {
		var post models.Post
		err := cursor.Decode(&post)
		if err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

func (h handler) GetPage(w http.ResponseWriter, r *http.Request) {
	user, _, _ := helpers.GetUserAndSessionFromRequest(r, *h.Collections)

	pageNumberString := mux.Vars(r)["page"]
	pageNumber, err := strconv.Atoi(pageNumberString)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid page")
		return
	}
	pageSize := 20
	sortOrder := "DESC"
	term := ""
	sortMode := "DATE"
	if r.URL.Query().Has("mode") {
		sortMode = r.URL.Query().Get("mode")
	}
	if r.URL.Query().Has("order") {
		sortOrder = r.URL.Query().Get("order")
	}
	if r.URL.Query().Has("term") {
		term = r.URL.Query().Get("term")
	}

	tags := []string{}
	if r.URL.Query().Has("tags") {
		for _, v := range strings.Split(r.URL.Query().Get("tags"), " ") {
			if v != "" && v != " " {
				tags = append(tags, v)
			}
		}
	}

	findOptions := options.Find()
	findOptions.SetLimit(int64(pageSize))
	findOptions.SetSkip(int64(pageSize) * (int64(pageNumber) - 1))
	if sortOrder == "DESC" {
		if sortMode == "DATE" {
			findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
		}
		if sortMode == "POPULARITY" {
			findOptions.SetSort(bson.D{{Key: "sort_vote_count", Value: -1}})
		}
	}
	if sortOrder == "ASC" {
		if sortMode == "DATE" {
			findOptions.SetSort(bson.D{{Key: "created_at", Value: 1}})
		}
		if sortMode == "POPULARITY" {
			findOptions.SetSort(bson.D{{Key: "sort_vote_count", Value: 1}})
		}
	}

	filter := bson.M{"image_pending": false}
	if len(tags) == 0 {
		if term != "" && term != " " {
			filter = bson.M{"image_pending": false,
				"$text": bson.M{
					"$search":        term,
					"$caseSensitive": false,
				},
			}
		}
	} else {
		if term == "" {
			filter = bson.M{"tags": bson.M{"$in": tags}, "image_pending": false}
		} else {
			filter = bson.M{"tags": bson.M{"$in": tags}, "image_pending": false,
				"$text": bson.M{
					"$search":        term,
					"$caseSensitive": false,
				},
			}
		}
	}
	cursor, curerr := h.Collections.PostCollection.Find(r.Context(), filter, findOptions)
	defer cursor.Close(r.Context())
	if curerr != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	var posts []models.Post
	for cursor.Next(r.Context()) {
		var post models.Post
		err := cursor.Decode(&post)
		if err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
		var votes models.PostVotes
		if err := h.Collections.PostVoteCollection.FindOne(r.Context(), bson.M{"_id": post.ID}).Decode(&votes); err != nil {
			if err != mongo.ErrNoDocuments {
				responseMessage(w, http.StatusInternalServerError, "Internal error")
				return
			} else {
				responseMessage(w, http.StatusNotFound, "Not found")
				return
			}
		} else {
			var positiveVotes int = 0
			var negativeVotes int = 0
			for _, v := range votes.Votes {
				if user != nil {
					if user.ID != v.Uid {
						if v.IsUpvote == true {
							positiveVotes++
						} else {
							negativeVotes++
						}
					} else {
						post.UsersVote = v
					}
				} else {
					if v.IsUpvote == true {
						positiveVotes++
					} else {
						negativeVotes++
					}
				}
			}
			post.PositiveVoteCount = positiveVotes
			post.NegativeVoteCount = negativeVotes
		}
		posts = append(posts, post)
	}

	count, err := h.Collections.PostCollection.EstimatedDocumentCount(r.Context())
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	postBytes, err := json.Marshal(posts)

	out := map[string]string{
		"count": fmt.Sprint(count),
		"posts": string(postBytes),
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(out)
}

func (h handler) GetPost(w http.ResponseWriter, r *http.Request) {
	user, _, _ := helpers.GetUserAndSessionFromRequest(r, *h.Collections)

	slug := mux.Vars(r)["slug"]

	var post models.Post
	if err := h.Collections.PostCollection.FindOne(r.Context(), bson.M{"slug": slug}).Decode(&post); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	var commentsDoc models.PostComments
	if err := h.Collections.PostCommentsCollection.FindOne(r.Context(), bson.M{"_id": post.ID}).Decode(&commentsDoc); err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	} else {
		comments := commentsDoc.Comments
		for _, pcv := range commentsDoc.Votes {
			for i, pc := range comments {
				if pc.ID == pcv.CommentID {
					if user != nil && user.ID == pcv.Uid {
						comments[i].UsersVote.Uid = user.ID
						comments[i].UsersVote.IsUpvote = pcv.IsUpvote
					} else {
						if pcv.IsUpvote {
							comments[i].PositiveVoteCount++
						} else {
							comments[i].NegativeVoteCount++
						}
					}
				}
			}
		}
		post.Comments = comments
	}

	var votes models.PostVotes
	if err := h.Collections.PostVoteCollection.FindOne(r.Context(), bson.M{"_id": post.ID}).Decode(&votes); err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	} else {
		var positiveVotes int = 0
		var negativeVotes int = 0
		for _, v := range votes.Votes {
			if user != nil {
				if user.ID != v.Uid {
					if v.IsUpvote == true {
						positiveVotes++
					} else {
						negativeVotes++
					}
				} else {
					post.UsersVote = v
				}
			} else {
				if v.IsUpvote == true {
					positiveVotes++
				} else {
					negativeVotes++
				}
			}
		}
		post.PositiveVoteCount = positiveVotes
		post.NegativeVoteCount = negativeVotes
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

func (h handler) GetPostImage(w http.ResponseWriter, r *http.Request) {
	rawId := mux.Vars(r)["id"]
	id, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var postImage models.PostImage
	if err := h.Collections.PostImageCollection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&postImage); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(postImage.Binary.Data)))
	if _, err := w.Write(postImage.Binary.Data); err != nil {
		log.Println("Unable to write image to response")
	}
}

func (h handler) GetPostThumb(w http.ResponseWriter, r *http.Request) {
	rawId := mux.Vars(r)["id"]
	id, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var postThumb models.PostThumb
	if err := h.Collections.PostThumbCollection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&postThumb); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(postThumb.Binary.Data)))
	if _, err := w.Write(postThumb.Binary.Data); err != nil {
		log.Println("Unable to write image to response")
	}
}

func (h handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	slug := mux.Vars(r)["slug"]

	var post models.Post
	if h.Collections.PostCollection.FindOne(r.Context(), bson.M{"slug": slug}).Decode(&post); err != nil {
		responseMessage(w, http.StatusNotFound, "Post not found")
		return
	}

	if _, isProtected := h.ProtectedIDs.Pids[post.ID]; isProtected {
		responseMessage(w, http.StatusUnauthorized, "You cannot modify example posts")
		return
	}

	if post.Author != user.ID {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	var postInput validation.Post
	if json.Unmarshal(body, &postInput); err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	validate := validator.New()
	if err := validate.Struct(postInput); err != nil {
		responseMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	tags := make([]string, 0)
	for _, str := range strings.Split(postInput.Tags, "#") {
		str = strings.TrimSpace(str)
		if str != "" {
			tags = append(tags, str)
		}
	}

	tags = helpers.RemoveDuplicates(tags)

	result, err := h.Collections.PostCollection.UpdateByID(r.Context(), post.ID, bson.M{
		"$set": bson.M{
			"title":       postInput.Title,
			"description": postInput.Description,
			"body":        postInput.Body,
			"tags":        tags,
			"updated_at":  primitive.NewDateTimeFromTime(time.Now()),
		},
	})

	if result.MatchedCount == 0 {
		responseMessage(w, http.StatusNotFound, "Post not found")
		return
	}

	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	responseMessage(w, http.StatusOK, "Post updated")
}

func (h handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	var postInput validation.Post
	if json.Unmarshal(body, &postInput); err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	validate := validator.New()
	if err := validate.Struct(postInput); err != nil {
		responseMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	post := &models.Post{}

	re := regexp.MustCompile("[^a-z0-9]+")
	slug := re.ReplaceAllString(strings.ToLower(postInput.Title), "-")
	slug = strings.TrimFunc(slug, func(r rune) bool {
		runes := []rune("-")
		return r == runes[0]
	}) + "-" + cuid.Slug()

	tags := make([]string, 0)
	for _, str := range strings.Split(postInput.Tags, "#") {
		str = strings.TrimSpace(str)
		if str != "" {
			tags = append(tags, str)
		}
	}

	tags = helpers.RemoveDuplicates(tags)

	post.ID = primitive.NewObjectIDFromTimestamp(time.Now())
	post.Author = user.ID
	post.Slug = slug
	post.Title = postInput.Title
	post.Description = postInput.Description
	post.Body = postInput.Body
	post.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	post.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	post.ImagePending = true
	post.Tags = tags
	post.SortVoteCount = 0

	inserted, err := h.Collections.PostCollection.InsertOne(r.Context(), post)
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	postVotes := &models.PostVotes{
		ID:    inserted.InsertedID.(primitive.ObjectID),
		Votes: []models.PostVote{},
	}
	if _, err := h.Collections.PostVoteCollection.InsertOne(r.Context(), postVotes); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	postComments := &models.PostComments{
		ID:       inserted.InsertedID.(primitive.ObjectID),
		Votes:    []models.PostCommentVote{},
		Comments: []models.PostComment{},
	}
	if _, err := h.Collections.PostCommentsCollection.InsertOne(r.Context(), postComments); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(slug)
}

func (h handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]

	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var post models.Post
	if h.Collections.PostCollection.FindOne(r.Context(), bson.M{"slug": slug}).Decode(&post); err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		} else {
			responseMessage(w, http.StatusNotFound, "Post not found")
		}
		return
	}

	if _, isProtected := h.ProtectedIDs.Pids[post.ID]; isProtected {
		responseMessage(w, http.StatusUnauthorized, "You cannot delete example posts")
		return
	}

	if post.Author != user.ID {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if h.Collections.PostCollection.DeleteOne(r.Context(), bson.M{"slug": slug}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	responseMessage(w, http.StatusOK, "Post deleted")
}

func (h handler) UploadPostImage(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]

	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var post models.Post
	if h.Collections.PostCollection.FindOne(r.Context(), bson.M{"slug": slug}).Decode(&post); err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		} else {
			responseMessage(w, http.StatusNotFound, "Post not found")
		}
		return
	}

	if post.Author != user.ID {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	r.ParseMultipartForm(32 << 20) // copy pasted this thing << something to do with binary shift whatever that is. Is used here to define maximum memory in bytes.

	file, handler, err := r.FormFile("file")
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
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
	var thumbImg image.Image
	var blurImg image.Image
	var decodeErr error
	if isJPEG {
		img, decodeErr = jpeg.Decode(src)
	} else {
		img, decodeErr = png.Decode(src)
	}
	thumbImg = img
	blurImg = img
	if decodeErr != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	buf := &bytes.Buffer{}
	thumbBuf := &bytes.Buffer{}
	blurBuf := &bytes.Buffer{}
	width := img.Bounds().Dx()
	if width > 1024 {
		img = resize.Resize(1024, 0, img, resize.Lanczos3)
	} else {
		img = resize.Resize(uint(width), 0, img, resize.Lanczos2)
	}
	if width > 600 {
		thumbImg = resize.Resize(600, 0, thumbImg, resize.Lanczos3)
	} else {
		thumbImg = resize.Resize(uint(width/2), 0, thumbImg, resize.Lanczos3)
	}
	blurImg = resize.Resize(10, 0, thumbImg, resize.Lanczos3)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	if err := jpeg.Encode(thumbBuf, thumbImg, nil); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	if err := jpeg.Encode(blurBuf, blurImg, nil); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	thumbRes, err := h.Collections.PostThumbCollection.UpdateByID(r.Context(), post.ID, bson.M{"$set": bson.M{"binary": primitive.Binary{Data: buf.Bytes()}}})
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	if thumbRes.MatchedCount == 0 {
		_, err := h.Collections.PostThumbCollection.InsertOne(r.Context(), models.PostThumb{
			ID:     post.ID,
			Binary: primitive.Binary{Data: thumbBuf.Bytes()},
		})
		if err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	}

	imgRes, err := h.Collections.PostImageCollection.UpdateByID(r.Context(), post.ID, bson.M{"$set": bson.M{"binary": primitive.Binary{Data: buf.Bytes()}}})
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	if imgRes.MatchedCount == 0 {
		_, err := h.Collections.PostImageCollection.InsertOne(r.Context(), models.PostImage{
			ID:     post.ID,
			Binary: primitive.Binary{Data: buf.Bytes()},
		})
		if err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	}

	if h.Collections.PostCollection.UpdateByID(r.Context(), post.ID, bson.M{
		"$set": bson.M{
			"image_pending": false,
			"img_blur":      "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(blurBuf.Bytes()),
		},
	}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	responseMessage(w, http.StatusCreated, "Image uploaded")
}
