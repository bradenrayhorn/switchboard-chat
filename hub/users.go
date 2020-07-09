package hub

import (
	"github.com/segmentio/ksuid"
	"sync"
)

// Thread-safe data structure to store a list of user IDs and their associated client IDs.
type Users struct {
	lock  sync.RWMutex
	users map[string][]ksuid.KSUID
}

func NewUsers() *Users {
	return &Users{
		users: make(map[string][]ksuid.KSUID, 0),
	}
}

func (u *Users) addUser(userID string, clientID ksuid.KSUID) {
	u.lock.Lock()
	defer u.lock.Unlock()
	if _, ok := u.users[userID]; ok {
		u.users[userID] = append(u.users[userID], clientID)
	} else {
		u.users[userID] = []ksuid.KSUID{clientID}
	}
}

func (u *Users) removeUser(userID string, clientID ksuid.KSUID) {
	u.lock.Lock()
	defer u.lock.Unlock()
	if clients, ok := u.users[userID]; ok {
		for i, c := range clients {
			if c == clientID {
				clients = append(clients[:i], clients[i+1:]...)
				break
			}
		}
		if len(clients) == 0 {
			delete(u.users, userID)
		} else {
			u.users[userID] = clients
		}
	}
}

func (u *Users) getUserIDs(prefix string) []string {
	u.lock.RLock()
	defer u.lock.RUnlock()
	userIDs := make([]string, len(u.users))
	i := 0
	for k := range u.users {
		userIDs[i] = prefix + k
		i++
	}
	return userIDs
}

func (u *Users) getClientIDs(userID string) []ksuid.KSUID {
	u.lock.RLock()
	defer u.lock.RUnlock()
	if clients, ok := u.users[userID]; ok {
		return clients
	} else {
		return make([]ksuid.KSUID, 0)
	}
}
