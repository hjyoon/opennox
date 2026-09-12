// Package noxgroupwalk preserves the traversal contract of GAME.EXE 00502670.
// The hooks keep its 32-bit identifiers separate from native-width pointers.
package noxgroupwalk

type Hooks[G comparable, I comparable] struct {
	Kind         func(G) uint8
	First        func(G) I
	Next         func(I) I
	IDs          func(I) (uint32, uint32)
	ResolveGroup func(uint32) G
	Visit        func(kind uint8, first, second uint32)
}

// Walk visits items of the requested kind, then recursively visits groups.
// A wall group also scans its items as group IDs after visiting its walls,
// matching the original switch fallthrough. Next is read after each callback
// or recursive call, so mutations made there affect the remainder of the walk.
func Walk[G comparable, I comparable](group G, expected int32, h Hooks[G, I]) {
	var nilGroup G
	var nilItem I
	if group == nilGroup {
		return
	}
	kind := h.Kind(group)
	if kind > 3 {
		return
	}
	switch kind {
	case 0, 1:
		if int32(kind) != expected {
			return
		}
		for it := h.First(group); it != nilItem; it = h.Next(it) {
			first, second := h.IDs(it)
			h.Visit(kind, first, second)
		}
	case 2:
		if expected != 2 {
			return
		}
		for it := h.First(group); it != nilItem; it = h.Next(it) {
			first, second := h.IDs(it)
			h.Visit(kind, first, second)
		}
		fallthrough
	case 3:
		for it := h.First(group); it != nilItem; it = h.Next(it) {
			first, _ := h.IDs(it)
			if child := h.ResolveGroup(first); child != nilGroup {
				Walk(child, expected, h)
			}
		}
	}
}
