package resources

type ActionKind string

const (
	ActionMine  ActionKind = "mine"  // mine iron at a mine, banking it
	ActionSmelt ActionKind = "smelt" // smelt banked ore into bars at a village blacksmith
)

// ActionNames is how each action reads in the UI.
var ActionNames = map[ActionKind]string{
	ActionMine:  "Mine iron",
	ActionSmelt: "Smelt iron",
}

// ActionPlaceKind is the kind of place each action happens at.
var ActionPlaceKind = map[ActionKind]PlaceKind{
	ActionMine:  PlaceMine,
	ActionSmelt: PlaceVillage,
}

type GoalKind string

const (
	GoalAmount GoalKind = "amount" // until the bank holds Amount of the action's product
	GoalTime   GoalKind = "time"   // for Minutes of the character's time
)

const (
	MaxActions    = 5
	MaxGoalAmount = 10000
	MaxGoalMins   = 24 * 60
)

// ActionSpec is one entry in a player's action list.
type ActionSpec struct {
	Kind    ActionKind `json:"kind"`
	Place   string     `json:"place"`
	Goal    GoalKind   `json:"goal"`
	Amount  int        `json:"amount"`
	Minutes int        `json:"minutes"`
}

// Product is the item an action's amount goal counts in the bank.
func (a ActionSpec) Product() ItemKind {
	if a.Kind == ActionSmelt {
		return ItemIronIngot
	}
	return ItemIronOre
}

// QueuedAction is an entry in the action list as the server reports it.
type QueuedAction struct {
	ActionSpec
	// For the action that's running: how many of the product are banked,
	// and for a time goal, the seconds left.
	Banked      int     `json:"banked"`
	SecondsLeft float64 `json:"secondsLeft"`
}

// Action list commands, sent by the client.
const (
	ActionOpAdd    = "add"
	ActionOpEdit   = "edit"
	ActionOpRemove = "remove"
	ActionOpUp     = "up"
	ActionOpDown   = "down"
	ActionOpStart  = "start"
	ActionOpStop   = "stop"
)

type ActionCommandContents struct {
	Op     string     `json:"op"`
	Index  int        `json:"index"`
	Action ActionSpec `json:"action"`
}

func NewActionCommandMessage(contents ActionCommandContents) *Message {
	return &Message{
		MessageType: ActionCommandMessage,
		Contents:    contents,
	}
}
