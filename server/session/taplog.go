package session

import (
	"fmt"
	"os"
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// TEMPORARY DEBUG LOGGING - delete this file and the two tapLog hooks in
// session.go (the tapLogMode field and the call in handlePacket) when done.
//
// It logs the packets that could carry "the player used / swung an item" for
// mobile clients, so we can see exactly what a quick tap sends versus a long
// press. Every line starts with "TAPLOG". Set DF_TAPLOG_ALL=1 to log every
// device instead of only Android / iOS / FireOS.

var tapLogAll = os.Getenv("DF_TAPLOG_ALL") == "1"

// tapLogWanted reports whether packets of this session should be logged.
func tapLogWanted(s *Session) bool {
	if tapLogAll {
		return true
	}
	switch s.conn.ClientData().DeviceOS {
	case protocol.DeviceAndroid, protocol.DeviceIOS, protocol.DeviceFireOS:
		return true
	}
	return false
}

// tapLogStr formats v with its type and field names, capped so one huge
// packet can't flood the log.
func tapLogStr(v any) string {
	str := fmt.Sprintf("%T %+v", v, v)
	if len(str) > 600 {
		str = str[:600] + "...(cut)"
	}
	return str
}

// tapLogHeld returns the encoded name of the item in the player's main hand.
func tapLogHeld(c Controllable) string {
	held, _ := c.HeldItems()
	if held.Empty() {
		return "empty"
	}
	name, _ := held.Item().EncodeItem()
	return name
}

// tapLog logs pk if it is relevant to item use. PlayerAuthInput arrives every
// tick, so it is only logged when it carries a use/swing/interaction flag, plus
// one "mode" line whenever the input mode or interaction model changes.
func tapLog(s *Session, pk packet.Packet, c Controllable) {
	if !tapLogWanted(s) {
		return
	}
	switch p := pk.(type) {
	case *packet.PlayerAuthInput:
		packed := uint64(p.InputMode)<<32 | uint64(uint32(p.InteractionModel)) | 1<<63
		if s.tapLogMode.Swap(packed) != packed {
			s.conf.Log.Info("TAPLOG mode",
				"inputMode", p.InputMode,
				"interactionModel", p.InteractionModel,
				"playMode", p.PlayMode,
				"deviceOS", s.conn.ClientData().DeviceOS,
			)
		}

		f := p.InputData
		var names []string
		if f.Load(packet.InputFlagMissedSwing) {
			names = append(names, "MissedSwing")
		}
		if f.Load(packet.InputFlagStartUsingItem) {
			names = append(names, "StartUsingItem")
		}
		if f.Load(packet.InputFlagPerformItemInteraction) {
			names = append(names, "PerformItemInteraction")
		}
		if f.Load(packet.InputFlagPerformBlockActions) {
			names = append(names, "PerformBlockActions")
		}
		if f.Load(packet.InputFlagPerformItemStackRequest) {
			names = append(names, "PerformItemStackRequest")
		}
		if len(names) == 0 {
			return
		}

		var interaction string
		if f.Load(packet.InputFlagPerformItemInteraction) {
			if d, ok := p.ItemInteractionData.Value(); ok {
				interaction = tapLogStr(d)
			}
		}
		s.conf.Log.Info("TAPLOG PlayerAuthInput",
			"flags", strings.Join(names, ","),
			"hotbarTouchOnly", f.Load(packet.InputFlagIsHotbarTouchOnly),
			"held", tapLogHeld(c),
			"tick", p.Tick,
			"itemInteraction", interaction,
		)
	case *packet.InventoryTransaction:
		s.conf.Log.Info("TAPLOG InventoryTransaction",
			"held", tapLogHeld(c),
			"data", tapLogStr(p.TransactionData),
		)
	case *packet.Animate, *packet.PlayerAction, *packet.MobEquipment, *packet.Interact:
		s.conf.Log.Info("TAPLOG "+fmt.Sprintf("%T", pk),
			"held", tapLogHeld(c),
			"data", tapLogStr(pk),
		)
	}
}
