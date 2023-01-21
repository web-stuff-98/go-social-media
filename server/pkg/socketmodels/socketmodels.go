package socketmodels

/*
	Models for messages sent through the websocket, encoded into []bytes from json marshal
*/

// TYPE: PRIVATE_MESSAGE (in)
type PrivateMessage struct {
	Type          string `json:"TYPE"`
	RecipientId   string `json:"recipient_id"`
	Content       string `json:"content"`
	HasAttachment bool   `json:"has_attachment"`
}

// TYPE: ROOM_MESSAGE (in)
type RoomMessage struct {
	Type          string `json:"TYPE"`
	RoomId        string `json:"room_id"`
	Content       string `json:"content"`
	HasAttachment bool   `json:"has_attachment"`
}

// TYPE: OPEN_SUBSCRIPTION/CLOSE_SUBSCRIPTION
type OpenCloseSubscription struct {
	Name string `json:"name"`
}

// TYPE: OPEN_SUBSCRIPTIONS
type OpenCloseSubscriptions struct {
	Names []string `json:"names"`
}

// TYPE: ROOM_MESSAGE/PRIVATE_MESSAGE/POST_VOTE/POST_COMMENT_VOTE/ATTACHMENT_PROGRESS/ATTACHMENT_COMPLETE/RESPONSE_MESSAGE
type OutMessage struct {
	Type string `json:"TYPE"`
	Data string `json:"DATA"`
}

// TYPE: CHANGE
type OutChangeMessage struct {
	Type   string `json:"TYPE"`
	Method string `json:"METHOD"`
	Data   string `json:"DATA"`
	Entity string `json:"ENTITY"`
}

/*
	Below are models for WebRTC chat events (server to client events)
*/

// TYPE: VID_USER_JOINED
type OutVidChatUserJoined struct {
	SignalJSON string `json:"signal_json"`
	CallerUID  string `json:"caller_uid"`
}

// TYPE: VID_USER_LEFT
type OutVidChatUserLeft struct {
	UID string `json:"uid"`
}

// TYPE: VID_SENDING_SIGNAL_OUT
type OutVidChatSendingSignal struct {
	SignalJSON   string `json:"signal_json"`
	CallerUID    string `json:"caller_uid"`
	UserToSignal string `json:"user_to_signal"`
}

// TYPE: VID_RETURNING_SIGNAL_OUT
type OutVidChatReturningSignal struct {
	SignalJSON string `json:"signal_json"`
	CallerUID  string `json:"caller_uid"`
}

// TYPE: VID_RECEIVING_RETURNED_SIGNAL
type OutVidChatReceivingReturnedSignal struct {
	SignalJSON string `json:"signal_json"`
	UID        string `json:"uid"`
}

// TYPE: VID_ALL_USERS
type OutVidChatAllUsers struct {
	UIDs []string `json:"uids"`
}

/*
	Below are models for WebRTC chat events (client to server events)
*/

// TYPE: VID_SENDING_SIGNAL_IN
type InVidChatSendingSignal struct {
	Type         string `json:"TYPE"`
	SignalJSON   string `json:"signal_json"`
	UserToSignal string `json:"user_to_signal"`
}

// TYPE: VID_RETURNING_SIGNAL_IN
type InVidChatReturningSignal struct {
	Type       string `json:"TYPE"`
	SignalJSON string `json:"signal_json"`
	CallerUID  string `json:"caller_uid"`
}

// TYPE: VID_JOIN
type InVidChatJoin struct {
	Type   string `json:"TYPE"`
	JoinID string `json:"join_id"` // Can be either a room ID or another user ID
	IsRoom bool   `json:"is_room"`
}
