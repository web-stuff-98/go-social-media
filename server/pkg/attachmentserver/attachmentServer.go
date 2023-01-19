package attachmentserver

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
	For attachment uploads
*/

type Upload struct {
	ChunksDone        int
	TotalChunks       int
	ChunkIDs          []primitive.ObjectID
	SubscriptionNames []string // Where to send progress updates to (inboxes / roomID)
}

type AttachmentServer struct {
	Uploaders map[primitive.ObjectID]map[primitive.ObjectID]Upload
}

func Init() (*AttachmentServer, error) {
	AttachmentServer := &AttachmentServer{
		Uploaders: make(map[primitive.ObjectID]map[primitive.ObjectID]Upload),
	}
	return AttachmentServer, nil
}
