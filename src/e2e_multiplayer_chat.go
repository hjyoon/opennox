package opennox

import (
	"fmt"
	"strings"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/opennox/v1/client/gui"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func e2eChatInput() string {
	win := noxClient.GUI.Focused()
	if win == nil || !win.DrawData().Style.IsEntryField() || legacy.Sub_46A4A0() == 0 {
		return ""
	}
	return alloc.GoString16(&(*gui.EntryFieldData)(win.WidgetData).Text[0])
}

// CheckMultiplayerChat supplies only seat input. Observations require the real
// chat modal, submitted packet's console echo and allocated client chat bubble.
// It runs in an owned host and a separate UDP peer without replacing handlers.
func (sc *e2eScenario) CheckMultiplayerChat(role, name string) {
	role = strings.ToUpper(role)
	if role != "HOST" && role != "PEER" {
		panic("multiplayer chat requires role host or peer")
	}
	sc.add(1, name, func() {
		lang := noxClient.Inp.Language()
		if noxflags.HasGame(noxflags.GameHost) != (role == "HOST") ||
			!nox_client_isConnected() || noxClient.ClientPlayerUnit() == nil ||
			legacy.Sub_46A4A0() != 0 || e2eConsoleFocused() || lang != 6 && lang != 8 {
			e2eError(fmt.Errorf("multiplayer chat requires an active Korean/Chinese client with closed chat/console"))
		}
	})
	for _, tc := range []struct {
		name, text string
		key        keybind.Key
		compose    bool
	}{
		{"english", "MPCHAT " + role + " ENGLISH 01", keybind.KeyEnter, false},
		{"unicode", "MPCHAT " + role + " 한글 中 02", keybind.KeyEnter, false},
		{"keypad", "MPCHAT " + role + " KEYPAD 03", keybind.KeyKpEnter, false},
		{"composition", "MPCHAT " + role + " IME 04 ", keybind.KeyEnter, true},
	} {
		sc.Key(keybind.KeyEnter, "open multiplayer chat")
		sc.add(2, "verify empty chat modal", func() {
			if legacy.Sub_46A4A0() == 0 || e2eChatInput() != "" {
				e2eError(fmt.Errorf("chat did not open with empty input"))
			}
		})
		sc.consoleTypeKeyboardText(tc.text)
		want := tc.text
		if tc.compose {
			sc.Input(1, "begin chat IME composition", &seat.TextEditEvent{Text: "ㅎ"})
			sc.Input(1, "confirm chat IME before commit in the same poll",
				&seat.KeyboardEvent{Key: tc.key, Pressed: true}, &seat.TextInputEvent{Text: "한"})
			sc.Input(1, "release IME confirming Return", &seat.KeyboardEvent{Key: tc.key, Pressed: false})
			want += "한"
		}
		sc.add(2, "verify unsent committed chat text", func() {
			if legacy.Sub_46A4A0() == 0 || e2eChatInput() != want || strings.Contains(e2eConsoleLines(), "> "+want) {
				e2eError(fmt.Errorf("chat %s committed text was lost, duplicated or submitted early: %q", tc.name, e2eChatInput()))
			}
		})
		sc.Key(tc.key, "submit multiplayer chat")
		sc.addWhen(1, "verify received chat echo", 300, func() bool {
			return strings.Contains(e2eConsoleLines(), "> "+want)
		}, func() {
			if legacy.Sub_46A4A0() != 0 || noxClient.GUI.Focused() != nil ||
				strings.Count(e2eConsoleLines(), "> "+want) != 1 ||
				legacy.Nox_xxx_netCode2ChatBubble_48D850(legacy.ClientPlayerNetCode()) == 0 {
				e2eError(fmt.Errorf("chat %s did not echo once, create a bubble and release input", tc.name))
				return
			}
			path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
			if err != nil {
				e2eError(err)
				return
			}
			e2eLog.Printf("MULTIPLAYER CHAT PASS: role=%s case=%s text=%q netcode=%d frame=%d path=%s", role, tc.name, want, legacy.ClientPlayerNetCode(), noxClient.Server.Frame(), path)
		})
	}
	sc.Key(keybind.KeyEnter, "open chat before cancellation")
	sc.consoleTypeKeyboardText("MPCHAT " + role + " CANCELLED")
	sc.Key(keybind.KeyEsc, "cancel multiplayer chat")
	sc.add(2, "verify cancelled chat", func() {
		if legacy.Sub_46A4A0() != 0 || strings.Contains(e2eConsoleLines(), "MPCHAT "+role+" CANCELLED") {
			e2eError(fmt.Errorf("Escape failed to cancel chat"))
		}
	})
	sc.Key(keybind.KeyEnter, "reopen chat after cancellation")
	sc.add(2, "verify cancellation cleared the reopened input", func() {
		if legacy.Sub_46A4A0() == 0 || e2eChatInput() != "" {
			e2eError(fmt.Errorf("cancelled text survived reopening chat"))
		}
	})
	sc.Key(keybind.KeyEnter, "close empty chat")
	sc.add(4, "verify empty chat releases input", func() {
		if legacy.Sub_46A4A0() != 0 || noxClient.GUI.Focused() != nil {
			e2eError(fmt.Errorf("empty chat did not close"))
			return
		}
		e2eLog.Printf("MULTIPLAYER CHAT COMPLETE: role=%s sent=4 cancelled=1 empty=1", role)
	})
}

// The host waits for the independently typed peer messages, then replies using
// the same input widget. The peer must receive that reply across the UDP link.
func (sc *e2eScenario) CheckMultiplayerChatExchange(role, name string) {
	role = strings.ToUpper(role)
	if role != "HOST" && role != "PEER" {
		panic("multiplayer chat exchange requires role host or peer")
	}
	sc.add(1, name, func() {
		if noxflags.HasGame(noxflags.GameHost) != (role == "HOST") {
			e2eError(fmt.Errorf("multiplayer chat exchange role does not match the connected process"))
		}
	})
	reply := "MPCHAT HOST REPLY 05 한글"
	if role == "HOST" {
		messages := []string{
			"MPCHAT PEER ENGLISH 01", "MPCHAT PEER 한글 中 02",
			"MPCHAT PEER KEYPAD 03", "MPCHAT PEER IME 04 한",
		}
		sc.addWhen(1, name, 18000, func() bool {
			lines := e2eConsoleLines()
			for _, text := range messages {
				if !strings.Contains(lines, "> "+text) {
					return false
				}
			}
			return true
		}, func() {
			for _, text := range messages {
				if strings.Count(e2eConsoleLines(), "> "+text) != 1 {
					e2eError(fmt.Errorf("peer message did not reach host exactly once: %q", text))
				}
			}
			e2eLog.Printf("MULTIPLAYER CHAT INBOUND: role=HOST received=4 source=independent-UDP-peer")
		})
		sc.Key(keybind.KeyEnter, "open host reply input")
		sc.Wait(2, "settle host reply input")
		sc.consoleTypeKeyboardText(reply)
		sc.Key(keybind.KeyEnter, "send host reply to independent peer")
	}
	sc.addWhen(1, "receive host reply", 600, func() bool {
		return strings.Contains(e2eConsoleLines(), "> "+reply)
	}, func() {
		if strings.Count(e2eConsoleLines(), "> "+reply) != 1 {
			e2eError(fmt.Errorf("host reply did not arrive exactly once"))
			return
		}
		e2eLog.Printf("MULTIPLAYER CHAT EXCHANGE PASS: role=%s reply=%q", role, reply)
	})
}
