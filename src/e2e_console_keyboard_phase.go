package opennox

import (
	"fmt"
	"unicode/utf16"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
)

// Run paired keyboard/text checks after the original command regression has
// restored the frame limiter. Extra character polls must not change the
// deterministic clock's pacing during the original SET/UNSET limiter checks.
func (sc *e2eScenario) checkConsoleKeyboardText() {
	sc.add(1, "verify paired console keyboard fixture", func() {
		lang := noxClient.Inp.Language()
		if !e2eConsoleFocused() || e2eConsoleInput() != "" || lang != 6 && lang != 8 {
			e2eError(fmt.Errorf("paired console fixture requires focused empty Korean/Chinese input"))
		}
	})
	for _, tc := range []struct{ name, text string }{
		{"single-character", "a"},
		{"repeated-letters-digits-and-shift", "aaa AAA 111 !_+"},
		{"committed-Korean-and-Chinese", "한글 中"},
	} {
		sc.consoleTypeKeyboardText(tc.text)
		sc.consoleAssertKeyboardText(tc.name, tc.text)
	}
	sc.Input(1, "paste one multi-character console commit", &seat.TextInputEvent{Text: "aaa 한글"})
	sc.consoleAssertKeyboardText("multi-character-paste", "aaa 한글")

	sc.Input(1, "console key before committed text", &seat.KeyboardEvent{Key: keybind.KeyA, Pressed: true})
	sc.add(1, "verify no scan-code text before commit", func() {
		if e2eConsoleInput() != "" {
			e2eError(fmt.Errorf("console inserted scan-code text before the text commit"))
		}
	})
	sc.Input(1, "commit console text on a later poll", &seat.TextInputEvent{Text: "a"})
	sc.Input(1, "release the earlier console key", &seat.KeyboardEvent{Key: keybind.KeyA, Pressed: false})
	sc.consoleAssertKeyboardText("key-before-later-text", "a")

	sc.Input(1, "console text before its physical key", &seat.TextInputEvent{Text: "a"})
	sc.Input(1, "console key on the later poll", &seat.KeyboardEvent{Key: keybind.KeyA, Pressed: true})
	sc.Input(1, "release the later console key", &seat.KeyboardEvent{Key: keybind.KeyA, Pressed: false})
	sc.consoleAssertKeyboardText("text-before-later-key", "a")

	sc.Input(1, "begin console composition", &seat.TextEditEvent{Text: "ㅎ"}, &seat.KeyboardEvent{Key: keybind.KeyG, Pressed: true})
	sc.add(1, "verify pending console composition", func() {
		if e2eConsoleInput() != "" || noxClient.Inp.GetTextEditBuf() != "ㅎ" {
			e2eError(fmt.Errorf("pending console composition leaked a printable scan-code character"))
		}
	})
	sc.Input(1, "commit console composition", &seat.TextInputEvent{Text: "한"})
	sc.Input(1, "release console composition key", &seat.KeyboardEvent{Key: keybind.KeyG, Pressed: false})
	sc.consoleAssertKeyboardText("pending-composition-then-commit", "한")

	sc.Input(1, "reverse paired console event order", &seat.TextInputEvent{Text: "a"}, &seat.KeyboardEvent{Key: keybind.KeyA, Pressed: true})
	sc.Input(1, "release reversed console key", &seat.KeyboardEvent{Key: keybind.KeyA, Pressed: false})
	sc.consoleAssertKeyboardText("text-before-key-same-poll", "a")

	// Submit an actual command using the same paired events; do not bypass
	// the native input widget or write synthetic command output.
	sc.consoleTypeKeyboardText("HELP SET")
	sc.add(2, "verify paired console command input", func() {
		if !e2eConsoleFocused() || e2eConsoleInput() != "HELP SET" {
			e2eError(fmt.Errorf("paired console command text was duplicated or lost"))
		}
	})
	sc.Key(keybind.KeyEnter, "execute paired console command")
	sc.add(4, "verify paired console command output", func() {
		if !e2eConsoleFocused() || e2eConsoleInput() != "" || !e2eConsoleHasHelp("set sysop") {
			e2eError(fmt.Errorf("paired console command did not execute and clear its native entry"))
			return
		}
		e2eLog.Printf("CONSOLE KEYBOARD COMMAND PASS: input=HELP SET frame=%d", noxServer.Frame())
	})
}

// Read real native text, then clear it with physical Backspace events. Checking
// every prefix also proves that a control key still removes exactly one UTF-16
// unit. The fixtures deliberately contain BMP characters, not surrogate pairs.
func (sc *e2eScenario) consoleAssertKeyboardText(name, want string) {
	sc.add(2, "verify console keyboard text: "+name, func() {
		if !e2eConsoleFocused() || e2eConsoleInput() != want {
			e2eError(fmt.Errorf("console keyboard %s: text=%q, want=%q", name, e2eConsoleInput(), want))
			return
		}
		e2eLog.Printf("CONSOLE KEYBOARD TEXT PASS: case=%s text=%q frame=%d", name, want, noxServer.Frame())
	})
	units := utf16.Encode([]rune(want))
	for i := len(units) - 1; i >= 0; i-- {
		prefix := string(utf16.Decode(units[:i]))
		sc.Input(1, "delete one console character", &seat.KeyboardEvent{Key: keybind.KeyBackspace, Pressed: true})
		sc.Input(1, "release console Backspace", &seat.KeyboardEvent{Key: keybind.KeyBackspace, Pressed: false})
		sc.add(1, "verify one-character console deletion", func() {
			if !e2eConsoleFocused() || e2eConsoleInput() != prefix {
				e2eError(fmt.Errorf("console Backspace %s: text=%q, want=%q", name, e2eConsoleInput(), prefix))
			}
		})
	}
	sc.add(1, "verify cleared console keyboard case: "+name, func() {
		if !e2eConsoleFocused() || e2eConsoleInput() != "" {
			e2eError(fmt.Errorf("console keyboard %s was not cleared by real Backspace events", name))
			return
		}
		e2eLog.Printf("CONSOLE KEYBOARD CLEAR PASS: case=%s frame=%d", name, noxServer.Frame())
	})
}
