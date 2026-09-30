package legacy

const questWinSource450770 = `C:\NoxPost\src\client\Gui\GUIBrief.c`

type questWinField450770 uint8

const (
	questWinKills450770 questWinField450770 = iota
	questWinGenerators450770
	questWinSecrets450770
	questWinCoopSecrets450770
)

type questWinHooks450770[P, Player, W, F, T any] struct {
	clearRows   func()
	storeTotal  func(uint32)
	storeStage  func(uint32)
	read16      func(P, uintptr) uint16
	read32      func(P, uintptr) uint32
	lookup      func(uint16) Player
	storePlayer func(int, Player)
	store16     func(int, questWinField450770, uint16)
	storeScore  func(int, uint32)
	sort        func(int)
	loadWidth   func() int32
	storeWidth  func(int32)
	child       func(int32) W
	loadText    func(string, string, int32) T
	font        func(W) F
	measure     func(F, T) int32
	lock        func(int32, int32, int8) int32
}

// questWinScreen450770 is the complete sealed 00450770..0045095A body.
// Six wire slots are kept in place; the original counts nonzero IDs (even
// unsuccessful lookups) and sorts that many PREFIX records, not a compacted
// list. Field reads and stores remain after lookup and in instruction order.
func questWinScreen450770[P, Player, W, F, T any](packet P, h questWinHooks450770[P, Player, W, F, T]) int32 {
	h.clearRows()
	h.storeTotal(0)
	h.storeTotal(uint32(h.read16(packet, 2)))
	h.storeStage(uint32(h.read16(packet, 4)))
	count := 0
	for i := 0; i < 6; i++ {
		off := uintptr(6 + 14*i)
		id := h.read16(packet, off)
		if id == 0 {
			continue
		}
		h.storePlayer(i, h.lookup(id))
		h.store16(i, questWinKills450770, h.read16(packet, off+8))
		h.store16(i, questWinGenerators450770, h.read16(packet, off+2))
		h.store16(i, questWinSecrets450770, h.read16(packet, off+4))
		h.store16(i, questWinCoopSecrets450770, h.read16(packet, off+6))
		h.storeScore(i, h.read32(packet, off+10))
		count++
	}
	h.sort(count)
	if h.loadWidth() == 0 {
		win := h.child(1010)
		labels := [...]struct {
			key  string
			line int32
		}{
			{"GUIBrief.c:GeneratorsDestroyed", 1656},
			{"GUIBrief.c:Kills", 1660},
			{"GUIBrief.c:numSecretsFound", 1664},
			{"GUIBrief.c:TotalScore", 1668},
		}
		for i, label := range labels {
			text := h.loadText(label.key, questWinSource450770, label.line)
			width := h.measure(h.font(win), text)
			cached := h.loadWidth() // The helpers can change the live cache.
			if width > cached {
				cached = width
				h.storeWidth(width)
			}
			// PE32 caps only after the last measurement, not after each one.
			if i == 3 && cached > 85 {
				h.storeWidth(85)
			}
		}
	}
	return h.lock(254, 1, 1)
}
