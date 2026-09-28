package opennox

import (
	"encoding/binary"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

const (
	creatureAcquirePacketSize48EA70   = 5
	creatureLosePacketSize48EA70      = 3
	creatureMonitorPacketSize48EA70   = 5
	creatureUnmonitorPacketSize48EA70 = 3
	interestingIDPacketSize48EA70     = 7
)

type creatureReference48EA70 struct {
	RawCode uint16
	Code    uint16
	TypeID  uint16
}

type creatureAcquireState48EA70 struct {
	creatureReference48EA70
	Silent bool
}

type interestingIDState48EA70 struct {
	creatureReference48EA70
	Action byte
	Flags  byte
}

type creatureTrackingHooks48EA70 struct {
	connected     func() bool
	byStatic      func(uint16) *client.Drawable
	byDynamic     func(uint16) *client.Drawable
	create        func(uint16, uint16) *client.Drawable
	minimapAdd    func(*client.Drawable, byte)
	minimapRemove func(*client.Drawable, byte)
	summonAcquire func(uint16, uint16, bool)
	summonLose    func(uint16, bool)
	monitorAdd    func(uint16)
	monitorRemove func(uint16)
}

func decodeCreatureAcquireState48EA70(data []byte) (creatureAcquireState48EA70, bool) {
	if len(data) < creatureAcquirePacketSize48EA70 {
		return creatureAcquireState48EA70{}, false
	}
	rawCode := binary.LittleEndian.Uint16(data[1:3])
	rawType := binary.LittleEndian.Uint16(data[3:5])
	return creatureAcquireState48EA70{
		creatureReference48EA70: creatureReference48EA70{
			RawCode: rawCode,
			Code:    nox_xxx_netClearHighBit_578B30(rawCode),
			TypeID:  nox_xxx_netClearHighBit_578B30(rawType),
		},
		Silent: nox_xxx_netTestHighBit_578B70(rawType),
	}, true
}

func decodeCreatureReference48EA70(data []byte, size int) (creatureReference48EA70, bool) {
	if len(data) < size {
		return creatureReference48EA70{}, false
	}
	rawCode := binary.LittleEndian.Uint16(data[1:3])
	state := creatureReference48EA70{
		RawCode: rawCode,
		Code:    nox_xxx_netClearHighBit_578B30(rawCode),
	}
	if size >= 5 {
		state.TypeID = binary.LittleEndian.Uint16(data[3:5])
	}
	return state, true
}

func decodeInterestingIDState48EA70(data []byte) (interestingIDState48EA70, bool) {
	ref, ok := decodeCreatureReference48EA70(data, interestingIDPacketSize48EA70)
	if !ok {
		return interestingIDState48EA70{}, false
	}
	return interestingIDState48EA70{
		creatureReference48EA70: ref,
		Action:                  data[5],
		Flags:                   data[6],
	}, true
}

func lookupCreatureDrawableNative48EA70(ref creatureReference48EA70, hooks creatureTrackingHooks48EA70) *client.Drawable {
	if nox_xxx_netTestHighBit_578B70(ref.RawCode) {
		return hooks.byStatic(ref.Code)
	}
	return hooks.byDynamic(ref.Code)
}

func handleCreatureAcquireNative48EA70(data []byte, hooks creatureTrackingHooks48EA70) int {
	state, ok := decodeCreatureAcquireState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return creatureAcquirePacketSize48EA70
	}
	hooks.summonAcquire(state.RawCode, state.TypeID, state.Silent)
	dr := lookupCreatureDrawableNative48EA70(state.creatureReference48EA70, hooks)
	if dr == nil {
		dr = hooks.create(state.TypeID, state.Code)
	}
	if dr != nil {
		hooks.minimapAdd(dr, 1)
	}
	hooks.monitorAdd(state.RawCode)
	return creatureAcquirePacketSize48EA70
}

func handleCreatureLoseNative48EA70(data []byte, hooks creatureTrackingHooks48EA70) int {
	state, ok := decodeCreatureReference48EA70(data, creatureLosePacketSize48EA70)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return creatureLosePacketSize48EA70
	}
	hooks.summonLose(state.Code, nox_xxx_netTestHighBit_578B70(state.RawCode))
	hooks.monitorRemove(state.Code)
	// GAME.EXE clears the high bit in the packet before this lookup, so lose
	// notifications always address the dynamic drawable namespace.
	if dr := hooks.byDynamic(state.Code); dr != nil {
		hooks.minimapRemove(dr, 1)
	}
	return creatureLosePacketSize48EA70
}

func handleCreatureMonitorNative48EA70(data []byte, hooks creatureTrackingHooks48EA70) int {
	state, ok := decodeCreatureReference48EA70(data, creatureMonitorPacketSize48EA70)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return creatureMonitorPacketSize48EA70
	}
	dr := lookupCreatureDrawableNative48EA70(state, hooks)
	if dr == nil {
		dr = hooks.create(state.TypeID, state.Code)
	}
	if dr != nil {
		hooks.minimapAdd(dr, 1)
	}
	hooks.monitorAdd(state.Code)
	return creatureMonitorPacketSize48EA70
}

func handleCreatureUnmonitorNative48EA70(data []byte, hooks creatureTrackingHooks48EA70) int {
	state, ok := decodeCreatureReference48EA70(data, creatureUnmonitorPacketSize48EA70)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return creatureUnmonitorPacketSize48EA70
	}
	hooks.monitorRemove(state.RawCode)
	if dr := lookupCreatureDrawableNative48EA70(state, hooks); dr != nil {
		hooks.minimapRemove(dr, 1)
	}
	return creatureUnmonitorPacketSize48EA70
}

func handleInterestingIDNative48EA70(data []byte, hooks creatureTrackingHooks48EA70) int {
	state, ok := decodeInterestingIDState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return interestingIDPacketSize48EA70
	}
	dr := lookupCreatureDrawableNative48EA70(state.creatureReference48EA70, hooks)
	if state.Action == 1 {
		if dr == nil {
			dr = hooks.create(state.TypeID, state.Code)
		}
		if dr != nil {
			hooks.minimapAdd(dr, state.Flags)
		}
	} else if dr != nil {
		hooks.minimapRemove(dr, state.Flags)
	}
	return interestingIDPacketSize48EA70
}

func (c *Client) creatureTrackingHooksNative48EA70() creatureTrackingHooks48EA70 {
	return creatureTrackingHooks48EA70{
		connected: nox_client_isConnected,
		byStatic: func(code uint16) *client.Drawable {
			return c.Objs.ByNetCodeStatic(int(code))
		},
		byDynamic: func(code uint16) *client.Drawable {
			return c.Objs.ByNetCodeDynamic(int(code))
		},
		create: func(typeID, code uint16) *client.Drawable {
			return c.Nox_xxx_spriteCreate_48E970(int(typeID), code, 0, 0)
		},
		minimapAdd:    c.Objs.MinimapAdd,
		minimapRemove: c.Objs.RemoveHealthBar,
		summonAcquire: func(code, typeID uint16, silent bool) {
			legacy.Nox_xxx_cliSummonCreat_4C2E50(int(code), int(typeID), silent)
		},
		summonLose: func(code uint16, silent bool) {
			legacy.Nox_xxx_cliSummonOnDieOrBanish_4C3140(int(code), silent)
		},
		monitorAdd: func(code uint16) {
			legacy.Sub_495060(int(code), 0, 0)
		},
		monitorRemove: func(code uint16) {
			legacy.Sub_4950C0(int(code))
		},
	}
}

func (c *Client) handleCreatureAcquirePacketNative48EA70(data []byte) int {
	return handleCreatureAcquireNative48EA70(data, c.creatureTrackingHooksNative48EA70())
}

func (c *Client) handleCreatureLosePacketNative48EA70(data []byte) int {
	return handleCreatureLoseNative48EA70(data, c.creatureTrackingHooksNative48EA70())
}

func (c *Client) handleCreatureMonitorPacketNative48EA70(data []byte) int {
	return handleCreatureMonitorNative48EA70(data, c.creatureTrackingHooksNative48EA70())
}

func (c *Client) handleCreatureUnmonitorPacketNative48EA70(data []byte) int {
	return handleCreatureUnmonitorNative48EA70(data, c.creatureTrackingHooksNative48EA70())
}

func (c *Client) handleInterestingIDPacketNative48EA70(data []byte) int {
	return handleInterestingIDNative48EA70(data, c.creatureTrackingHooksNative48EA70())
}
