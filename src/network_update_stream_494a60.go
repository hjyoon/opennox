package opennox

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netalias"
	"github.com/opennox/opennox/v1/internal/netlist"
)

type updateStreamHooks494A60 struct {
	frame        func() uint32
	lookupAlias  func(byte) netalias.Entry
	reserveAlias func(code, typeID uint16, now, deadline uint32) (byte, bool)
	sendAlias    func(alias byte, entry netalias.Entry)
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
	lookup func(byte) netalias.Entry,
	store func(byte, netalias.Entry),
) (byte, bool) {
	ind := netalias.FindAvailable(int32(code), int32(typeID), now, lookup)
	if ind == netalias.NoAlias {
		return ind, false
	}
	store(ind, netalias.Entry{Code: code, TypeID: typeID, Deadline: deadline})
	return ind, true
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
	if alias != netalias.NoAlias {
		entry := hooks.lookupAlias(alias)
		return entry.Code, entry.TypeID, true
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
		hooks.sendAlias(alias, netalias.Entry{
			Code: code, TypeID: typeID, Deadline: deadline,
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

func updateStreamRuntimeAliasTable494A60() netalias.Table {
	return memmap.Slice(0x5D4594, 1198020)[:netalias.TableSize]
}

func (c *Client) handleUpdateStreamPacketNative494A60(ind ntype.PlayerInd, data []byte) int {
	table := updateStreamRuntimeAliasTable494A60()
	return handleUpdateStreamNative494A60(data, updateStreamHooks494A60{
		frame:       c.Server.Frame,
		lookupAlias: table.Lookup,
		reserveAlias: func(code, typeID uint16, now, deadline uint32) (byte, bool) {
			return reserveUpdateStreamAlias494A60(code, typeID, now, deadline, table.Lookup, table.Store)
		},
		sendAlias: func(alias byte, entry netalias.Entry) {
			var packet [10]byte
			packet[0] = byte(netmsg.MSG_NEW_ALIAS)
			packet[1] = alias
			entry.PutRecord(packet[2:])
			c.srv.NetList.AddToMsgListCli(ind, netlist.Kind0, packet[:])
		},
		create:       c.Nox_xxx_spriteCreate_48E970,
		updateCamera: nox_xxx_cliUpdateCameraPos_435600,
		addToIndex2D: c.Objs.AddIndex2D,
	})
}
