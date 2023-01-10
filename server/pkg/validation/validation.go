package validation

type Credentials struct {
	Username string `json:"username" validate:"required,min=2,max=15"`
	Password string `json:"password" validate:"required,min=2,max=100"`
}

type Post struct {
	Title       string `json:"title" validate:"required,min=2,max=80"`
	Description string `json:"description" validate:"required,min=10,max=100"`
	Body        string `json:"body" validate:"required,min=10,max=8000"`
	Tags        string `json:"tags"`
}

type PostComment struct {
	Content string `json:"content" bson:"content"`
}
