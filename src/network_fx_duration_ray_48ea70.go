package opennox

import (
	"encoding/binary"
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

// The PE32 ray list at 0x5D4594:1303924 contains drawable pointers in
// uint32 slots. Keep those pointers at their native width on 64-bit hosts.
type clientDurationRay48EA70 struct {
	drawable       *client.Drawable
	source, target uint16
	kind           byte
}

var clientDurationRayNames48EA70 = [...]string{
	"PlasmaRay", "OrbRay", "DynamicChainLightning", "DynamicEnergyBolt",
	"DrainManaRay", "HealRay", "HarpoonRope",
}

type durationRayHooks48EA70 struct {
	connected func() bool
	byCode    func(uint16) *client.Drawable
	typeID    func(int, string) int
	spawn     func(int, image.Point) *client.Drawable
	remove    func(*client.Drawable)
}

func durationRayContainsDrawable48EA70(slots *[96]clientDurationRay48EA70, dr *client.Drawable) bool {
	if dr == nil {
		return false
	}
	for i := range slots {
		if slots[i].drawable == dr {
			return true
		}
	}
	return false
}

func (c *Client) isDurationRayDrawable48EA70(dr *client.Drawable) bool {
	return durationRayContainsDrawable48EA70(&c.fxDurationRays, dr)
}

// The ray draw callbacks consume an unaligned PE32 payload inside the
// drawable union, even though the union itself moves on native-width ABIs.
func setDurationRayDrawable48EA70(dr *client.Drawable, subtype byte, source, target uint16) {
	data := unsafe.Slice((*byte)(unsafe.Pointer(&dr.Union)), 13)
	data[0] = 1
	binary.LittleEndian.PutUint32(data[1:], uint32(subtype))
	binary.LittleEndian.PutUint32(data[5:], uint32(source))
	binary.LittleEndian.PutUint32(data[9:], uint32(target))
}

func handleDurationRayNative48EA70(data []byte, slots *[96]clientDurationRay48EA70, hooks durationRayHooks48EA70) int {
	if len(data) < 7 {
		return -1
	}
	kind := data[1]
	if kind < 1 || kind > 14 {
		return -1
	}
	if !hooks.connected() {
		return 7
	}
	source := binary.LittleEndian.Uint16(data[3:])
	target := binary.LittleEndian.Uint16(data[5:])
	if kind > 7 {
		// 004FEF90 sends the two object codes in the opposite order from
		// 004FF130. Accept that wire order first, then the PE32 client's
		// original same-order form for packets from other servers.
		for _, reversed := range [...]bool{true, false} {
			for i := range slots {
				slot := &slots[i]
				match := slot.source == source && slot.target == target
				if reversed {
					match = slot.source == target && slot.target == source
				}
				if slot.drawable != nil && slot.kind == kind-7 && match {
					hooks.remove(slot.drawable)
					*slot = clientDurationRay48EA70{}
					return 7
				}
			}
		}
		return 7
	}
	free := -1
	for i := range slots {
		if slots[i].drawable == nil {
			free = i
			break
		}
	}
	if free < 0 {
		return 7
	}
	from, to := hooks.byCode(source), hooks.byCode(target)
	if from == nil || to == nil {
		return 7
	}
	pos := image.Pt(
		from.PosVec.X+(to.PosVec.X-from.PosVec.X)/2,
		from.PosVec.Y+(to.PosVec.Y-from.PosVec.Y)/2,
	)
	index := int(kind - 1)
	dr := hooks.spawn(hooks.typeID(index, clientDurationRayNames48EA70[index]), pos)
	if dr == nil {
		return 7
	}
	setDurationRayDrawable48EA70(dr, data[2], source, target)
	slots[free] = clientDurationRay48EA70{drawable: dr, source: source, target: target, kind: kind}
	return 7
}

func (c *Client) handleDurationRayPacketNative48EA70(data []byte) int {
	return handleDurationRayNative48EA70(data, &c.fxDurationRays, durationRayHooks48EA70{
		connected: nox_client_isConnected,
		byCode:    c.Objs.ByNetCode,
		typeID: func(index int, name string) int {
			return resolvePointSpriteFXTypeNative48EA70(&c.fxDurationRayTypes[index], name, c.Things.IndByID)
		},
		spawn:  c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		remove: c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable,
	})
}
