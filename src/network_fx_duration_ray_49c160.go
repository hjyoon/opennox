package opennox

import (
	"encoding/binary"
	"image"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
)

var legacyDurationRayNames49C160 = [...]string{
	"DynamicLightning", "DynamicChainLightning", "DynamicEnergyBolt", "OrbRay",
	"PlasmaRay", "DrainManaRay", "HealRay", "CharmRay",
	"DrainManaOrb", "HealOrb", "CharmOrb", "HarpoonRope",
}

var legacyDurationRayTypeIndex49C160 = [...]int{4, 7, 1, 2, 5, 6, 11}

func addLegacyDurationRay49C160(packet [7]byte, slots *[96]clientDurationRay48EA70, cache *[12]int, count int32,
	lookup func(string) int, byCode func(uint16) *client.Drawable, spawn func(int, image.Point) *client.Drawable,
) *client.Drawable {
	// The PE32 entry checks the count before looking up any sprite types.
	if count >= int32(len(slots)) {
		return nil
	}
	if cache[0] == 0 {
		for i, name := range legacyDurationRayNames49C160 {
			cache[i] = lookup(name)
		}
	}
	source := binary.LittleEndian.Uint16(packet[3:])
	target := binary.LittleEndian.Uint16(packet[5:])
	from, to := byCode(source), byCode(target)
	if from == nil || to == nil {
		return nil
	}
	kind := packet[1]
	var index int
	switch {
	case kind >= 1 && int(kind) <= len(legacyDurationRayTypeIndex49C160):
		index = legacyDurationRayTypeIndex49C160[kind-1]
	case kind == 0x8c:
		index = 0
	default:
		return nil
	}
	x, y := int32(from.PosVec.X), int32(from.PosVec.Y)
	pos := image.Pt(int(x+(int32(to.PosVec.X)-x)/2), int(y+(int32(to.PosVec.Y)-y)/2))
	dr := spawn(cache[index], pos)
	if dr == nil {
		return nil
	}
	setDurationRayDrawable48EA70(dr, packet[2], source, target)
	// The stock entry assigns the first free slot after creating the sprite.
	for i := range slots {
		if slots[i].drawable == nil {
			slots[i] = clientDurationRay48EA70{drawable: dr, source: source, target: target, kind: kind}
			break
		}
	}
	return dr
}

func (c *Client) AddLegacyDurationRay(packet [7]byte) *client.Drawable {
	if c == nil || c.Client == nil {
		return nil
	}
	return addLegacyDurationRay49C160(packet, &c.fxDurationRays, &c.fxLegacyRayTypes, memmap.Int32(0x5D4594, 1304312),
		c.Things.IndByID, c.Objs.ByNetCode, c.Nox_xxx_spriteLoadAdd_45A360_drawable)
}

func removeLegacyDurationRay49C450(packet [7]byte, slots *[96]clientDurationRay48EA70, remove func(*client.Drawable)) {
	source := uint32(binary.LittleEndian.Uint16(packet[3:]))
	target := uint32(binary.LittleEndian.Uint16(packet[5:]))
	for i := range slots {
		slot := &slots[i]
		if slot.drawable == nil {
			continue
		}
		data := durationRayPayload48EA70(slot.drawable)
		if binary.LittleEndian.Uint32(data[5:]) == source && binary.LittleEndian.Uint32(data[9:]) == target {
			remove(slot.drawable)
			*slot = clientDurationRay48EA70{}
			return
		}
	}
}

func (c *Client) RemoveLegacyDurationRay(packet [7]byte) {
	if c != nil {
		removeLegacyDurationRay49C450(packet, &c.fxDurationRays, c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable)
	}
}

func clearDurationRays49C4F0(slots *[96]clientDurationRay48EA70, remove func(*client.Drawable)) {
	for i := range slots {
		if dr := slots[i].drawable; dr != nil {
			remove(dr)
			slots[i] = clientDurationRay48EA70{}
		}
	}
}

func (c *Client) ClearDurationRays() {
	if c != nil {
		clearDurationRays49C4F0(&c.fxDurationRays, c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable)
	}
}
