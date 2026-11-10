package legacy

/*
#include "defs.h"
#include "GAME1.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME4_1.h"
#include "client__draw__fx.h"
extern uint32_t dword_5d4594_1313740, dword_5d4594_1313788;
extern uint32_t dword_5d4594_1313532, dword_5d4594_1313536, dword_5d4594_1313540;
*/
import "C"

import (
	"encoding/binary"
	"image"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type game3EffectServices struct {
	frame       func() uint32
	typeID      func(string) uint32
	load        func(uint32, image.Point) *client.Drawable
	newChild    func(uint16) *client.Drawable
	deleteSelf  func(*client.Drawable)
	deleteChild func(*client.Drawable)
	update      func(*client.Drawable, image.Point)
	first       func() *client.Drawable
	byCode      func(uint32) *client.Drawable
	each        func(image.Rectangle, func(*client.Drawable))
	pointFX     func(uint32, image.Point)
	direction   func(byte) byte
	render      Render2
	distance    func(int32, int32) uint32
}

func game3EffectsRuntime() game3EffectServices {
	cl := GetClient()
	c := cl.Cli()
	return game3EffectServices{
		frame: func() uint32 { return GetServer().S().Frame() }, typeID: func(s string) uint32 { return uint32(c.Things.IndByID(s)) },
		load: func(id uint32, p image.Point) *client.Drawable {
			return cl.Nox_xxx_spriteLoadAdd_45A360_drawable(int(id), p)
		}, newChild: func(id uint16) *client.Drawable { return cl.Nox_new_drawable_for_thing(int(id)) },
		deleteSelf: cl.Nox_xxx_spriteDeleteStatic_45A4E0_drawable, deleteChild: func(d *client.Drawable) { cl.Nox_xxx_spriteDelete_45A4B0(d) },
		update: func(d *client.Drawable, p image.Point) { c.Nox_xxx_updateSpritePosition_49AA90(d, p.X, p.Y) }, first: c.Objs.FirstList1,
		byCode: func(code uint32) *client.Drawable {
			if code&0x8000 != 0 {
				return c.Objs.ByNetCodeStatic(int(code & 0x7fff))
			}
			return c.Objs.ByNetCodeDynamic(int(code & 0x7fff))
		},
		each: c.Objs.EachInRect, pointFX: func(id uint32, p image.Point) {
			C.nox_xxx_makePointFxCli_499610(C.int(id), 50, 1000, 30, C.int(p.X), C.int(p.Y))
		},
		direction: func(v byte) byte { return byte(C.nox_xxx_math_509EA0(C.int(v))) }, render: cl.R2(), distance: func(x, y int32) uint32 { return uint32(C.sub_48C6B0(C.int(x), C.int(y))) },
	}
}

var game3EffectsFactory = game3EffectsRuntime

// game3Bounce keeps the signed BYTE velocity and signed WORD altitude domain.
func game3Bounce(d *client.Drawable, frame func() uint32) int16 {
	v := d.VelZ
	d.ZVal += uint16(int16(v))
	z := int16(d.ZVal)
	result := z
	if z >= 0 {
		if frame()&1 != 0 {
			d.VelZ = v - 1
		}
	} else {
		d.ZVal = uint16(-z)
		result = int16(26209 * int32(v))
		d.VelZ = int8(-9 * int32(v) / 10)
		if d.VelZ < 2 {
			d.ZVal = 0
			d.VelZ = 0
		}
	}
	return result
}

func game3SparkDraw(vp *noxrender.Viewport, d *client.Drawable, bright, dim uint32, h game3EffectServices) int32 {
	e := d.UnionEffect()
	duration := int32(e.Field_112 - e.Field_111)
	remain := int32(e.Field_112 - h.frame())
	if remain == duration {
		remain--
	}
	if remain <= 0 {
		h.deleteSelf(d)
		return 0
	}
	pos := vp.ToScreenPos(d.PosVec)
	pos.Y -= int(int16(d.ZVal2)) + int(int16(d.ZVal))
	if pos.X-10 >= vp.Screen.Min.X && pos.Y-10 >= vp.Screen.Min.Y && pos.X+10 < vp.Screen.Max.X && pos.Y+10 < vp.Screen.Max.Y {
		r := 4 * remain / duration
		h.render.DrawGlow(pos, noxcolor.RGBA5551(dim), int(2*r+1), int(int8(5*remain/duration)))
		h.render.Data().SetColor2(noxcolor.RGBA5551(bright))
		h.render.DrawPoint(pos, int(r), noxcolor.RGBA5551(bright))
	}
	return 1
}

//export sub_4B6880
func sub_4B6880(vp *C.uint32_t, dr *C.nox_drawable, bright, dim C.int) C.int {
	return C.int(game3SparkDraw((*noxrender.Viewport)(unsafe.Pointer(vp)), asDrawable(dr), uint32(bright), uint32(dim), game3EffectsFactory()))
}

//export sub_4B6970
func sub_4B6970(vp *C.uint32_t, dr *C.nox_drawable, bright, dim C.int) C.int {
	d := asDrawable(dr)
	h := game3EffectsFactory()
	e := d.UnionEffect()
	speed := int32(e.Field_110)
	off := uintptr(192088 + 8*uint32(d.Field_74_4))
	e.Field_108 += uint32(speed * memmap.Int32(0x587000, off))
	e.Field_109 += uint32(speed * memmap.Int32(0x587000, off+4))
	h.update(d, image.Pt(int(e.Field_108>>12), int(int32(e.Field_109)>>12)))
	game3Bounce(d, h.frame)
	return C.int(game3SparkDraw((*noxrender.Viewport)(unsafe.Pointer(vp)), d, uint32(bright), uint32(dim), h))
}

//export sub_4B69F0
func sub_4B69F0(dr *C.nox_drawable) C.short {
	return C.short(game3Bounce(asDrawable(dr), game3EffectsFactory().frame))
}

func game3GlowColors(id uint32, h game3EffectServices) (uint32, uint32) {
	offsets := []uintptr{1313660, 1313664, 1313668, 1313672, 1313676, 1313680, 1313684}
	names := []string{"DrainManaOrb", "HealOrb", "CharmOrb", "WhiteOrb", "ManaBombOrb", "WhiteMoveOrb", "BlueMoveOrb"}
	if memmap.Uint32(0x5D4594, offsets[0]) == 0 {
		for i, off := range offsets {
			*memmap.PtrUint32(0x5D4594, off) = h.typeID(names[i])
		}
	}
	switch id {
	case memmap.Uint32(0x5D4594, offsets[0]), memmap.Uint32(0x5D4594, offsets[6]):
		return uint32(C.dword_5d4594_1313536), uint32(C.dword_5d4594_1313540)
	case memmap.Uint32(0x5D4594, offsets[2]):
		return memmap.Uint32(0x5D4594, 1313580), memmap.Uint32(0x5D4594, 1313584)
	case memmap.Uint32(0x5D4594, offsets[3]), memmap.Uint32(0x5D4594, offsets[4]), memmap.Uint32(0x5D4594, offsets[5]):
		return memmap.Uint32(0x5D4594, 1313588), memmap.Uint32(0x5D4594, 1313592)
	default:
		return memmap.Uint32(0x5D4594, 1313528), uint32(C.dword_5d4594_1313532)
	}
}

func game3GlowDraw(vp *noxrender.Viewport, d *client.Drawable, moving bool, h game3EffectServices) int32 {
	dim, bright := game3GlowColors(d.TypeIDVal, h)
	data := (*[20]byte)(unsafe.Pointer(&d.Union))
	if moving {
		dx := int32(binary.LittleEndian.Uint16(data[0:2])) - int32(d.PosVec.X)
		dy := int32(binary.LittleEndian.Uint16(data[2:4])) - int32(d.PosVec.Y)
		distance := h.distance(dx, dy) + 1
		if distance <= 10 {
			h.deleteSelf(d)
			return 0
		}
		speed := int32(data[11])
		h.update(d, image.Pt(int(int32(d.PosVec.X)+dx*speed/int32(distance)), int(int32(d.PosVec.Y)+dy*speed/int32(distance))))
	}
	pos := vp.ToScreenPos(d.PosVec)
	pos.Y -= 22
	r := int(data[12])
	if pos.X-r < vp.Screen.Min.X || pos.Y-r < vp.Screen.Min.Y || pos.X+r >= vp.Screen.Max.X || pos.Y+r >= vp.Screen.Max.Y {
		return 1
	}
	h.render.DrawGlow(pos, noxcolor.RGBA5551(dim), r, 5)
	h.render.Data().SetColor2(noxcolor.RGBA5551(bright))
	h.render.DrawPoint(pos, r>>1, noxcolor.RGBA5551(bright))
	h.render.Data().SetColor2(noxcolor.RGBA5551(bright))
	h.render.AddPoint(pos)
	h.render.AddPointRel(image.Pt(int(int32(d.Field_8)-int32(d.PosVec.X)), int(int32(d.Field_9)-int32(d.PosVec.Y))))
	h.render.DrawLineFromPoints(noxcolor.RGBA5551(bright))
	if data[13] != 0 {
		data[14]--
		if data[14] == 0 {
			data[14] = data[13]
			data[12]--
			if data[12] == 0 {
				h.deleteSelf(d)
				return 0
			}
		}
	}
	return 1
}

//export sub_4B6B80
func sub_4B6B80(vp *C.int, dr *C.nox_drawable, moving C.int) C.int {
	return C.int(game3GlowDraw((*noxrender.Viewport)(unsafe.Pointer(vp)), asDrawable(dr), moving != 0, game3EffectsFactory()))
}

func game3Summon(id int16, pos image.Point, thing uint16, direction byte, lifetime int16, h game3EffectServices) uintptr {
	d := h.load(h.typeID("SummonEffect"), pos)
	if d == nil {
		return 0
	}
	child := h.newChild(thing)
	if child == nil {
		return 0
	}
	child.PosVec = pos
	child.AnimDir = h.direction(direction)
	child.AnimInd = 8
	data := d.UnionSummon()
	data.Child = child
	data.Lifetime = uint16(lifetime)
	data.ID = uint16(id)
	result := h.frame()
	d.AnimStart = h.frame()
	return uintptr(result)
}

//export game3_summon_packet_native
func game3_summon_packet_native(id C.short, xy *C.ushort, thing C.ushort, direction C.uchar, lifetime C.short) C.uintptr_t {
	raw := unsafe.Slice((*byte)(unsafe.Pointer(xy)), 4)
	pos := image.Pt(int(binary.LittleEndian.Uint16(raw)), int(binary.LittleEndian.Uint16(raw[2:])))
	return C.uintptr_t(game3Summon(int16(id), pos, uint16(thing), byte(direction), int16(lifetime), game3EffectsFactory()))
}

//export sub_4B7EE0
func sub_4B7EE0(id C.short) {
	h := game3EffectsFactory()
	p := memmap.PtrUint32(0x5D4594, 1313744)
	if *p == 0 {
		*p = h.typeID("SummonEffect")
	}
	if C.dword_5d4594_1313740 == 0 {
		C.dword_5d4594_1313740 = C.uint32_t(h.typeID("BlueSpark"))
	}
	for d := h.first(); d != nil; d = d.Next() {
		if d.TypeIDVal == *p && d.UnionSummon().ID == uint16(id) {
			h.pointFX(uint32(C.dword_5d4594_1313740), d.PosVec)
			h.deleteChild(d.UnionSummon().Child)
			h.deleteSelf(d)
			return
		}
	}
}

//export nox_xxx_spriteShieldLoad_4B7F90
func nox_xxx_spriteShieldLoad_4B7F90() C.int {
	h := game3EffectsFactory()
	names := []string{"SphericalShieldNW", "SphericalShieldN", "SphericalShieldNE", "SphericalShieldW", "", "SphericalShieldE", "SphericalShieldSW", "SphericalShieldS", "SphericalShieldSE"}
	var last uint32
	for i, name := range names {
		if name == "" {
			continue
		}
		last = h.typeID(name)
		*memmap.PtrUint32(0x5D4594, uintptr(1313748+4*i)) = last
	}
	*memmap.PtrUint32(0x5D4594, 1313764) = 0
	*memmap.PtrUint32(0x5D4594, 1313784) = 1
	return C.int(last)
}
func game3ShieldScan(d *client.Drawable, code uint32) {
	for i := 0; i < 9; i++ {
		if d.TypeIDVal == memmap.Uint32(0x5D4594, uintptr(1313748+4*i)) && d.UnionEffect().Field_108 == code {
			C.dword_5d4594_1313788 = 1
		}
	}
}

//export nox_xxx_spriteScanForShield_4B81E0
func nox_xxx_spriteScanForShield_4B81E0(dr *C.nox_drawable, data unsafe.Pointer) {
	game3ShieldScan(asDrawable(dr), *(*uint32)(data))
}
func game3Shield(code uint32, direction int32, h game3EffectServices) uintptr {
	if memmap.Uint32(0x5D4594, 1313784) == 0 {
		nox_xxx_spriteShieldLoad_4B7F90()
	}
	names := map[int32]string{0: "SphericalShieldNW", 1: "ShpericalShieldN", 2: "SphericalShieldNW", 3: "SphericalShieldW", 5: "SphericalShieldE", 6: "SphericalShieldSW", 7: "SphericalShieldS", 8: "SphericalShieldSE"}
	if name := names[direction]; name != "" {
		h.typeID(name)
	}
	d := h.byCode(code)
	if d == nil {
		return 0
	}
	p := d.PosVec
	rect := image.Rect(p.X-10, p.Y-10, p.X+10, p.Y+10)
	C.dword_5d4594_1313788 = 0
	h.each(rect, func(other *client.Drawable) { game3ShieldScan(other, code) })
	if C.dword_5d4594_1313788 == 1 {
		return 1
	}
	shield := h.load(memmap.Uint32(0x5D4594, uintptr(uint32(1313748+4*direction))), image.Pt(p.X, p.Y+3))
	if shield == nil {
		return 0
	}
	shield.UnionEffect().Field_108 = code
	return uintptr(shield.C())
}

//export game3_fx_shield_native
func game3_fx_shield_native(code C.uint, direction C.int) C.uintptr_t {
	return C.uintptr_t(game3Shield(uint32(code), int32(direction), game3EffectsFactory()))
}

type game3MaterialServices struct {
	definition func(uint32) *server.Modifier
	set        func(int, int, int, int)
}

func game3MaterialRuntime() game3MaterialServices {
	return game3MaterialServices{definition: func(bit uint32) *server.Modifier {
		s := GetServer().S()
		return s.Modif.Nox_xxx_getProjectileClassById413250(s.Weapons.Sub_415840(bit))
	}, set: func(i, r, g, b int) { GetClient().R2().Data().SetMaterialRGB(i, r, g, b) }}
}

var game3MaterialFactory = game3MaterialRuntime

func game3EquipmentMaterials(entries *[server.PlayerWeaponCnt]server.EquipmentData, token uintptr, h game3MaterialServices) uintptr {
	var found *server.EquipmentData
	for i := range entries {
		if uintptr(entries[i].Field0) == token {
			found = &entries[i]
			break
		}
	}
	if found == nil {
		return uintptr(unsafe.Add(unsafe.Pointer(entries), unsafe.Sizeof(*entries)))
	}
	def := h.definition(uint32(token))
	if def == nil {
		return 0
	}
	for i := 1; i <= 6; i++ {
		c := def.Colors12[i]
		h.set(i, int(c.R), int(c.G), int(c.B))
	}
	inds := def.ColorIndexes()
	var result unsafe.Pointer
	for i, p := range found.Field4 {
		result = p
		if p != nil {
			c := (*server.ModifierEff)(p).Color24
			h.set(int(inds[i]), int(c.R), int(c.G), int(c.B))
		}
	}
	return uintptr(result)
}

//export game3_equipment_materials_native
func game3_equipment_materials_native(entries unsafe.Pointer, token C.uintptr_t) C.uintptr_t {
	return C.uintptr_t(game3EquipmentMaterials((*[server.PlayerWeaponCnt]server.EquipmentData)(entries), uintptr(token), game3MaterialFactory()))
}
