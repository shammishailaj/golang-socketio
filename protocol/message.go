package protocol

const (
	// Message with connection options
	MessageTypeOpen = iota

	// Close connection and destroy all handle routines
	MessageTypeClose = iota

	// Ping request message
	MessageTypePing = iota

	// Pong response message
	MessageTypePong = iota

	// Empty message
	MessageTypeEmpty = iota

	// Emit request, no response
	MessageTypeEmit = iota

	// Emit request, wait for response (ack)
	MessageTypeAckRequest = iota

	// ack response
	MessageTypeAckResponse = iota

	// Upgrade message
	MessageTypeUpgrade = iota

	// Blank message
	MessageTypeBlank = iota
)

type Message struct {
	Type   int
	AckId  int
	Method string
	Args   string
	Source string
}
