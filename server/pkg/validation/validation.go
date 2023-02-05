package validation

/*
	The client also validates user input using Zod.
*/

type Credentials struct {
	Username string `json:"username" validate:"required,min=2,max=16"`
	Password string `json:"password" validate:"required,min=2,max=100"`
}

type Post struct {
	Title       string `json:"title" validate:"required,min=2,max=80"`
	Description string `json:"description" validate:"required,min=10,max=100"`
	Body        string `json:"body" validate:"required,min=10,max=8000"`
	Tags        string `json:"tags" validate:"required,min=2,max=100"`
}

type Vote struct {
	IsUpvote bool `json:"is_upvote"`
}

type Room struct {
	Name    string `json:"name" validate:"required,min=2,max=16"`
	Private bool   `json:"private" validate:"required"`
}

type PostComment struct {
	Content string `json:"content" validate:"required,min=1,max=200"`
}

type AttachmentMetadata struct {
	MimeType          string   `json:"type" validate:"required,min=1,max=40"`
	Name              string   `json:"name" validate:"required,min=1,max=500"`
	Size              int      `json:"size" validate:"required"`
	Length            float32  `json:"length"` // <- Video length, will be 0 for anything other than videos
	SubscriptionNames []string `json:"subscription_names"`
}
