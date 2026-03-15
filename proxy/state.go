package proxy

import (
	"math"
	"sync"

	"github.com/google/uuid"
	"mefproxy/pkg/math/vec3"
)

type Player struct {
	Nick            string
	EntityRuntimeID uint32
	EntityUniqueID  int32
	UUID            uuid.UUID
	Skin            string
	Location        vec3.Vector3
	Health          float32
}

type State struct {
	mu sync.RWMutex

	myEID      uint32
	myLocation vec3.Vector3

	players    map[uint32]*Player
	playerList map[string]string
	uuidToEID  map[string]uint32
}

func newState() *State {
	return &State{
		myEID:      12898,
		players:    make(map[uint32]*Player),
		playerList: make(map[string]string),
		uuidToEID:  make(map[string]uint32),
	}
}

func (s *State) MyEID() uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.myEID
}

func (s *State) SetMyEID(eid uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.myEID = eid
}

func (s *State) MyLocation() vec3.Vector3 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.myLocation
}

func (s *State) SetMyLocation(loc vec3.Vector3) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.myLocation = loc
}

func (s *State) Players() []*Player {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Player, 0, len(s.players))
	for _, p := range s.players {
		out = append(out, p)
	}
	return out
}

func (s *State) NearestPlayer() *Player {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var nearest *Player
	var nearestDist float32 = math.MaxFloat32
	for _, p := range s.players {
		d := vec3.Distance(s.myLocation, p.Location)
		if d < nearestDist {
			nearestDist = d
			nearest = p
		}
	}
	return nearest
}

func (s *State) NearestPlayers(maxCount int) []*Player {
	s.mu.RLock()
	defer s.mu.RUnlock()
	type entry struct {
		p    *Player
		dist float32
	}
	var entries []entry
	for _, p := range s.players {
		d := vec3.Distance(s.myLocation, p.Location)
		entries = append(entries, entry{p, d})
	}
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].dist < entries[i].dist {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	if maxCount > len(entries) {
		maxCount = len(entries)
	}
	out := make([]*Player, maxCount)
	for i := 0; i < maxCount; i++ {
		out[i] = entries[i].p
	}
	return out
}

func (s *State) PlayerByEID(eid uint32) *Player {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.players[eid]
}

func (s *State) PlayerByNick(nick string) *Player {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.players {
		if p.Nick == nick {
			return p
		}
	}
	return nil
}

func (s *State) PlayerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.players)
}

func (s *State) AddPlayer(p *Player) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.players[p.EntityRuntimeID]; !exists {
		if skin, ok := s.playerList[p.UUID.String()]; ok {
			p.Skin = skin
		}
		s.players[p.EntityRuntimeID] = p
		s.uuidToEID[p.UUID.String()] = p.EntityRuntimeID
	}
}

func (s *State) RemovePlayer(eid uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.players[eid]; ok {
		delete(s.uuidToEID, p.UUID.String())
		delete(s.players, eid)
	}
}

func (s *State) RemovePlayerByUUID(uid string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if eid, ok := s.uuidToEID[uid]; ok {
		delete(s.players, eid)
		delete(s.uuidToEID, uid)
	}
	delete(s.playerList, uid)
}

func (s *State) UpdatePlayerLocation(eid uint32, loc vec3.Vector3) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.players[eid]; ok {
		p.Location = loc
	}
}

func (s *State) UpdatePlayerHealth(eid uint32, health float32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.players[eid]; ok {
		p.Health = health
	}
}

func (s *State) IsPlayer(eid uint32) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.players[eid]
	return ok
}

func (s *State) CacheSkin(uid, skin string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playerList[uid] = skin
}

func (s *State) RemoveSkinCache(uid string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.playerList, uid)
}

func (s *State) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.myEID = 12898
	s.myLocation = vec3.Vector3{}
	s.players = make(map[uint32]*Player)
	s.playerList = make(map[string]string)
	s.uuidToEID = make(map[string]uint32)
}

func (s *State) resetPlayers() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.myEID = 12898
	s.players = make(map[uint32]*Player)
	s.uuidToEID = make(map[string]uint32)
}
