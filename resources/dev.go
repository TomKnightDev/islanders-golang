package resources

// Developer tools, for testing. The server only honours them when started
// with -dev.

const MaxGameSpeed = 20

// DevCommandContents asks the server to change the game speed.
type DevCommandContents struct {
	Speed int `json:"speed"`
}

func NewDevCommandMessage(c DevCommandContents) *Message {
	return &Message{MessageType: DevCommandMessage, Contents: c}
}

// DevStateContents tells clients whether dev tools are on and the current
// game speed.
type DevStateContents struct {
	Enabled bool `json:"enabled"`
	Speed   int  `json:"speed"`
}

func NewDevStateMessage(c DevStateContents) *Message {
	return &Message{MessageType: DevStateMessage, Contents: c}
}
