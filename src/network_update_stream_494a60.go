package opennox

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

const (
	updateStreamAliasCount494A60 = 255
	updateStreamAliasSize494A60  = 8
)

type updateStreamAliasEntry494A60 struct {
	code     uint16
	typeID   uint16
	deadline uint32
}

type updateStreamHooks494A60 struct {
	frame        func() uint32
	lookupAlias  func(byte) updateStreamAliasEntry494A60
	reserveAlias func(code, typeID uint16, now, deadline uint32) (byte, bool)
	sendAlias    func(alias byte, entry updateStreamAliasEntry494A60)
	create       func(typeID int, code uint16, x, y int) *client.Drawable
	updateCamera func(x, y int)
	addToIndex2D func(*client.Drawable)
}

type updateStreamReader494A60 struct {
	data []byte
	off  int
}

func (r *updateStreamReader494A60) readByte() (byte, bool) {
	if r.off >= len(r.data) {
		return 0, false
	}
	v := r.data[r.off]
	r.off++
	return v, true
}

func (r *updateStreamReader494A60) readUint16() (uint16, bool) {
	if len(r.data)-r.off < 2 {
		return 0, false
	}
	v := binary.LittleEndian.Uint16(r.data[r.off:])
	r.off += 2
	return v, true
}

func (r *updateStreamReader494A60) readInt16() (int, bool) {
	v, ok := r.readUint16()
	return int(int16(v)), ok
}

func reserveUpdateStreamAlias494A60(
	code, typeID uint16,
	now, deadline uint32,
	lookup func(byte) updateStreamAliasEntry494A60,
	store func(byte, updateStreamAliasEntry494A60),
) (byte, bool) {
	ind := byte(code)
	if ind == 0 || ind == 0xff {
		ind = 1
	}
	start := ind
	for {
		entry := lookup(ind)
		if (entry.code == code && entry.typeID == typeID) || entry.deadline < now {
			store(ind, updateStreamAliasEntry494A60{code: code, typeID: typeID, deadline: deadline})
			return ind, true
		}
		ind++
		if ind == 0xff {
			ind = 1
		}
		if ind == start {
			return 0xff, false
		}
	}
}

func readUpdateStreamAlias494A60(
	r *updateStreamReader494A60,
	hooks updateStreamHooks494A60,
	permanent bool,
) (code, typeID uint16, ok bool) {
	alias, ok := r.readByte()
	if !ok {
		return 0, 0, false
	}
	if alias != 0xff {
		entry := hooks.lookupAlias(alias)
		return entry.code, entry.typeID, true
	}
	code, ok = r.readUint16()
	if !ok {
		return 0, 0, false
	}
	typeID, ok = r.readUint16()
	if !ok {
		return 0, 0, false
	}
	now := hooks.frame()
	deadline := now + 60
	if permanent {
		deadline = ^uint32(0)
	}
	alias, ok = hooks.reserveAlias(code, typeID, now, deadline)
	if ok {
		hooks.sendAlias(alias, updateStreamAliasEntry494A60{
			code: code, typeID: typeID, deadline: deadline,
		})
	}
	return code, typeID, true
}

func updateStreamDirection494A60(control byte) byte {
	dir := (control >> 4) & 7
	if dir > 3 {
		dir++
	}
	return dir
}

func applyFirstUpdateStreamDrawable494A60(
	dr *client.Drawable,
	control, frame, animation byte,
	hooks updateStreamHooks494A60,
) {
	dr.Field_72 = hooks.frame()
	dr.AnimDir = updateStreamDirection494A60(control)
	if dr.AnimInd != uint32(animation) {
		dr.AnimInd = uint32(animation)
		dr.AnimStart = hooks.frame()
	}
	dr.SetFrameMB(int(frame))
}

func handleUpdateStreamNative494A60(data []byte, hooks updateStreamHooks494A60) int {
	if len(data) < 2 || netmsg.Op(data[0]) != netmsg.MSG_UPDATE_STREAM {
		return 0
	}
	r := updateStreamReader494A60{data: data, off: 1}
	code, typeID, ok := readUpdateStreamAlias494A60(&r, hooks, true)
	if !ok {
		return 0
	}
	x, ok := r.readUint16()
	if !ok {
		return 0
	}
	y, ok := r.readUint16()
	if !ok {
		return 0
	}
	control, ok := r.readByte()
	if !ok {
		return 0
	}
	var drawableFrame byte
	if control&0x80 != 0 {
		drawableFrame, ok = r.readByte()
		if !ok {
			return 0
		}
	}
	animation, ok := r.readByte()
	if !ok {
		return 0
	}
	if code != 0 || typeID != 0 {
		if dr := hooks.create(int(typeID), code, int(x), int(y)); dr != nil {
			applyFirstUpdateStreamDrawable494A60(dr, control, drawableFrame, animation, hooks)
		}
	}
	hooks.updateCamera(int(x), int(y))
	if r.off >= len(data) {
		return 0
	}

	curX, curY := int(x), int(y)
	for {
		if len(data)-r.off < 3 {
			return 0
		}
		if data[r.off] == 0 && data[r.off+1] == 0 && data[r.off+2] == 0 {
			return r.off + 3
		}

		fullPosition := data[r.off] == 0
		if fullPosition {
			r.off++
		}
		code, typeID, ok = readUpdateStreamAlias494A60(&r, hooks, false)
		if !ok {
			return 0
		}
		if fullPosition {
			curX, ok = r.readInt16()
			if !ok {
				return 0
			}
			curY, ok = r.readInt16()
			if !ok {
				return 0
			}
		} else {
			dx, ok := r.readByte()
			if !ok {
				return 0
			}
			dy, ok := r.readByte()
			if !ok {
				return 0
			}
			curX += int(int8(dx))
			curY += int(int8(dy))
		}
		if curX < 0 || curX > 6000 || curY < 0 || curY > 6000 {
			return r.off
		}

		dr := hooks.create(int(typeID), code, curX, curY)
		if dr == nil {
			return r.off
		}
		if !dr.Class().Has(object.ClassComplex) {
			dr.Field_72 = hooks.frame()
			hooks.addToIndex2D(dr)
			curX, curY = dr.PosVec.X, dr.PosVec.Y
			continue
		}

		control, ok = r.readByte()
		if !ok {
			return 0
		}
		var hasFrame bool
		if control&0x80 != 0 {
			drawableFrame, ok = r.readByte()
			if !ok {
				return 0
			}
			hasFrame = true
		}
		animation = control & 0x0f
		if dr.Class().Has(object.ClassPlayer) {
			animation, ok = r.readByte()
			if !ok {
				return 0
			}
		}
		dr.Field_72 = hooks.frame()
		dr.AnimDir = updateStreamDirection494A60(control)
		if hasFrame {
			dr.SetFrameMB(int(drawableFrame))
		}
		if dr.AnimInd != uint32(animation) {
			dr.AnimInd = uint32(animation)
			dr.AnimStart = hooks.frame()
		}
		curX, curY = dr.PosVec.X, dr.PosVec.Y
	}
}

type updateStreamAliasTable494A60 []byte

func updateStreamRuntimeAliasTable494A60() updateStreamAliasTable494A60 {
	data := memmap.Slice(0x5D4594, 1198020)
	return data[:updateStreamAliasCount494A60*updateStreamAliasSize494A60]
}

func (t updateStreamAliasTable494A60) lookup(ind byte) updateStreamAliasEntry494A60 {
	off := int(ind) * updateStreamAliasSize494A60
	return updateStreamAliasEntry494A60{
		code:     binary.LittleEndian.Uint16(t[off:]),
		typeID:   binary.LittleEndian.Uint16(t[off+2:]),
		deadline: binary.LittleEndian.Uint32(t[off+4:]),
	}
}

func (t updateStreamAliasTable494A60) store(ind byte, entry updateStreamAliasEntry494A60) {
	off := int(ind) * updateStreamAliasSize494A60
	binary.LittleEndian.PutUint16(t[off:], entry.code)
	binary.LittleEndian.PutUint16(t[off+2:], entry.typeID)
	binary.LittleEndian.PutUint32(t[off+4:], entry.deadline)
}

func (c *Client) handleUpdateStreamPacketNative494A60(ind ntype.PlayerInd, data []byte) int {
	table := updateStreamRuntimeAliasTable494A60()
	return handleUpdateStreamNative494A60(data, updateStreamHooks494A60{
		frame:       c.Server.Frame,
		lookupAlias: table.lookup,
		reserveAlias: func(code, typeID uint16, now, deadline uint32) (byte, bool) {
			return reserveUpdateStreamAlias494A60(code, typeID, now, deadline, table.lookup, table.store)
		},
		sendAlias: func(alias byte, entry updateStreamAliasEntry494A60) {
			var packet [10]byte
			packet[0] = byte(netmsg.MSG_NEW_ALIAS)
			packet[1] = alias
			binary.LittleEndian.PutUint16(packet[2:], entry.code)
			binary.LittleEndian.PutUint16(packet[4:], entry.typeID)
			binary.LittleEndian.PutUint32(packet[6:], entry.deadline)
			c.srv.NetList.AddToMsgListCli(ind, netlist.Kind0, packet[:])
		},
		create:       c.Nox_xxx_spriteCreate_48E970,
		updateCamera: nox_xxx_cliUpdateCameraPos_435600,
		addToIndex2D: c.Objs.AddIndex2D,
	})
}
