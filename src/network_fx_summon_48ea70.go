package opennox

import (
	"encoding/binary"
	"image"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

type summonFXHooks48EA70 struct {
	connected    func() bool
	typeID       func(string) int
	spawnParent  func(int, image.Point) *client.Drawable
	newChild     func(int) *client.Drawable
	direction    func(byte) byte
	frame        func() uint32
	each         func(func(*client.Drawable))
	pointSpark   func(image.Point)
	deleteChild  func(*client.Drawable)
	deleteParent func(*client.Drawable)
}

func handleSummonFXNative48EA70(data []byte, hooks summonFXHooks48EA70) int {
	if len(data) < 12 {
		return -1
	}
	if !hooks.connected() {
		return 12
	}
	pos := image.Pt(
		int(binary.LittleEndian.Uint16(data[1:3])),
		int(binary.LittleEndian.Uint16(data[3:5])),
	)
	parent := hooks.spawnParent(hooks.typeID("SummonEffect"), pos)
	if parent == nil {
		return 12
	}
	child := hooks.newChild(int(binary.LittleEndian.Uint16(data[7:9])))
	if child == nil {
		return 12
	}
	child.PosVec = pos
	child.AnimDir = hooks.direction(data[9])
	child.AnimInd = 8
	state := parent.UnionSummon()
	state.Child = child
	state.Lifetime = binary.LittleEndian.Uint16(data[10:12])
	state.ID = binary.LittleEndian.Uint16(data[5:7])
	parent.AnimStart = hooks.frame()
	return 12
}

func handleSummonCancelFXNative48EA70(data []byte, hooks summonFXHooks48EA70) int {
	if len(data) < 3 {
		return -1
	}
	if !hooks.connected() {
		return 3
	}
	wantType := uint32(hooks.typeID("SummonEffect"))
	wantID := binary.LittleEndian.Uint16(data[1:3])
	var found *client.Drawable
	hooks.each(func(dr *client.Drawable) {
		if found == nil && dr != nil && dr.TypeIDVal == wantType && dr.UnionSummon().ID == wantID {
			found = dr
		}
	})
	if found == nil {
		return 3
	}
	hooks.pointSpark(found.PosVec)
	state := found.UnionSummon()
	if state.Child != nil {
		hooks.deleteChild(state.Child)
		state.Child = nil
	}
	hooks.deleteParent(found)
	return 3
}

func (c *Client) summonFXHooksNative48EA70() summonFXHooks48EA70 {
	return summonFXHooks48EA70{
		connected:   nox_client_isConnected,
		typeID:      c.Things.IndByID,
		spawnParent: c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		newChild:    c.Nox_new_drawable_for_thing,
		direction: func(v byte) byte {
			return byte(legacy.Nox_xxx_math_509EA0(int(v)))
		},
		frame: c.srv.Frame,
		each: func(fnc func(*client.Drawable)) {
			for _, dr := range c.Objs.AllList1() {
				fnc(dr)
			}
		},
		pointSpark: func(pos image.Point) {
			c.spawnPointSparkFXNative48EA70(pointSparkFXSpec48EA70{cache: 0, name: "BlueSpark", count: 50, speed: 1000, minLife: 30}, pos)
		},
		deleteChild: func(child *client.Drawable) {
			c.Nox_xxx_spriteDelete_45A4B0(child)
		},
		deleteParent: c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable,
	}
}

func (c *Client) handleSummonFXPacketNative48EA70(data []byte) int {
	return handleSummonFXNative48EA70(data, c.summonFXHooksNative48EA70())
}

func (c *Client) handleSummonCancelFXPacketNative48EA70(data []byte) int {
	return handleSummonCancelFXNative48EA70(data, c.summonFXHooksNative48EA70())
}
