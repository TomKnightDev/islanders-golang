package messages

import "golang.org/x/image/math/f64"

// type ConnectRequest struct {
// 	Username string `json:"username"`
// }

// type ConnectResponse struct {
// 	ClientId uint16 `json:"clientId"`
// }

// type ChatLoopMessage struct {
// 	ClientId uint16 `json:"clientId"`
// 	Message  string `json:"message"`
// }

// type GameLoopMessage struct {
// 	EntityMessages []EntityMessage `json:"entityMessages"`
// }

// type EntityMessage struct {
// 	EntityId   uint16   `json:"entityId"`
// 	EntityPos  f64.Vec2 `json:"entityPos"`
// 	EntityTile f64.Vec2 `json:"entityTile"`
// }

type MessageType string

const (
	ConnectRequestMessage     MessageType = "connectRequestMessage"
	ConnectResponseMessage    MessageType = "connectResponseMessage"
	FailedToConnectMessage    MessageType = "failedToConnectMessage"
	ChatMessage               MessageType = "chatMessage"
	UpdateMessage             MessageType = "updateMessage"
	ServerEntityUpdateMessage MessageType = "serverEntityUpdateMessage"
)

type Message struct {
	MessageType MessageType `json:"messageType"`
	Contents    interface{} `json:"contents"`
	ClientId    uint16      `json:"clientId"`
}

type ConnectRequestContents struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewConnectRequestMessage(connectContents ConnectRequestContents) *Message {
	m := &Message{
		MessageType: ConnectRequestMessage,
		Contents:    connectContents,
	}
	return m
}

func NewConnectResponseMessage(clientId uint16) *Message {
	m := &Message{
		MessageType: ConnectResponseMessage,
		Contents:    clientId,
	}

	return m
}

func NewFailedToConnectMessage(failedMessage error) *Message {
	m := &Message{
		MessageType: FailedToConnectMessage,
		Contents:    failedMessage,
	}

	return m
}

func NewChatMessage(clientId uint16, message string) *Message {
	m := &Message{
		MessageType: ChatMessage,
		Contents:    message,
		ClientId:    clientId,
	}

	return m
}

type UpdateContents struct {
	Pos  f64.Vec2 `json:"pos"`
	Tile f64.Vec2 `json:"tile"`
}

func NewUpdateMessage(clientId uint16, contents UpdateContents) *Message {
	m := &Message{
		MessageType: UpdateMessage,
		Contents:    contents,
		ClientId:    clientId,
	}

	return m
}

type ServerEntityUpdateContents struct {
	EntityId uint16   `json:"entityId"`
	Pos      f64.Vec2 `json:"pos"`
	Tile     f64.Vec2 `json:"tile"`
}

func NewServerEntityUpdateMessage(contents []ServerEntityUpdateContents) *Message {
	m := &Message{
		MessageType: ServerEntityUpdateMessage,
		Contents:    contents,
	}

	return m
}
