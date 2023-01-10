package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Username  string             `bson:"username,maxlength=15" json:"username" validate:"required,min=2,max=15"`
	Password  string             `bson:"password" json:"-" validate:"required,min=2,max=100"`
	Base64pfp string             `bson:"-" json:"base64pfp,omitempty"`
}

type Inbox struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Messages []PrivateMessage   `bson:"messages" json:"messages"`
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
}

type Pfp struct {
	ID     primitive.ObjectID `bson:"_id, omitempty"` //id should be the same id as the uid
	Binary primitive.Binary   `bson:"binary"`
}

type Session struct {
	UID       primitive.ObjectID `bson:"_uid"`
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

//socket message JSON from the client
type MessageEvent struct {
	Content       string `json:"content"`
	HasAttachment bool   `json:"has_attachment"`
}

type Room struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"ID"` // omitempty to protect against zeroed _id insertion
	Name      string             `bson:"name,maxlength=24" json:"name"`
	Author    primitive.ObjectID `bson:"author_id" json:"author_id"`
	CreatedAt primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt primitive.DateTime `bson:"updated_at" json:"updated_at"`
	Messages  []RoomMessage      `bson:"messages" json:"messages"`
	ImgBlur   string             `bson:"img_blur" json:"img_blur,omitempty"`
}

type RoomImage struct {
	ID     primitive.ObjectID `bson:"_id, omitempty"` //should be the same as the rooms id
	Binary primitive.Binary   `bson:"binary"`
}

type Attachment struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"ID"` // ID should be the same as the message
	Binary   primitive.Binary   `bson:"binary"`
	MimeType string             `bson:"attachment_type" json:"-"`
}

type Post struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"ID"`
	Author       primitive.ObjectID `bson:"author_id" json:"author_id"`
	CreatedAt    primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt    primitive.DateTime `bson:"updated_at" json:"updated_at"`
	Slug         string             `bson:"slug" json:"slug"`
	Title        string             `bson:"title" json:"title" validate:"required,min=2,max=80"`
	Description  string             `bson:"description" json:"description" validate:"required,min=10,max=100"`
	Body         string             `bson:"body" json:"body" validate:"required,min=10,max=8000"`
	ImagePending bool               `bson:"image_pending" json:"image_pending"`
	Tags         []string           `bson:"tags" json:"tags"`
	ImgBlur      string             `bson:"img_blur" json:"img_blur"`
	Comments     []PostComment      `bson:"-" json:"comments"`
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
	Content   string             `bson:"content" json:"content" valdate:"required,min=1,max=200"`
	CreatedAt primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt primitive.DateTime `bson:"updated_at" json:"updated_at"`
	ParentID  string             `bson:"parent_id" json:"parent_id"`
}
