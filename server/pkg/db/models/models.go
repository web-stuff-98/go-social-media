package models

import "go.mongodb.org/mongo-driver/bson/primitive"

/*
	Private messages are kept in Inbox collection, and room messages are kept
	in RoomMessage collection because then when querying for Rooms or Users
	the messages aren't returned also, which is slower.
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

type PrivateMessage struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"ID"` // omitempty to protect against zeroed _id insertion
	Content           string             `bson:"content,maxlength=200" json:"content"`
	Uid               primitive.ObjectID `bson:"uid" json:"uid"`
	CreatedAt         primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt         primitive.DateTime `bson:"updated_at" json:"updated_at"`
	HasAttachment     bool               `bson:"has_attachment" json:"has_attachment"`
	AttachmentPending bool               `bson:"attachment_pending" json:"attachment_pending"`
	AttachmentType    string             `bson:"attachment_type" json:"attachment_type"`
	AttachmentError   bool               `bson:"attachment_error" json:"attachment_error"`
	RecipientId       primitive.ObjectID `bson:"-" json:"recipient_id"`
}

type Pfp struct {
	ID     primitive.ObjectID `bson:"_id, omitempty"` //id should be the same id as the uid
	Binary primitive.Binary   `bson:"binary"`
}

type Session struct {
	UID       primitive.ObjectID `bson:"_uid"` // I dont know why i put an underscore here but it doesn't matter
	ExpiresAt primitive.DateTime `bson:"exp"`
}

type RoomMessage struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"ID"` // omitempty to protect against zeroed _id insertion
	Content           string             `bson:"content,maxlength=200" json:"content"`
	Uid               primitive.ObjectID `bson:"uid" json:"uid"`
	CreatedAt         primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt         primitive.DateTime `bson:"updated_at" json:"updated_at"`
	HasAttachment     bool               `bson:"has_attachment" json:"has_attachment"`
	AttachmentPending bool               `bson:"attachment_pending" json:"attachment_pending"`
	AttachmentType    string             `bson:"attachment_type" json:"attachment_type"`
	AttachmentError   bool               `bson:"attachment_error" json:"attachment_error"`
}

type Attachment struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"ID"` // ID should be the same as the message
	Binary   primitive.Binary   `bson:"binary"`
	MimeType string             `bson:"attachment_type" json:"-"`
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
	NegativeVoteCount int                `bson:"-" json:"vote_pos_count"` // The vote count is sent to the client (excluding the users own vote)
	PositiveVoteCount int                `bson:"-" json:"vote_neg_count"` // The vote count is sent to the client (excluding the users own vote)
	UsersVote         PostVote           `bson:"-" json:"my_vote"`        // The clients own vote is sent to the client
}

type PostVotes struct {
	ID    primitive.ObjectID `bson:"id,omitempty" json:"id"`
	Votes []PostVote         `bson:"votes" json:"votes"`
}

type PostVote struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
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
}

type PostComment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Author    primitive.ObjectID `bson:"author_id" json:"author_id"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt primitive.DateTime `bson:"updated_at" json:"updated_at"`
	ParentID  string             `bson:"parent_id" json:"parent_id"`
}

type Room struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"ID"` // omitempty to protect against zeroed _id insertion
	Name      string             `bson:"name,maxlength=24" json:"name"`
	Author    primitive.ObjectID `bson:"author_id" json:"author_id"`
	CreatedAt primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt primitive.DateTime `bson:"updated_at" json:"updated_at"`
	ImgBlur   string             `bson:"img_blur" json:"img_blur,omitempty"`
	Messages  []PrivateMessage   `bson:"-" json:"messages"`
}
type RoomMessages struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Messages []PrivateMessage   `bson:"messages" json:"messages"`
}

type RoomImage struct {
	ID     primitive.ObjectID `bson:"_id, omitempty"` //should be the same as the rooms id
	Binary primitive.Binary   `bson:"binary"`
}
