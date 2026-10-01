package legacy

const (
	questGameOverTimerSource49B6E0 = `C:\NoxPost\src\client\Gui\GUIGGOvr.c`
	questGameOverTimerKey49B6E0    = "Rules.c:Time"
	questGameOverTimerFormat49B6E0 = "%s - %d"
)

type questGameOverTimerHooks49B6E0[W, P comparable, T any] struct {
	root      func() W
	hidden    func(W) int32
	fps       func() uint32
	frame     func() uint32
	start     func() uint32
	player    func() P
	index     func(P) uint8
	copyEmpty func()
	loadText  func(string, string, int32) T
	format    func(T, int32)
	child     func(W, int32) W
	setText   func(W) int32
}

// questGameOverTimer49B6E0 implements the entire sealed GAME.EXE body
// 0049B6E0..0049B79D. FPS is cached once for both the DWORD countdown and
// unsigned division. The final child lookup reloads the window root after
// copying/formatting text; the player index is a native BYTE field.
func questGameOverTimer49B6E0[W, P comparable, T any](h questGameOverTimerHooks49B6E0[W, P, T]) int32 {
	var noWindow W
	root := h.root()
	if root == noWindow {
		return 0
	}
	if result := h.hidden(root); result != 0 {
		return result
	}
	fps := h.fps()
	frame := h.frame()
	start := h.start()
	remaining := int32(fps*30 - frame + start)
	if remaining < 0 {
		remaining = 0
	}
	var noPlayer P
	player := h.player()
	if player != noPlayer && h.index(player) == 31 {
		h.copyEmpty()
	} else {
		// Zero FPS faults here, before the string lookup, exactly as DIV ECX.
		seconds := uint32(remaining) / fps
		text := h.loadText(questGameOverTimerKey49B6E0, questGameOverTimerSource49B6E0, 265)
		h.format(text, int32(seconds))
	}
	return h.setText(h.child(h.root(), 10712))
}
