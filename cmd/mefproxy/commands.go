package main

import (
	"strings"

	"mefproxy/pkg/packet"
)

func registerCommandHook() {
	px.OnClientPacket("_commands", func(pk []byte) bool {
		if pk[0] != packet.IDTextPacket {
			return false
		}
		decoded, err := packet.ParseDataPacket(pk).TryDecodePacket()
		if err != nil {
			return false
		}
		tp, ok := decoded.(*packet.TextPacket)
		if !ok || !strings.HasPrefix(tp.Message, ".") {
			return false
		}
		return handleCommand(strings.TrimPrefix(tp.Message, "."))
	})
}

func handleCommand(input string) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}
	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	if cmdModules(cmd, args) {
		return true
	}
	if cmdSettings(cmd, args) {
		return true
	}
	if cmdServer(cmd, args) {
		return true
	}
	if cmdVersion(cmd, args) {
		return true
	}
	return false
}
