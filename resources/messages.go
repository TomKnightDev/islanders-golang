package resources

import (
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
	MoveRequestMessage        MessageType = "moveRequestMessage"
	InventoryMessage          MessageType = "inventoryMessage"
	ObjectUpdateMessage       MessageType = "objectUpdateMessage"
	ActionCommandMessage      MessageType = "actionCommandMessage"
	EquipCommandMessage       MessageType = "equipCommandMessage"
	PathMessage               MessageType = "pathMessage"
	ChooseClassMessage        MessageType = "chooseClassMessage"     // server: a new name; pick a class
	CreateCharacterMessage    MessageType = "createCharacterMessage" // client: the class picked
	FightMessage              MessageType = "fightMessage"           // server: a fight the player is in
	ShopCommandMessage        MessageType = "shopCommandMessage"     // client: view a store or buy from it
	StoreMessage              MessageType = "storeMessage"           // server: a village store's prices
	DevCommandMessage         MessageType = "devCommandMessage"      // client: change the game speed (dev servers only)
	DevStateMessage           MessageType = "devStateMessage"        // server: dev tools on or off, and the game speed
	BankCommandMessage        MessageType = "bankCommandMessage"     // client: move items at the bank
	InnMessage                MessageType = "innMessage"             // server: quests and rumours at an inn
	StationCommandMessage     MessageType = "stationCommandMessage"  // client: make or improve gear at an anvil or the fletcher's hut
	RepairCommandMessage      MessageType = "repairCommandMessage"   // client: repair a worn piece
	CombineCommandMessage     MessageType = "combineCommandMessage"  // client: combine a material's stacks
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

type ConnectResponseContents struct {
	ClientId uint16   `json:"clientId"`
	Pos      f64.Vec2 `json:"pos"`
	Tile     f64.Vec2 `json:"tile"`
	WorldMap WorldMap `json:"world"`
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

func NewChatMessage(clientId uint16, message string) *Message {
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
	Health       int      `json:"health"`
	MaxHealth    int      `json:"maxHealth"`
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
	EntityId  uint16   `json:"entityId"`
	Pos       f64.Vec2 `json:"pos"`
	Tile      f64.Vec2 `json:"tile"`
	Username  string   `json:"username"`
	Health    int      `json:"health"`
	MaxHealth int      `json:"maxHealth"`
	Gone      bool     `json:"gone"` // removed from the map, e.g. a defeated enemy
}

func NewServerEntityUpdateMessage(contents []ServerEntityUpdateContents) *Message {
	m := &Message{
		MessageType: ServerEntityUpdateMessage,
		Contents:    contents,
	}

	return m
}

// MoveRequestContents asks the server to walk the player to a tile. If the
// tile is blocked (a rock, a building) the player walks next to it instead.
type MoveRequestContents struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func NewMoveRequestMessage(clientId uint16, contents MoveRequestContents) *Message {
	m := &Message{
		MessageType: MoveRequestMessage,
		Contents:    contents,
		ClientId:    clientId,
	}

	return m
}

// InventoryContents is sent to a player whenever their backpack, bank
// account, current activity, found places or action list changes.
type InventoryContents struct {
	Backpack Inventory `json:"backpack"`
	// At is the building the player is standing at with the action list
	// stopped, if any, in village AtVillage; its window opens. Bank is only
	// filled in at a bank.
	At        Building  `json:"at"`
	AtVillage string    `json:"atVillage"`
	Bank      Inventory `json:"bank"`
	Activity  string    `json:"activity"`
	// ActionSeconds is how long the current action takes, or 0 if it has no
	// set length. ActionElapsed is how far into it the player already is.
	ActionSeconds float64 `json:"actionSeconds"`
	ActionElapsed float64 `json:"actionElapsed"`

	Found   []string       `json:"found"`   // names of places the player has found
	Actions []QueuedAction `json:"actions"` // the action list; [0] runs first
	Running bool           `json:"running"` // whether the list is being worked through
	Skills  []SkillState   `json:"skills"`

	Health     int            `json:"health"`
	MaxHealth  int            `json:"maxHealth"`
	Stamina    int            `json:"stamina"`
	Coins      int            `json:"coins"`
	MaxStamina int            `json:"maxStamina"`
	Equipment  []EquippedItem `json:"equipment"`
	Class      ClassID        `json:"class"` // which armour the player can wear

	// At an anvil or the fletcher's hut: Making is the item being made
	// (with MakeLeft still to make), Improving the piece being improved (0
	// if none), and StationNote how the last step went or why work stopped.
	Making      ItemKind `json:"making"`
	MakeLeft    int      `json:"makeLeft"`
	Improving   uint32   `json:"improving"`
	StationNote string   `json:"stationNote"`
}

func NewInventoryMessage(contents InventoryContents) *Message {
	return &Message{
		MessageType: InventoryMessage,
		Contents:    contents,
	}
}

// ObjectUpdateContents changes one tile on the object layer, such as a rock
// becoming depleted or respawning.
type ObjectUpdateContents struct {
	X    int `json:"x"`
	Y    int `json:"y"`
	Tile int `json:"tile"`
}

func NewObjectUpdateMessage(contents ObjectUpdateContents) *Message {
	return &Message{
		MessageType: ObjectUpdateMessage,
		Contents:    contents,
	}
}

// PathContents is the route a player is walking, sent only to that player so
// their client can draw it. Tiles are in walking order; empty means stopped.
type PathContents struct {
	Tiles [][2]int `json:"tiles"`
}

func NewPathMessage(c PathContents) *Message {
	return &Message{MessageType: PathMessage, Contents: c}
}

// Who made an attack in a FightContents event.
const (
	AttackerPlayer = "player"
	AttackerEnemy  = "enemy"
)

// FightContents tells a player about the fight they're in, for the fight
// panel. It's sent when a fight starts (Attacker empty), after every attack
// by either side, and with Active false when the fight ends. Only the
// fighter receives it.
type FightContents struct {
	Active          bool   `json:"active"`
	EnemyName       string `json:"enemyName"`
	EnemySprite     [2]int `json:"enemySprite"` // pixel position in characters.png
	EnemyHealth     int    `json:"enemyHealth"`
	EnemyMaxHealth  int    `json:"enemyMaxHealth"`
	PlayerHealth    int    `json:"playerHealth"`
	PlayerMaxHealth int    `json:"playerMaxHealth"`
	Attacker        string `json:"attacker"` // AttackerPlayer, AttackerEnemy, or empty
	Hit             bool   `json:"hit"`
	Damage          int    `json:"damage"`
}

func NewFightMessage(c FightContents) *Message {
	return &Message{MessageType: FightMessage, Contents: c}
}
