package hub

// Client messages

type SocketMessageType int

const (
	GroupChange SocketMessageType = iota
	Message
)

type SocketMessage struct {
	SocketMessageType SocketMessageType `json:"type"`
	Body              interface{}       `json:"body"`
}

// Service messages
