package messages

import "golang.org/x/image/math/f64"

type ConnectRequest struct {
	Username string `json:"username"`
}

type ConnectResponse struct {
	ClientId uint16 `json:"clientId"`
}

type ChatLoopMessage struct {
	ClientId uint16 `json:"clientId"`
	Message  string `json:"message"`
}

type GameLoopMessage struct {
	ClientId  uint16   `json:"clientId"`
	ClientPos f64.Vec2 `json:"clientPos"`
}
