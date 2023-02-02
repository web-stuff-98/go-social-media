package socketmodels

/*
	Models for messages sent through the websocket, encoded into []bytes from json marshal

	When a socket message is sent out the "event type" is keyed as TYPE, when a socket message
	is recieved on the server it should be keyed as event_type, this is just so that its a bit
	easier to tell which models for sending data out, and which are for receiving data from
	the client.
*/

// TYPE: OPEN_CONV & EXIT_CONV
type OpenExitConv struct {
	Uid string `json:"uid"`
}

// TYPE: PRIVATE_MESSAGE (type needs to be TYPE, not event_type, because this goes OUT not in)
type PrivateMessage struct {
	Type          string `json:"TYPE"`
	RecipientId   string `json:"recipient_id"`
	Content       string `json:"content"`
	HasAttachment bool   `json:"has_attachment"`
}

// TYPE: ROOM_MESSAGE (type needs to be TYPE, not event_type, because this goes OUT not in)
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

// TYPE: ROOM_MESSAGE/PRIVATE_MESSAGE/POST_VOTE/POST_COMMENT_VOTE/ATTACHMENT_PROGRESS/ATTACHMENT_COMPLETE/RESPONSE_MESSAGE/NOTIFICATIONS
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
	SignalJSON   string `json:"signal_json"`
	UserToSignal string `json:"user_to_signal"`
}

// TYPE: VID_RETURNING_SIGNAL_IN
type InVidChatReturningSignal struct {
	SignalJSON string `json:"signal_json"`
	CallerUID  string `json:"caller_uid"`
}

// TYPE: VID_JOIN
type InVidChatJoin struct {
	JoinID string `json:"join_id"` // Can be either a room ID or another user ID
	IsRoom bool   `json:"is_room"`
}

// TYPE: VID_LEAVE
type InVidChatLeave struct {
	ID     string `json:"id"` // Can be either a room ID or another user ID
	IsRoom bool   `json:"is_room"`
}
