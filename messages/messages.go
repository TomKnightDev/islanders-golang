package messages

type ConnectRequest struct {
	Username string `json:"username"`
}

type ConnectResponse struct {
	ClientId uint16 `json:"clientId"`
}

type GameLoopMessage struct {
	ClientId uint16 `json:"clientId"`
	Message  string `json:"message"`
}
