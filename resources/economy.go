package resources

// Economy: coins, item values and the village general stores. See
// docs/ECONOMY.md.

const (
	StartingCoins = 50
	InnFee        = 5 // coins per rest; a character who can't pay sleeps in the stable

	// StoreBuysAt is the share of an item's value a store pays at full demand.
	StoreBuysAt = 0.4
)

// Values is every item's base value in coins. Store prices are worked out
// from it.
var Values = map[ItemKind]int{
	ItemCopperOre: 2, ItemIronOre: 5, ItemGoldOre: 12, ItemMithrilOre: 30,
	ItemCopperBar: 6, ItemIronIngot: 15, ItemGoldBar: 36, ItemMithrilBar: 90,
	ItemOakLogs: 2, ItemPineLogs: 5, ItemMapleLogs: 12, ItemDarkOakLogs: 30,
	ItemOakArrows: 1, ItemPineArrows: 2, ItemMapleArrows: 4, ItemDarkOakArrows: 8,
	ItemCabbage: 3, ItemCotton: 4,

	ItemTrainingSword: 10,
	ItemCopperSword:   40, ItemIronSword: 120, ItemGoldSword: 350, ItemMithrilSword: 1000,
	ItemOakShortbow: 40, ItemPineShortbow: 120, ItemMapleShortbow: 350, ItemDarkOakShortbow: 1000,
	ItemApprenticeStaff: 40, ItemAdeptStaff: 120, ItemBattleStaff: 350, ItemArchmageStaff: 1000,
	ItemBoneStaff: 40, ItemGraveStaff: 120, ItemWraithStaff: 350, ItemLichStaff: 1000,
	ItemCopperHelmet: 30, ItemIronHelmet: 90, ItemGoldHelmet: 250, ItemMithrilHelmet: 700,
	ItemCopperBody: 60, ItemIronBody: 180, ItemGoldBody: 500, ItemMithrilBody: 1400,
	ItemCopperShield: 35, ItemIronShield: 110, ItemGoldShield: 300, ItemMithrilShield: 850,
}

// StoreItem is something every general store sells, at a markup on value.
type StoreItem struct {
	Item   ItemKind
	Markup float64
}

// StoreStock is what the general stores sell. Prices are high on purpose:
// other players should usually be the cheaper source.
var StoreStock = []StoreItem{
	{ItemCabbage, 1.5},
	{ItemCopperOre, 2}, {ItemOakLogs, 2}, {ItemCopperBar, 2},
	{ItemOakArrows, 1.5},
	{ItemCopperSword, 1.5}, {ItemOakShortbow, 1.5}, {ItemApprenticeStaff, 1.5}, {ItemBoneStaff, 1.5},
	{ItemCopperHelmet, 1.5}, {ItemCopperBody, 1.5}, {ItemCopperShield, 1.5},
}

func init() {
	// Copper tools too, so a lost or outgrown tool can always be replaced.
	for _, slot := range ToolSlots {
		StoreStock = append(StoreStock, StoreItem{ToolItem("Copper", slot), 1.5})
	}
}

// StorePrice is what a store charges for an item.
func StorePrice(s StoreItem) int {
	return max(1, int(float64(Values[s.Item])*s.Markup+0.5))
}

// Sellable reports whether stores will buy an item.
func Sellable(k ItemKind) bool { return Values[k] > 0 }

// Shop window commands, sent by the client while standing at the store.
// Purchases go into the backpack; sales come out of it.
const (
	ShopOpView = "view" // ask for the store's prices
	ShopOpBuy  = "buy"
	ShopOpSell = "sell"
)

type ShopCommandContents struct {
	Op      string   `json:"op"`
	Village string   `json:"village"`
	Item    ItemKind `json:"item"`
	ID      uint32   `json:"id"` // sell: a piece of gear or a tool
	Count   int      `json:"count"`
}

func NewShopCommandMessage(c ShopCommandContents) *Message {
	return &Message{MessageType: ShopCommandMessage, Contents: c}
}

// StorePrices is one item at a village store: what it costs to buy there
// (0 if the store doesn't sell it) and what the store pays right now for
// one at the quality the player carries (or, for gear, for the piece ID).
type StorePrices struct {
	Item ItemKind `json:"item"`
	ID   uint32   `json:"id,omitempty"`
	Buy  int      `json:"buy"`
	Pays int      `json:"pays"`
}

// StoreContents is a village store's prices, sent when the player looks.
type StoreContents struct {
	Village string        `json:"village"`
	Prices  []StorePrices `json:"prices"`
}

func NewStoreMessage(c StoreContents) *Message {
	return &Message{MessageType: StoreMessage, Contents: c}
}
