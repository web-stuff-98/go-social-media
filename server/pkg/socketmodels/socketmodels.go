package socketmodels

/*
	Models for messages sent through the websocket, encoded into []bytes from json marshal
*/

type PrivateMessage struct {
	Type          string `json:"TYPE"`
	RecipientId   string `json:"recipient_id"`
	Content       string `json:"content"`
	HasAttachment bool   `json:"has_attachment"`
}

type RoomMessage struct {
	Type          string `json:"TYPE"`
	RoomId        string `json:"room_id"`
	Content       string `json:"content"`
	HasAttachment bool   `json:"has_attachment"`
}

type OpenCloseSubscription struct {
	Name string `json:"name"`
}

type OpenCloseSubscriptions struct {
	Names []string `json:"names"`
}

type OutMessage struct {
	Type string `json:"TYPE"`
	Data string `json:"DATA"`
}

type OutChangeMessage struct {
	Type   string `json:"TYPE"`
	Method string `json:"METHOD"`
	Data   string `json:"DATA"`
	Entity string `json:"ENTITY"`
}
