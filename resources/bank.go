package resources

// Bank window commands, sent by the client while standing at a bank.
const (
	BankOpDeposit    = "deposit"    // backpack to bank
	BankOpWithdraw   = "withdraw"   // bank to backpack
	BankOpDepositAll = "depositAll" // everything in the backpack
)

// BankCommandContents moves Count of Item (only of Quality, if set), or
// the piece with ID, between the backpack and the bank.
type BankCommandContents struct {
	Op      string   `json:"op"`
	Item    ItemKind `json:"item"`
	Quality int      `json:"quality"`
	ID      uint32   `json:"id"`
	Count   int      `json:"count"`
}

// CombineCommandContents merges every stack of a material into as few
// slots as it fits in, at their average quality: in the backpack anywhere,
// or in the bank while standing at one.
type CombineCommandContents struct {
	Item ItemKind `json:"item"`
	Bank bool     `json:"bank"`
}

func NewCombineCommandMessage(c CombineCommandContents) *Message {
	return &Message{MessageType: CombineCommandMessage, Contents: c}
}

func NewBankCommandMessage(c BankCommandContents) *Message {
	return &Message{MessageType: BankCommandMessage, Contents: c}
}
