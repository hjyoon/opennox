package legacy

const (
	itemNameClassMask4E77E0       = uint32(0x13001000)
	itemNameWeaponClassMask4E77E0 = uint32(0x11001000)
)

type objectItemNameHooks4E77E0[O, A, D, M, S, N comparable] struct {
	class          func(O) uint32
	initData       func(O) A
	typeInd        func(O) uint16
	weaponDef      func(uint16) D
	armorDef       func(uint16) D
	unitName       func(O) N
	noInfo         func() S
	noDescription  func() S
	clear          func()
	copy           func(S)
	formatNoInfo   func(S, N)
	modifier       func(A, int) M
	modifierDesc   func(M) S
	modifierIdent  func(M) S
	definitionDesc func(D) S
	append         func(S)
	appendSpace    func()
}

// objectItemName4E77E0 is the pointer-width-independent control-flow and
// string-composition contract for GAME.EXE 004E77E0. Hooks deliberately keep
// every observable load in its original order. In particular, modifier slots
// two and three are reloaded after their leading space is appended.
func objectItemName4E77E0[O, A, D, M, S, N comparable](obj O, h objectItemNameHooks4E77E0[O, A, D, M, S, N]) {
	class := h.class(obj)
	if class&itemNameClassMask4E77E0 == 0 {
		h.copy(h.noDescription())
		return
	}

	attrs := h.initData(obj)
	typeInd := h.typeInd(obj)
	var def D
	if class&itemNameWeaponClassMask4E77E0 != 0 {
		def = h.weaponDef(typeInd)
	} else {
		def = h.armorDef(typeInd)
	}
	var zeroDef D
	if def == zeroDef {
		name := h.unitName(obj)
		format := h.noInfo()
		h.formatNoInfo(format, name)
		return
	}

	h.clear()
	var zeroMod M
	var zeroString S
	for slot := 0; slot < 2; slot++ {
		mod := h.modifier(attrs, slot)
		if mod == zeroMod {
			continue
		}
		desc := h.modifierDesc(mod)
		if desc == zeroString {
			continue
		}
		h.append(desc)
		h.appendSpace()
	}
	if desc := h.definitionDesc(def); desc != zeroString {
		h.append(desc)
	}

	mod := h.modifier(attrs, 2)
	if mod != zeroMod && h.modifierDesc(mod) != zeroString {
		h.appendSpace()
		mod = h.modifier(attrs, 2)
		h.append(h.modifierDesc(mod))
	}

	mod = h.modifier(attrs, 3)
	if mod != zeroMod && h.modifierIdent(mod) != zeroString {
		h.appendSpace()
		mod = h.modifier(attrs, 3)
		h.append(h.modifierIdent(mod))
	}
}
