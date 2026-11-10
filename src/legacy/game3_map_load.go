package legacy

import (
	"encoding/binary"
	"image"
	"io"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

// Stream fields remain PE32 BYTE/WORD/DWORD, independently of native object
// and modifier pointer widths. Hooks expose only the real external services.
type game3MapServices struct {
	rw         func([]byte)
	crc        func([]byte)
	seek       func(int32)
	sprite     func(int32, image.Point) *client.Drawable
	failed     func()
	typeID     func(string) uint32
	team       func(*client.Drawable)
	modifier   func(string) unsafe.Pointer
	readOnly   func() bool
	class      func(uint16, int) bool
	toc        func(uint16) uint16
	floatInt   func(float32) int32
	attachWall func(*client.Drawable, int32, int32)
}

func game3MapRuntime() game3MapServices {
	return game3MapServices{
		rw: func(p []byte) { _, _ = cryptfile.Global().ReadWrite(p) }, crc: func(p []byte) { _ = cryptfile.Global().ReadMaybeAlign(p) },
		seek: func(n int32) { _ = cryptfile.Global().Seek(int64(n), io.SeekCurrent) },
		sprite: func(id int32, p image.Point) *client.Drawable {
			return GetClient().Nox_xxx_spriteLoadAdd_45A360_drawable(int(id), p)
		},
		failed: game3MapLoadError,
		typeID: func(name string) uint32 { return uint32(GetClient().Cli().Things.IndByID(name)) },
		team: func(d *client.Drawable) {
			Nox_xxx_createAtImpl_4191D0(server.TeamID(d.TeamVal.ID), &d.TeamVal, 0, int(d.NetCode32), 0)
		},
		modifier: func(name string) unsafe.Pointer {
			id := GetServer().S().Modif.Nox_xxx_modifGetIdByName413290(name)
			return unsafe.Pointer(GetServer().S().Modif.Nox_xxx_modifGetDescById413330(id))
		},
		readOnly: func() bool { return cryptfile.Global().ReadOnly() }, class: game3MapClass, toc: game3MapTOC, floatInt: game3MapFloatInt,
		attachWall: game3AttachWall,
	}
}

type game3MapStream struct {
	h     game3MapServices
	count *uint32
}

func (s game3MapStream) bytes(p []byte) { s.h.rw(p); *s.count += uint32(len(p)) }
func (s game3MapStream) u8() byte       { var p [1]byte; s.bytes(p[:]); return p[0] }
func (s game3MapStream) u16() uint16 {
	var p [2]byte
	s.bytes(p[:])
	return binary.LittleEndian.Uint16(p[:])
}
func (s game3MapStream) u32() uint32 {
	var p [4]byte
	s.bytes(p[:])
	return binary.LittleEndian.Uint32(p[:])
}
func (s game3MapStream) skip(n int32) { s.h.seek(n); *s.count += uint32(n) }
func (s game3MapStream) name() string {
	n := int(s.u8())
	p := make([]byte, n)
	s.bytes(p)
	for i, b := range p {
		if b == 0 {
			return string(p[:i])
		}
	}
	return string(p)
}
func game3MapFlags(d *client.Drawable, flags uint32) {
	d.ObjFlags = object.Flags(flags | (uint32(d.ObjFlags) & 0xEEBF7E9D))
}
func game3MapFlags70(d *client.Drawable, flags uint32) {
	v := d.Flags70Val
	v = (v & 0xffffff00) | (v & 0xa1)
	d.Flags70Val = flags | v
}

func game3LegacySpriteLoad(id int32, version, mapVersion int16, count *uint32, h game3MapServices) *client.Drawable {
	s := game3MapStream{h, count}
	code, flags := s.u32(), s.u32()
	var pos image.Point
	if mapVersion < 40 || version < 4 {
		pos = image.Pt(int(int32(s.u32())), int(int32(s.u32())))
	} else {
		x := math.Float32frombits(s.u32())
		y := math.Float32frombits(s.u32())
		pos = image.Pt(int(h.floatInt(x)), int(h.floatInt(y)))
	}
	if mapVersion >= 10 {
		_ = s.name()
	}
	// These locals are undefined in the old C for the older formats. Zero is
	// the explicit native default; no legacy stack garbage is reproduced.
	var team byte
	var flags70 uint32
	if mapVersion >= 20 {
		team = s.u8()
	}
	if mapVersion >= 30 {
		_ = s.u8()
	}
	if mapVersion >= 40 {
		s.skip(4)
		if version >= 2 {
			var n uint16
			if version < 5 {
				n = uint16(s.u32())
			} else {
				n = s.u16()
			}
			s.skip(int32(uint16(4 * n)))
		}
		if version >= 3 {
			flags70 = s.u32()
		}
	}
	d := h.sprite(id, pos)
	if d == nil {
		return nil
	}
	game3MapFlags(d, flags)
	game3MapFlags70(d, flags70)
	d.NetCode32 = code
	d.TeamVal.ID = server.TeamID(team)
	if !noxflags.HasGame(noxflags.GameFlag(1)) && team != 0 && (uint32(d.ObjClass)&0x10000000 == 0 || !noxflags.HasGame(noxflags.GameFlag(128))) {
		h.team(d)
	}
	return d
}

func game3SpriteLoad(id int32, mapVersion int16, count *uint32, h game3MapServices) *client.Drawable {
	s := game3MapStream{h, count}
	version := int16(1)
	if mapVersion >= 40 {
		version = int16(s.u16())
	}
	if version < 61 {
		return game3LegacySpriteLoad(id, version, mapVersion, count, h)
	}
	code := s.u32()
	_ = s.u32()
	x := math.Float32frombits(s.u32())
	y := math.Float32frombits(s.u32())
	// Preserve the two conversion calls' original Y-before-X order.
	iy := h.floatInt(y)
	ix := h.floatInt(x)
	d := h.sprite(id, image.Pt(int(ix), int(iy)))
	if d == nil {
		return nil
	}
	d.NetCode32 = code
	if s.u8() != 0 {
		game3MapFlags(d, s.u32())
		s.skip(int32(s.u8()))
		d.TeamVal.ID = server.TeamID(s.u8())
		_ = s.u8()
		s.skip(int32(4 * uint32(s.u16())))
		game3MapFlags70(d, s.u32())
		if version >= 63 {
			s.skip(2)
			s.skip(int32(s.u32()))
			s.skip(4)
		}
		if version >= 64 {
			s.skip(4)
		}
	}
	if !noxflags.HasGame(noxflags.GameFlag(1)) {
		p := memmap.PtrUint32(0x5D4594, 1309788)
		if *p == 0 {
			*p = h.typeID("FlagMarker")
		}
		if d.TeamVal.ID != 0 && d.TypeIDVal != *p {
			h.team(d)
		}
	}
	return d
}

func game3LoadObject(id int32, kind int, h game3MapServices) int32 {
	var count uint32 = 2
	s := game3MapStream{h, &count}
	version := int16(s.u16())
	count = 2
	d := game3SpriteLoad(id, version, &count, h)
	if d == nil {
		h.failed()
		return 0
	}
	switch kind {
	case 0: // movable/door-shaped object
		flags := uint32(d.ObjClass)
		if flags&0x80 == 0 {
			if flags&0x200 != 0 {
				w, ht := float32(int32(s.u32())), float32(int32(s.u32()))
				if w > 60 {
					w = 60
				}
				if ht > 60 {
					ht = 60
				}
				d.Shape.Box.W = w
				d.Shape.Box.H = ht
				d.Shape.Box.Calc()
			}
		} else {
			direction := s.u32()
			active := s.u32()
			v := (*[2]byte)(unsafe.Pointer(&d.Union))
			v[0] = 0
			if byte(active) != 0 {
				v[0] = 1
			}
			v[1] = byte(active)
			index := direction
			if version >= 41 {
				index = s.u32()
			}
			d.Field_74_4 = byte(direction)
			x := int32(d.PosVec.X) + memmap.Int32(0x587000, uintptr(uint32(196184+8*index)))/2
			y := int32(d.PosVec.Y) + memmap.Int32(0x587000, uintptr(uint32(196188+8*index)))/2
			h.attachWall(d, int32(x/23), int32(y/23))
		}
	case 1:
		game3ColorLight(d, version, s)
	case 2: // four modifier addresses use the native pointer-bearing union
		d.Field_72 = 0
		d.SetActive()
		data := (*client.DrawableUnionItem)(unsafe.Pointer(&d.Union))
		slots := []*unsafe.Pointer{&data.Field_108, &data.Field_109, &data.Field_110, &data.Field_111}
		for _, slot := range slots {
			*slot = h.modifier(s.name())
			data.Field_112_0 = -1
			data.Field_112_2 = -1
		}
		return int32(count)
	case 3:
		d.Field_72 = 0
		d.SetActive()
		d.Shape.Box.W = float32(int32(s.u32()))
		d.Shape.Box.H = float32(int32(s.u32()))
		d.Shape.Box.Calc()
		data := (*[6]byte)(unsafe.Pointer(&d.Union))
		copy(data[:], []byte{90, 90, 90, 10, 10, 10})
		if version >= 41 {
			for i := range data {
				data[i] = s.u8()
			}
		}
		return int32(count)
	case 4:
		if version >= 61 {
			_ = s.u32()
			if s.u8() == 1 {
				GetClient().Cli().Objs.MinimapAdd(d, 1)
			}
		}
	}
	d.Field_72 = 0
	d.SetActive()
	return int32(count)
}

func game3ColorLight(d *client.Drawable, version int16, s game3MapStream) {
	light := d.LightXferData()
	process := false
	for _, span := range [][2]int{{0, 4}, {4, 4}, {8, 4}, {12, 4}, {16, 12}, {28, 2}, {30, 2}, {32, 4}} {
		s.bytes(light[span[0] : span[0]+span[1]])
	}
	if version >= 2 {
		for _, span := range [][2]int{{40, 2}, {42, 48}, {90, 16}, {106, 16}, {122, 2}, {124, 2}, {126, 2}, {128, 4}, {134, 2}, {136, 2}, {138, 1}} {
			s.bytes(light[span[0] : span[0]+span[1]])
		}
		if version > 40 {
			if version >= 42 {
				s.bytes(light[36:40])
			} else {
				binary.LittleEndian.PutUint32(light[36:], uint32(s.u8()))
			}
		}
		process = s.h.readOnly()
	} else {
		for _, span := range [][2]int{{40, 2}, {122, 2}, {124, 2}, {126, 2}, {128, 4}, {134, 2}} {
			clear(light[span[0] : span[0]+span[1]])
		}
		light[138] = 128
		if s.h.readOnly() {
			if math.Float32frombits(binary.LittleEndian.Uint32(light[4:])) <= 63 && float64(int32(binary.LittleEndian.Uint32(light[12:])))*(*memmap.PtrFloat64(0x581450, 9752)) <= (*memmap.PtrFloat64(0x581450, 9744)) {
				process = true
			} else {
				invLightClampLegacy4F5AA0(&light, func(v float32) uint32 { return uint32(client.LightRadius(v)) })
				process = s.h.readOnly()
			}
		}
	}
	if process {
		data := (*[3]byte)(unsafe.Pointer(&d.Union))
		clear(data[:])
		for i := 0; i < 16; i++ {
			if light[42+3*i] != 0 || light[43+3*i] != 0 || light[44+3*i] != 0 {
				data[0]++
			}
			if light[90+i] != 0 {
				data[1]++
			}
			if light[106+i] != 0 {
				data[2]++
			}
		}
		a := float64(binary.LittleEndian.Uint16(light[28:]))
		n := int64(a * (*memmap.PtrFloat64(0x581450, 9752)) * (*memmap.PtrFloat64(0x581450, 9736)))
		binary.LittleEndian.PutUint16(light[132:], uint16(n))
	}
	d.ApplyLightXferData(&light)
}

func game3ClientLoadSome(id uint16, h game3MapServices) int32 {
	cells := []uintptr{1309792, 1309796, 1309800, 1309804}
	names := []string{"ColorLight", "ColorLightMovable", "TeamBase", "PressurePlate"}
	if memmap.Uint32(0x5D4594, cells[0]) == 0 {
		for i, off := range cells {
			*memmap.PtrUint32(0x5D4594, off) = h.typeID(names[i])
		}
	}
	kind := -1
	switch uint32(id) {
	case memmap.Uint32(0x5D4594, cells[0]), memmap.Uint32(0x5D4594, cells[1]):
		kind = 1
	case memmap.Uint32(0x5D4594, cells[2]):
		kind = 2
	case memmap.Uint32(0x5D4594, cells[3]):
		kind = 3
	default:
		if h.class(id, 0) {
			kind = 4
		} else if h.class(id, 1) {
			kind = 0
		}
	}
	if kind < 0 {
		return 0
	}
	return game3LoadObject(int32(id), kind, h)
}

func game3SpecialObjectData(h game3MapServices) int32 {
	var count uint32
	s := game3MapStream{h, &count}
	version := int16(s.u16())
	if version > 1 || !h.readOnly() {
		return 0
	}
	id := s.u16()
	for id != 0 {
		var data [4]byte
		h.crc(data[:])
		length := int32(binary.LittleEndian.Uint32(data[:]))
		tt := h.toc(id)
		if tt == 0 {
			return 0
		}
		if h.class(tt, 2) {
			length -= game3ClientLoadSome(tt, h)
		}
		if length > 0 {
			h.seek(length)
		}
		id = s.u16()
	}
	return 1
}
