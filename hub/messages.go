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

type RedisMessageType int

const (
	RedisGroupsChanged RedisMessageType = iota
)

type RedisMessage struct {
	RedisMessageType RedisMessageType `json:"type"`
	Body             interface{}      `json:"body"`
}

type RedisGroupChangedMessage struct {
	RedisMessage
	Body struct {
		GroupJoined string `json:"group_joined"`
		GroupLeft   string `json:"group_left"`
	} `json:"body"`
}
