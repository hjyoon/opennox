package opennox

import (
	"strings"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
)

// e2eConsoleKey supplies the physical key accompanying a committed character
// on the default keyboard layout. Non-ASCII characters use an ordinary IME
// key; their text is supplied only by the separate committed-text event.
func e2eConsoleKey(ch rune) (keybind.Key, bool) {
	if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' {
		return keybind.KeyByName(strings.ToUpper(string(ch))), ch >= 'A' && ch <= 'Z'
	}
	if ch == ' ' {
		return keybind.KeySpace, false
	}
	keys := [...]keybind.Key{
		keybind.Key1, keybind.Key2, keybind.Key3, keybind.Key4, keybind.Key5,
		keybind.Key6, keybind.Key7, keybind.Key8, keybind.Key9, keybind.Key0,
		keybind.KeyMinus, keybind.KeyEqual, keybind.KeyLBracket, keybind.KeyRBracket,
		keybind.KeyBslash, keybind.KeySemiColon, keybind.KeySquote, keybind.KeyApos,
		keybind.KeyComma, keybind.KeyPeriod, keybind.KeySlash,
	}
	if i := strings.IndexRune("1234567890-=[]\\;'`,./", ch); i >= 0 {
		return keys[i], false
	}
	if i := strings.IndexRune("!@#$%^&*()_+{}|:\"~<>?", ch); i >= 0 {
		return keys[i], true
	}
	return keybind.KeyA, false
}

// consoleTypeKeyboardText models SDL's paired key/text events one character
// at a time. Releases occur on a later poll, so the raw key is really consumed
// by the focused widget instead of being overwritten by its release.
func (sc *e2eScenario) consoleTypeKeyboardText(text string) {
	for _, ch := range text {
		key, shift := e2eConsoleKey(ch)
		var down, up []seat.InputEvent
		if shift {
			down = append(down, &seat.KeyboardEvent{Key: keybind.KeyLShift, Pressed: true})
		}
		down = append(down, &seat.KeyboardEvent{Key: key, Pressed: true}, &seat.TextInputEvent{Text: string(ch)})
		up = append(up, &seat.KeyboardEvent{Key: key, Pressed: false})
		if shift {
			up = append(up, &seat.KeyboardEvent{Key: keybind.KeyLShift, Pressed: false})
		}
		sc.Input(1, "type console key and committed character", down...)
		sc.Input(1, "release console character key", up...)
	}
}
