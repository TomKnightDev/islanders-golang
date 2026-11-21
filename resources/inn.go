package resources

// InnContents is what's on offer at an inn: quests (none yet) and rumours
// of places the player hasn't found.
type InnContents struct {
	Village string   `json:"village"`
	Rumours []string `json:"rumours"`
}

func NewInnMessage(c InnContents) *Message {
	return &Message{MessageType: InnMessage, Contents: c}
}
