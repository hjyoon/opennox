package legacy

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/server"
)

type clientAskInfoDeps4BF050 struct {
	scratch, initial, space, prefixFormat *uint16
	copyText, appendText                  func(*uint16, *uint16)
	formatText                            func(*uint16, *uint16, unsafe.Pointer)
	prettyName                            func(uint32) *uint16
	typeName                              func(uint32) *byte
	weaponDef, armorDef                   func(uint32) *server.Modifier
	language                              func() int
	loadString                            func(string, int) *uint16
	spellTitle, creatureName, abilityName func(uint32) *uint16
	unitCode                              func(*client.Drawable) uint16
	send                                  func(int, []byte, *server.Object, int) int
}

// clientAskInfoNative4BF050 restores the complete GAME.EXE tooltip contract.
// Drawable identity and modifier/string pointers are native-width; only the
// four-byte book information request is serialized as fixed-width wire data.
func clientAskInfoNative4BF050(dr *client.Drawable, d clientAskInfoDeps4BF050) *uint16 {
	*d.scratch = 0
	d.copyText(d.scratch, d.initial)
	if dr == nil {
		return d.scratch
	}
	class := uint32(dr.ObjClass)
	appendText := func(p *uint16) { d.appendText(d.scratch, p) }
	if class&0x13001000 == 0 {
		if class&0x100 == 0 {
			if p := d.prettyName(dr.TypeIDVal); p != nil {
				return p
			}
			return d.scratch
		}
		subclass := uint32(dr.ObjSubClass)
		var kind byte
		var pending uint32
		var title func(uint32) *uint16
		switch {
		case subclass&1 != 0:
			kind, pending, title = 1, 137, d.spellTitle
		case subclass&2 != 0:
			kind, pending, title = 2, 41, d.creatureName
		case subclass&4 != 0:
			kind, pending, title = 4, 6, d.abilityName
		default:
			if p := d.prettyName(dr.TypeIDVal); p != nil {
				return p
			}
			return d.scratch
		}
		id := dr.UnionEffect().Field_108
		if id == 0 {
			packet := [4]byte{0xE2, 0, 0, kind}
			binary.LittleEndian.PutUint16(packet[1:3], d.unitCode(dr))
			// Publish the pending sentinel before the network callback. The
			// callback may synchronously replace it with the server's reply.
			dr.UnionEffect().Field_108 = pending
			d.send(31, packet[:], nil, 1)
			return d.scratch
		}
		if id == pending {
			return d.scratch
		}
		lang := d.language()
		if kind == 2 {
			if lang == 3 || lang == 5 {
				appendText(d.loadString("LoreScroll", 313))
				appendText(d.space)
				appendText(title(dr.UnionEffect().Field_108))
			} else {
				d.formatText(d.scratch, d.prefixFormat, unsafe.Pointer(title(id)))
				appendText(d.loadString("LoreScroll", 320))
			}
		} else if lang == 6 {
			appendText(title(id))
			appendText(d.space)
			line := 288
			if kind == 4 {
				line = 342
			}
			appendText(d.loadString("BookOf", line))
		} else {
			line := 292
			if kind == 4 {
				line = 346
			}
			d.formatText(d.scratch, d.prefixFormat, unsafe.Pointer(d.loadString("BookOf", line)))
			appendText(title(dr.UnionEffect().Field_108))
		}
		return d.scratch
	}
	var def *server.Modifier
	if class&0x11001000 != 0 {
		def = d.weaponDef(dr.TypeIDVal)
	} else {
		def = d.armorDef(dr.TypeIDVal)
	}
	if def == nil {
		name := d.typeName(dr.TypeIDVal)
		format := d.loadString("NoArmsInfo", 53)
		d.formatText(d.scratch, format, unsafe.Pointer(name))
		return d.scratch
	}
	var mods [4]*uint16
	weapon := uint32(dr.ObjClass)&0x1000000 != 0
	// The entry class was already selected above. The definition callback may
	// change the live class; GAME.EXE does not add another class-validity guard.
	item := (*client.DrawableUnionItem)(unsafe.Pointer(&dr.Union))
	if !weapon || uint32(dr.ObjSubClass)&0x7800000 == 0 {
		if m := (*server.ModifierEff)(item.Field_108); m != nil {
			mods[0] = m.Description16()
		}
		if m := (*server.ModifierEff)(item.Field_109); m != nil {
			mods[1] = m.Description16()
		}
	}
	base := def.Description16()
	if !weapon || uint32(dr.ObjSubClass)&0x7800000 == 0 {
		if m := (*server.ModifierEff)(item.Field_110); m != nil {
			mods[2] = m.Description16()
		}
		if m := (*server.ModifierEff)(item.Field_111); m != nil {
			mods[3] = m.SecondaryDescription16()
		}
	}
	// Presence means a non-null UTF-16 pointer, not a nonempty decoded Go
	// string. Empty descriptions still contribute GAME.EXE's exact spaces.
	prefix := func(p *uint16) {
		if p != nil {
			appendText(p)
			appendText(d.space)
		}
	}
	suffix := func(p *uint16) {
		if p != nil {
			appendText(d.space)
			appendText(p)
		}
	}
	switch d.language() {
	case 2:
		appendText(base)
		suffix(mods[2])
		suffix(mods[3])
		prefix(mods[1])
		suffix(mods[0])
	case 3:
		prefix(mods[0])
		appendText(base)
		suffix(mods[1])
		suffix(mods[2])
		suffix(mods[3])
	case 5:
		appendText(base)
		suffix(mods[1])
		suffix(mods[0])
		suffix(mods[2])
		suffix(mods[3])
	case 6:
		prefix(mods[0])
		prefix(mods[2])
		prefix(mods[3])
		suffix(mods[1])
		appendText(base)
	default:
		prefix(mods[0])
		prefix(mods[1])
		appendText(base)
		suffix(mods[2])
		suffix(mods[3])
	}
	return d.scratch
}
