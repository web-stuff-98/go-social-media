package socketmodels

/*
	Models for messages sent through the websocket, encoded into []bytes from json marshal
*/

type InMessage struct {
	Content string `json:"content"`
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
