package models

import "go.mongodb.org/mongo-driver/bson/primitive"

/*
	Private messages are kept in Inbox collection, and room messages are kept
	in RoomMessage collection because then when querying for Rooms or Users
	the messages aren't returned also, which is slower, also messages shouldn't
	trigger changestream events so they shouldn't be stored inside the same collection
	as posts or rooms.
*/

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Username  string             `bson:"username,maxlength=15" json:"username"`
	Password  string             `bson:"password" json:"-"`
	Base64pfp string             `bson:"-" json:"base64pfp,omitempty"`
}

type Inbox struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"ID"`
	Messages       []PrivateMessage     `bson:"messages" json:"messages"`
	MessagesSentTo []primitive.ObjectID `bson:"messages_sent_to" json:"-"` // list of all the people the user has messaged, needed to join both users messages together for display
}

type Pfp struct {
	ID     primitive.ObjectID `bson:"_id, omitempty"` //id should be the same id as the uid
	Binary primitive.Binary   `bson:"binary"`
}

type Session struct {
	UID       primitive.ObjectID `bson:"_uid"` // I dont know why i put an underscore here but it doesn't matter
	ExpiresAt primitive.DateTime `bson:"exp"`
}

type PrivateMessage struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Content       string             `bson:"content,maxlength=200" json:"content"`
	Uid           primitive.ObjectID `bson:"uid" json:"uid"`
	CreatedAt     primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt     primitive.DateTime `bson:"updated_at" json:"updated_at"`
	HasAttachment bool               `bson:"has_attachment" json:"has_attachment"`
	RecipientId   primitive.ObjectID `bson:"-" json:"recipient_id"`
}

type RoomMessage struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Content       string             `bson:"content,maxlength=200" json:"content"`
	Uid           primitive.ObjectID `bson:"uid" json:"uid"`
	CreatedAt     primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt     primitive.DateTime `bson:"updated_at" json:"updated_at"`
	HasAttachment bool               `bson:"has_attachment" json:"has_attachment"`
}

type AttachmentMetadata struct {
	ID       primitive.ObjectID `bson:"_id" json:"ID"` // Should be the same as message ID
	MimeType string             `bson:"mime_type" json:"mime_type"`
	Name     string             `bson:"name" json:"name"`
	Size     int                `bson:"size" json:"size"`
	Pending  bool               `bson:"pending"`
	Failed   bool               `bson:"failed"`
}

type AttachmentChunk struct {
	ID        primitive.ObjectID `bson:"_id"`     // The first chunk should be the same as the message ID
	NextChunk primitive.ObjectID `bson:"next_id"` // If its the last chunk NextChunk will be nil ObjectID (000000000000000000000000)
	Bytes     primitive.Binary   `bson:"bytes"`
}

type Post struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Author            primitive.ObjectID `bson:"author_id" json:"author_id"`
	CreatedAt         primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt         primitive.DateTime `bson:"updated_at" json:"updated_at"`
	Slug              string             `bson:"slug" json:"slug"`
	Title             string             `bson:"title" json:"title"`
	Description       string             `bson:"description" json:"description"`
	Body              string             `bson:"body" json:"body"`
	ImagePending      bool               `bson:"image_pending" json:"image_pending"`
	Tags              []string           `bson:"tags" json:"tags"`
	ImgBlur           string             `bson:"img_blur" json:"img_blur"`
	Comments          []PostComment      `bson:"-" json:"comments"`
	PositiveVoteCount int                `bson:"-" json:"vote_pos_count"`  // The vote count is sent to the client (excluding the users own vote)
	NegativeVoteCount int                `bson:"-" json:"vote_neg_count"`  // The vote count is sent to the client (excluding the users own vote)
	UsersVote         PostVote           `bson:"-" json:"my_vote"`         // The clients own vote is sent to the client... the client checks if uid of own vote is 0000000000000, to make sure that the client actually voted
	SortVoteCount     int                `bson:"sort_vote_count" json:"-"` // Used serverside when sorting by popularity. positive vote count - negative vote count.
}

type PostVotes struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Votes []PostVote         `bson:"votes" json:"votes"`
}

type PostVote struct {
	Uid      primitive.ObjectID `bson:"uid" json:"uid"`
	IsUpvote bool               `bson:"is_upvote" json:"is_upvote"`
}

type PostImage struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"` //should be the same as the posts id
	Binary primitive.Binary   `bson:"binary"`
}

type PostThumb struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"` //should be the same as the posts id
	Binary primitive.Binary   `bson:"binary"`
}

type PostComments struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Comments []PostComment      `bson:"comments" json:"comments"`
	Votes    []PostCommentVote  `bson:"votes" json:"-"`
}

type PostCommentVote struct {
	Uid       primitive.ObjectID `json:"uid" bson:"uid"`
	IsUpvote  bool               `bson:"is_upvote" json:"is_upvote"`
	CommentID primitive.ObjectID `bson:"comment_id" json:"comment_id"`
}

type PostComment struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Author            primitive.ObjectID `bson:"author_id" json:"author_id"`
	Content           string             `bson:"content" json:"content"`
	CreatedAt         primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt         primitive.DateTime `bson:"updated_at" json:"updated_at"`
	ParentID          string             `bson:"parent_id" json:"parent_id"`
	PositiveVoteCount int                `bson:"-" json:"vote_pos_count"` // The vote count is sent to the client (excluding the users own vote)
	NegativeVoteCount int                `bson:"-" json:"vote_neg_count"` // The vote count is sent to the client (excluding the users own vote)
	UsersVote         PostVote           `bson:"-" json:"my_vote"`        // The clients own vote is sent to the client... the client checks if uid of own vote is 0000000000000, to make sure that the client actually voted
}

type Room struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Name         string             `bson:"name,maxlength=24" json:"name"`
	Author       primitive.ObjectID `bson:"author_id" json:"author_id"`
	CreatedAt    primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt    primitive.DateTime `bson:"updated_at" json:"updated_at"`
	ImgBlur      string             `bson:"img_blur" json:"img_blur,omitempty"`
	Messages     []RoomMessage      `bson:"-" json:"messages"`
	ImagePending bool               `bson:"image_pending" json:"image_pending"`
}
type RoomMessages struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Messages []RoomMessage      `bson:"messages" json:"messages"`
}

type RoomImage struct {
	ID     primitive.ObjectID `bson:"_id, omitempty"` //should be the same as the rooms id
	Binary primitive.Binary   `bson:"binary"`
}
