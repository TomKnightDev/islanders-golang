package resources

// Bank window commands, sent by the client while standing at a bank.
const (
	BankOpDeposit    = "deposit"    // backpack to bank
	BankOpWithdraw   = "withdraw"   // bank to backpack
	BankOpDepositAll = "depositAll" // everything in the backpack
)

// BankCommandContents moves Count of Item, or the piece with ID, between
// the backpack and the bank.
type BankCommandContents struct {
	Op    string   `json:"op"`
	Item  ItemKind `json:"item"`
	ID    uint32   `json:"id"`
	Count int      `json:"count"`
}

func NewBankCommandMessage(c BankCommandContents) *Message {
	return &Message{MessageType: BankCommandMessage, Contents: c}
}
