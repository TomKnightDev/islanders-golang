package resources

import (
	"github.com/google/uuid"
	"golang.org/x/image/math/f64"
)

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
	ClientId    uuid.UUID   `json:"clientId"`
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

type ConnectResponseContents struct {
	ClientId uuid.UUID `json:"clientId"`
	Pos      f64.Vec2  `json:"pos"`
	Tile     f64.Vec2  `json:"tile"`
	WorldMap WorldMap  `json:"world"`
}

func NewConnectResponseMessage(contents ConnectResponseContents) *Message {
	m := &Message{
		MessageType: ConnectResponseMessage,
		Contents:    contents,
	}

	return m
}

func NewFailedToConnectMessage(failedMessage string) *Message {
	m := &Message{
		MessageType: FailedToConnectMessage,
		Contents:    failedMessage,
	}

	return m
}

func NewChatMessage(clientId uuid.UUID, message string) *Message {
	m := &Message{
		MessageType: ChatMessage,
		Contents:    message,
		ClientId:    clientId,
	}

	return m
}

type UpdateContents struct {
	Pos          f64.Vec2 `json:"pos"`
	Tile         f64.Vec2 `json:"tile"`
	Disconnected bool     `json:"disconnected"`
	Username     string   `json:"username"`
}

func NewUpdateMessage(clientId uuid.UUID, contents UpdateContents) *Message {
	m := &Message{
		MessageType: UpdateMessage,
		Contents:    contents,
		ClientId:    clientId,
	}

	return m
}

type ServerEntityUpdateContents struct {
	EntityId uuid.UUID `json:"entityId"`
	Pos      f64.Vec2  `json:"pos"`
	Tile     f64.Vec2  `json:"tile"`
}

func NewServerEntityUpdateMessage(contents []ServerEntityUpdateContents) *Message {
	m := &Message{
		MessageType: ServerEntityUpdateMessage,
		Contents:    contents,
	}

	return m
}
