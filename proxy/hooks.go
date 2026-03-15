package proxy

import "sync"

type PacketHook func(pk []byte) (drop bool)
type EventHook func()

type hookRegistry struct {
	mu sync.RWMutex

	clientPacket    map[string]PacketHook
	serverPacket    map[string]PacketHook
	onConnect       map[string]EventHook
	onDisconnect    map[string]EventHook
	onSrvDisconnect map[string]EventHook
}

func newHookRegistry() *hookRegistry {
	return &hookRegistry{
		clientPacket:    make(map[string]PacketHook),
		serverPacket:    make(map[string]PacketHook),
		onConnect:       make(map[string]EventHook),
		onDisconnect:    make(map[string]EventHook),
		onSrvDisconnect: make(map[string]EventHook),
	}
}

func (h *hookRegistry) addClientPacket(key string, fn PacketHook) {
	h.mu.Lock()
	h.clientPacket[key] = fn
	h.mu.Unlock()
}

func (h *hookRegistry) addServerPacket(key string, fn PacketHook) {
	h.mu.Lock()
	h.serverPacket[key] = fn
	h.mu.Unlock()
}

func (h *hookRegistry) addConnect(key string, fn EventHook) {
	h.mu.Lock()
	h.onConnect[key] = fn
	h.mu.Unlock()
}

func (h *hookRegistry) addDisconnect(key string, fn EventHook) {
	h.mu.Lock()
	h.onDisconnect[key] = fn
	h.mu.Unlock()
}

func (h *hookRegistry) addSrvDisconnect(key string, fn EventHook) {
	h.mu.Lock()
	h.onSrvDisconnect[key] = fn
	h.mu.Unlock()
}

func (h *hookRegistry) remove(key string) {
	h.mu.Lock()
	delete(h.clientPacket, key)
	delete(h.serverPacket, key)
	delete(h.onConnect, key)
	delete(h.onDisconnect, key)
	delete(h.onSrvDisconnect, key)
	h.mu.Unlock()
}

func (h *hookRegistry) runClientPacket(pk []byte) bool {
	h.mu.RLock()
	hooks := make([]PacketHook, 0, len(h.clientPacket))
	for _, fn := range h.clientPacket {
		hooks = append(hooks, fn)
	}
	h.mu.RUnlock()
	for _, fn := range hooks {
		if fn(pk) {
			return true
		}
	}
	return false
}

func (h *hookRegistry) runServerPacket(pk []byte) bool {
	h.mu.RLock()
	hooks := make([]PacketHook, 0, len(h.serverPacket))
	for _, fn := range h.serverPacket {
		hooks = append(hooks, fn)
	}
	h.mu.RUnlock()
	for _, fn := range hooks {
		if fn(pk) {
			return true
		}
	}
	return false
}

func (h *hookRegistry) runConnect() {
	h.mu.RLock()
	hooks := make([]EventHook, 0, len(h.onConnect))
	for _, fn := range h.onConnect {
		hooks = append(hooks, fn)
	}
	h.mu.RUnlock()
	for _, fn := range hooks {
		fn()
	}
}

func (h *hookRegistry) runDisconnect() {
	h.mu.RLock()
	hooks := make([]EventHook, 0, len(h.onDisconnect))
	for _, fn := range h.onDisconnect {
		hooks = append(hooks, fn)
	}
	h.mu.RUnlock()
	for _, fn := range hooks {
		fn()
	}
}

func (h *hookRegistry) runSrvDisconnect() {
	h.mu.RLock()
	hooks := make([]EventHook, 0, len(h.onSrvDisconnect))
	for _, fn := range h.onSrvDisconnect {
		hooks = append(hooks, fn)
	}
	h.mu.RUnlock()
	for _, fn := range hooks {
		fn()
	}
}
