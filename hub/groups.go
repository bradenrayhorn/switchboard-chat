package hub

import (
	"github.com/segmentio/ksuid"
	"sync"
)

// Thread-safe data structure to store groups and associated clients.
type Groups struct {
	lock   sync.RWMutex
	groups map[string]map[ksuid.KSUID]*Client
}

func NewGroups() *Groups {
	return &Groups{
		groups: make(map[string]map[ksuid.KSUID]*Client, 0),
	}
}

func (g *Groups) addClient(groups []string, client *Client) bool {
	g.lock.Lock()
	defer g.lock.Unlock()
	changed := false
	for _, groupID := range groups {
		if _, ok := g.groups[groupID]; ok {
			g.groups[groupID][client.id] = client
		} else {
			g.groups[groupID] = map[ksuid.KSUID]*Client{client.id: client}
			changed = true
		}
	}
	return changed
}

func (g *Groups) removeClient(client *Client) bool {
	g.lock.Lock()
	defer g.lock.Unlock()
	changed := false
	for _, groupID := range client.groupIDs {
		delete(g.groups[groupID], client.id)
		if len(g.groups[groupID]) == 0 {
			delete(g.groups, groupID)
			changed = true
		}
	}
	return changed
}

func (g *Groups) getClientMap(groupID string) map[ksuid.KSUID]*Client {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.groups[groupID]
}

func (g *Groups) getGroupIDs(prefix string) []string {
	g.lock.RLock()
	defer g.lock.RUnlock()
	groupIDs := make([]string, len(g.groups))
	i := 0
	for k := range g.groups {
		groupIDs[i] = prefix + k
		i++
	}
	return groupIDs
}
