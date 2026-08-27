package legacy

/*
#include "GAME1_1.h"
#include "GAME2_1.h"
#include "GAME3_2.h"
#include "GAME5_2.h"
#include "client__gui__guiinv.h"
#include "common__magic__speltree.h"
#include "server__ability__ability.h"
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

const playerInfoPayloadSize41A590 = 97

// playerAttribWriteNative41A590 preserves version 5 of GAME.EXE's fixed-width
// player attribute stream. PlayerInfo itself has no native pointers, while the
// PlayerUpdateData.Player link used for the trailing fields is resolved at the
// host pointer width.
func playerAttribWriteNative41A590(cf *cryptfile.CryptFile, unit *server.Object, info *server.PlayerInfo) error {
	if cf == nil || info == nil {
		return fmt.Errorf("missing crypt file or player info")
	}
	raw := unsafe.Slice((*byte)(info.C()), playerInfoPayloadSize41A590)
	nameLen := 0
	for nameLen < 25 && binary.LittleEndian.Uint16(raw[2*nameLen:]) != 0 {
		nameLen++
	}
	if nameLen >= 25 {
		return fmt.Errorf("player name is not terminated within 24 UTF-16 code units")
	}

	mode := uint32(2)
	if noxflags.HasGame(noxflags.GameModeQuest) {
		mode = 4
	} else if noxflags.HasGame(noxflags.GameModeCoop) {
		mode--
	}
	if err := cf.WriteU16(5); err != nil {
		return err
	}
	if err := cf.WriteU32(mode); err != nil {
		return err
	}
	if err := cf.WriteU8(byte(nameLen)); err != nil {
		return err
	}
	if _, err := cf.Write(raw[:2*nameLen]); err != nil {
		return err
	}
	// GAME.EXE writes every byte from field2235 through Field2273, although
	// the original calls split the color triples into separate fragments.
	if _, err := cf.Write(raw[50:89]); err != nil {
		return err
	}

	var extraLives uint32
	var pl *server.Player
	if unit != nil && unit.UpdateData != nil {
		ud := (*server.PlayerUpdateData)(unit.UpdateData)
		extraLives = ud.ExtraLives
		pl = ud.Player
	}
	if err := cf.WriteU32(extraLives); err != nil {
		return err
	}
	var questStage uint32
	if pl != nil {
		questStage = pl.QuestStage
	}
	return cf.WriteU32(questStage)
}

type playerAttribReadHooks41A590 struct {
	questMode      func() bool
	disconnect     func(byte)
	protectName    func(uint32, []byte)
	protectInt     func(uint32, uint32)
	protectClass   func(uint32, byte)
	initColors     func(uint32)
	maxExtraLives  func() int32
	resetQuest     func(*server.Object)
	sendQuestStage func(*server.Player)
}

func playerAttribReadExact41A590(cf *cryptfile.CryptFile, dst []byte) error {
	if len(dst) == 0 {
		return nil
	}
	n, err := cf.Read(dst)
	if err != nil {
		return err
	}
	if n != len(dst) {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// playerAttribReadNative41A590 preserves the fixed-width file representation
// while resolving Object, PlayerUpdateData, and Player at the host pointer
// width. The entry Player is cached for mode validation and quest-stage
// storage; the name/protection/report paths deliberately reload ud.Player.
func playerAttribReadNative41A590(cf *cryptfile.CryptFile, unit *server.Object, info *server.PlayerInfo, h playerAttribReadHooks41A590) error {
	if cf == nil || !cf.ReadOnly() || info == nil {
		return fmt.Errorf("missing read-only crypt file or player info")
	}
	var update *server.PlayerUpdateData
	var cachedPlayer *server.Player
	if unit != nil && unit.UpdateData != nil {
		update = (*server.PlayerUpdateData)(unit.UpdateData)
		cachedPlayer = update.Player
	}

	rawVersion, err := cf.ReadU16()
	if err != nil {
		return err
	}
	version := int16(rawVersion)
	if version > 5 {
		return fmt.Errorf("unsupported attribute version %d", version)
	}
	if version >= 5 {
		mode, err := cf.ReadU32()
		if err != nil {
			return err
		}
		if cachedPlayer != nil && cachedPlayer.PlayerInd != byte(server.HostPlayerIndex) {
			quest := h.questMode()
			if (quest && mode != 4) || (!quest && mode == 4) {
				h.disconnect(cachedPlayer.PlayerInd)
				return fmt.Errorf("player mode %d does not match quest=%t", mode, quest)
			}
		}
	}

	nameLen, err := cf.ReadU8()
	if err != nil {
		return err
	}
	if nameLen >= 25 {
		return fmt.Errorf("player name has %d UTF-16 code units", nameLen)
	}
	raw := unsafe.Slice((*byte)(info.C()), playerInfoPayloadSize41A590)
	nameBytes := int(nameLen) * 2
	if err := playerAttribReadExact41A590(cf, raw[:nameBytes]); err != nil {
		return err
	}
	binary.LittleEndian.PutUint16(raw[nameBytes:], 0)
	if update != nil {
		player := update.Player
		if player == nil {
			return fmt.Errorf("player update has no live Player link")
		}
		player.SetName(info.Name())
		h.protectName(player.ProtPlayerOrigName, raw[:nameBytes])
	}

	if err := playerAttribReadExact41A590(cf, raw[50:54]); err != nil {
		return err
	}
	if err := playerAttribReadExact41A590(cf, raw[54:58]); err != nil {
		return err
	}
	if update != nil {
		player := update.Player
		if player == nil {
			return fmt.Errorf("player update lost its Player link")
		}
		h.protectInt(player.ProtPlayerField2239, info.Field2239())
		h.protectInt(player.ProtPlayerField2235, info.Field2235())
	}
	if err := playerAttribReadExact41A590(cf, raw[58:62]); err != nil {
		return err
	}
	if err := playerAttribReadExact41A590(cf, raw[62:66]); err != nil {
		return err
	}
	if err := playerAttribReadExact41A590(cf, raw[66:67]); err != nil {
		return err
	}
	if update != nil {
		player := update.Player
		if player == nil {
			return fmt.Errorf("player update lost its Player link")
		}
		h.protectClass(player.ProtPlayerClass, raw[66])
	}
	if err := playerAttribReadExact41A590(cf, raw[67:68]); err != nil {
		return err
	}
	for off := 68; off < 83; off += 3 {
		if err := playerAttribReadExact41A590(cf, raw[off:off+3]); err != nil {
			return err
		}
	}
	if version >= 2 {
		for off := 83; off < 88; off++ {
			if err := playerAttribReadExact41A590(cf, raw[off:off+1]); err != nil {
				return err
			}
		}
	}
	if unit != nil {
		h.initColors(unit.NetCode)
	}
	if err := playerAttribReadExact41A590(cf, raw[88:89]); err != nil {
		return err
	}
	if version >= 3 {
		extraLives, err := cf.ReadU32()
		if err != nil {
			return err
		}
		if update != nil {
			update.ExtraLives = extraLives
		}
		if update != nil && int32(update.ExtraLives) > h.maxExtraLives() {
			return fmt.Errorf("extra lives %d exceed maximum", update.ExtraLives)
		}
		if rawVersion == 3 {
			for range 9 {
				if _, err := cf.ReadU32(); err != nil {
					return err
				}
			}
		}
	}
	h.resetQuest(unit)
	if version >= 4 {
		stage, err := cf.ReadU32()
		if err != nil {
			return err
		}
		if cachedPlayer != nil {
			cachedPlayer.QuestStage = stage
		}
		if update != nil {
			player := update.Player
			if player == nil {
				return fmt.Errorf("player update lost its Player link")
			}
			h.sendQuestStage(player)
		}
	}
	return nil
}

func playerAttribReadRuntime41A590(cf *cryptfile.CryptFile, unit *server.Object, info *server.PlayerInfo) error {
	return playerAttribReadNative41A590(cf, unit, info, playerAttribReadHooks41A590{
		questMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeQuest)
		},
		disconnect: func(ind byte) {
			Nox_xxx_playerCallDisconnect_4DEAB0(ntype.PlayerInd(ind), 1)
		},
		protectName: func(token uint32, name []byte) {
			var ptr *C.int
			if len(name) != 0 {
				ptr = (*C.int)(unsafe.Pointer(&name[0]))
			}
			crc := C.nox_xxx_protectionStringCRCLen_56FAE0(
				ptr, C.uint(len(name)),
			)
			C.nox_xxx_playerResetProtectionCRC_56F7D0(C.int(token), crc)
		},
		protectInt: func(token, value uint32) {
			C.sub_56F780(C.int(token), C.int(value))
		},
		protectClass: func(token uint32, class byte) {
			C.sub_56F820(C.int(token), C.uchar(class))
		},
		initColors: func(netCode uint32) {
			if player := GetServer().S().Players.ByID(int(netCode)); player != nil {
				Nox_xxx_playerInitColors_461460(player)
			}
		},
		maxExtraLives: func() int32 {
			value := GetServer().S().Balance.Float("MaxExtraLives")
			return int32(C.nox_float2int(C.float(value)))
		},
		resetQuest: Sub_4D6000,
		sendQuestStage: func(player *server.Player) {
			C.sub_4D7450(C.int(player.PlayerInd), C.short(player.QuestStage))
		},
	})
}

// playerStatusWriteNative41AA30 preserves version 2 of the server status
// section without indexing a native Object as a PE32 uint32 array.
func playerStatusWriteNative41AA30(cf *cryptfile.CryptFile, unit *server.Object, info *server.PlayerInfo) error {
	if cf == nil || unit == nil || info == nil || unit.UpdateData == nil {
		return fmt.Errorf("missing status save state")
	}
	if err := cf.WriteU16(2); err != nil {
		return err
	}
	coop := noxflags.HasGame(noxflags.GameModeCoop)
	if err := cf.WriteU8(byte(bool2int(coop))); err != nil || !coop {
		return err
	}
	ud := (*server.PlayerUpdateData)(unit.UpdateData)
	if unit.HealthData == nil {
		return fmt.Errorf("player has no health data")
	}
	for _, value := range []uint16{
		Nox_xxx_unitGetMaxHP_4EE7A0(unit),
		Nox_xxx_playerGetMaxMana_4EECB0(unit),
		unit.HealthData.Cur,
		ud.ManaCur,
	} {
		if err := cf.WriteU16(value); err != nil {
			return err
		}
	}
	if err := cf.WriteU8(unit.Poison540); err != nil {
		return err
	}
	if err := cf.WriteU8(unit.Field541); err != nil {
		return err
	}
	if err := cf.WriteU16(unit.Field542); err != nil {
		return err
	}
	if err := cf.WriteU32(math.Float32bits(unit.Experience)); err != nil {
		return err
	}
	return cf.WriteU16(uint16(unit.Direction1))
}

type playerStatusReadHooks41AA30 struct {
	playerExists      func(uint32) bool
	coopMode          func() bool
	setMaxHP          func(*server.Object, uint16)
	setHP             func(*server.Object, uint16)
	setMaxMana        func(*server.Object, uint16)
	refreshMana       func(*server.Object)
	storeCurrentHP    func(uint16)
	storeCurrentMana  func(uint16)
	setPoison         func(*server.Object, byte)
	protectExperience func(uint32, float32)
	reportExperience  func(*server.Object)
}

// playerStatusReadNative41AA30 retains the version-2 stream widths and the
// entry-cached PlayerUpdateData while keeping Object, HealthData, and Player
// links at native width. Current HP and mana remain deferred in the two
// loader globals until the outer 0041A2E0 completion path restores them.
func playerStatusReadNative41AA30(cf *cryptfile.CryptFile, unit *server.Object, h playerStatusReadHooks41AA30) error {
	if cf == nil || !cf.ReadOnly() || unit == nil || unit.UpdateData == nil {
		return fmt.Errorf("missing status load state")
	}
	update := (*server.PlayerUpdateData)(unit.UpdateData)
	if !h.playerExists(unit.NetCode) {
		return fmt.Errorf("player %d is not registered", unit.NetCode)
	}
	rawVersion, err := cf.ReadU16()
	if err != nil {
		return err
	}
	version := int16(rawVersion)
	if version > 2 {
		return fmt.Errorf("unsupported status version %d", version)
	}
	present, err := cf.ReadU8()
	if err != nil {
		return err
	}
	if present == 0 {
		return nil
	}
	if !h.coopMode() {
		return fmt.Errorf("cooperative status payload outside cooperative mode")
	}
	maximumHP, err := cf.ReadU16()
	if err != nil {
		return err
	}
	h.setMaxHP(unit, maximumHP)
	h.setHP(unit, maximumHP)

	maximumMana, err := cf.ReadU16()
	if err != nil {
		return err
	}
	h.setMaxMana(unit, maximumMana)
	h.refreshMana(unit)
	if unit.HealthData == nil {
		return fmt.Errorf("player has no live HealthData")
	}
	currentHP, err := cf.ReadU16()
	if err != nil {
		return err
	}
	h.storeCurrentHP(currentHP)
	currentMana, err := cf.ReadU16()
	if err != nil {
		return err
	}
	h.storeCurrentMana(currentMana)

	poison, err := cf.ReadU8()
	if err != nil {
		return err
	}
	h.setPoison(unit, poison)
	if err := playerAttribReadExact41A590(cf, unsafe.Slice(&unit.Field541, 1)); err != nil {
		return err
	}
	field542 := unsafe.Slice((*byte)(unsafe.Pointer(&unit.Field542)), 2)
	if err := playerAttribReadExact41A590(cf, field542); err != nil {
		return err
	}
	experience := unsafe.Slice((*byte)(unsafe.Pointer(&unit.Experience)), 4)
	if err := playerAttribReadExact41A590(cf, experience); err != nil {
		return err
	}
	player := update.Player
	if player == nil {
		return fmt.Errorf("player update has no live Player link")
	}
	h.protectExperience(player.ProtUnitExperience, unit.Experience)
	h.reportExperience(unit)
	if version >= 2 {
		direction, err := cf.ReadU16()
		if err != nil {
			return err
		}
		unit.Direction1 = server.Dir16(direction)
		unit.Direction2 = server.Dir16(direction)
	}
	return nil
}

func playerStatusReadRuntime41AA30(cf *cryptfile.CryptFile, unit *server.Object) error {
	return playerStatusReadNative41AA30(cf, unit, playerStatusReadHooks41AA30{
		playerExists: func(netCode uint32) bool {
			return GetServer().S().Players.ByID(int(netCode)) != nil
		},
		coopMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeCoop)
		},
		setMaxHP: func(unit *server.Object, maximum uint16) {
			Nox_xxx_unitSetMaxHP_4EE7C0(unit, maximum)
		},
		setHP: Nox_xxx_unitSetHP_4E4560,
		setMaxMana: func(unit *server.Object, maximum uint16) {
			Nox_xxx_playerSetMaxMana_4EECD0(unit, maximum)
		},
		refreshMana: func(unit *server.Object) {
			Nox_xxx_playerManaRefresh_4EECF0(unit)
		},
		storeCurrentHP: func(current uint16) {
			*memmap.PtrUint32(0x5D4594, 527696) = uint32(current)
		},
		storeCurrentMana: func(current uint16) {
			*memmap.PtrUint32(0x5D4594, 527700) = uint32(current)
		},
		setPoison: func(unit *server.Object, poison byte) {
			setSomePoisonDataCall4EEA90(unit, int32(poison))
		},
		protectExperience: func(token uint32, experience float32) {
			C.sub_56F8C0(C.int(token), C.float(experience))
		},
		reportExperience: func(unit *server.Object) {
			GetServer().S().NetReportExperience(unit)
		},
	})
}

func playerInventoryGridCount41AC30() int {
	return int(C.sub_41B3B0())
}

func playerInventoryGridCodes41AC30(column, row int) []uint32 {
	cnt := int(C.sub_467810(C.int(column), C.int(row)))
	if cnt <= 0 {
		return nil
	}
	ptr := C.sub_467870(C.int(column), C.int(row))
	if ptr == nil {
		return nil
	}
	vals := unsafe.Slice((*C.int)(unsafe.Pointer(ptr)), cnt)
	out := make([]uint32, cnt)
	for i, value := range vals {
		out[i] = uint32(value)
	}
	return out
}

func playerInventorySelectedCodes41AC30() (secondary, quiver uint32) {
	return uint32(C.sub_4678B0()), uint32(C.sub_4678C0())
}

func playerInventoryNotifyLoaded41AC30(ind byte) {
	C.nox_xxx_netMsgInventoryLoaded_4D96E0(C.int(ind))
}

func playerInventoryQuestSaveable41AC30(item *server.Object, glyph uint16) bool {
	return !item.Class().Has(object.ClassKey) && item.TypeInd != glyph
}

func playerInventoryFindNetCode41AC30(unit *server.Object, code uint32) *server.Object {
	for item := unit.InvFirstItem; item != nil; item = item.InvNextItem {
		if item.NetCode == code {
			return item
		}
	}
	return nil
}

func playerInventoryWriteItem41AC30(cf *cryptfile.CryptFile, item *server.Object) error {
	typ := GetServer().S().Types.ByInd(int(item.TypeInd))
	if typ == nil {
		return fmt.Errorf("inventory object has unknown type %d", item.TypeInd)
	}
	name := typ.ID()
	if len(name) > 0xff {
		return fmt.Errorf("inventory type name is too long: %q", name)
	}
	if err := cf.WriteU8(byte(len(name))); err != nil {
		return err
	}
	if _, err := cf.Write([]byte(name)); err != nil {
		return err
	}
	if err := item.CallXfer(nil); err != nil {
		return fmt.Errorf("inventory %s: %w", name, err)
	}
	return nil
}

// playerInventoryWriteNative41AC30 follows the version 3 write branch of
// GAME.EXE. The GUI-grid branch remains significant in cooperative play: it
// defines item serialization order when every server item has a client cell.
func playerInventoryWriteNative41AC30(cf *cryptfile.CryptFile, unit *server.Object, info *server.PlayerInfo) error {
	if cf == nil || unit == nil || info == nil || unit.UpdateData == nil {
		return fmt.Errorf("missing inventory save state")
	}
	ud := (*server.PlayerUpdateData)(unit.UpdateData)
	pl := ud.Player
	if pl == nil {
		return fmt.Errorf("player update has no Player link")
	}
	if err := cf.WriteU16(3); err != nil {
		return err
	}
	present := !(noxflags.HasGame(noxflags.GameOnline) && !noxflags.HasGame(noxflags.GameModeQuest))
	if err := cf.WriteU8(byte(bool2int(present))); err != nil || !present {
		if err == nil {
			playerInventoryNotifyLoaded41AC30(pl.PlayerInd)
		}
		return err
	}
	if !noxflags.HasGame(noxflags.GameModeCoop) && !noxflags.HasGame(noxflags.GameModeQuest) {
		return fmt.Errorf("inventory section requires cooperative or quest mode")
	}
	if err := cf.WriteU32(pl.GoldVal); err != nil {
		return err
	}

	quest := noxflags.HasGame(noxflags.GameModeQuest)
	glyph := uint16(GetServer().S().Types.IndByID("Glyph"))
	count := 0
	for item := unit.InvFirstItem; item != nil; item = item.InvNextItem {
		if !quest || playerInventoryQuestSaveable41AC30(item, glyph) {
			count++
		}
	}
	if err := cf.WriteU32(uint32(count)); err != nil {
		return err
	}

	gridCount := 0
	if noxflags.HasGame(noxflags.GameModeCoop) {
		gridCount = playerInventoryGridCount41AC30()
	}
	if count == gridCount && noxflags.HasGame(noxflags.GameModeCoop) {
		for row := 0; row < 20; row++ {
			for column := 0; column < 4; column++ {
				for _, code := range playerInventoryGridCodes41AC30(column, row) {
					item := playerInventoryFindNetCode41AC30(unit, code)
					if item == nil {
						return fmt.Errorf("inventory grid refers to missing net code %d", code)
					}
					if err := playerInventoryWriteItem41AC30(cf, item); err != nil {
						return err
					}
				}
			}
		}
	} else {
		for item := unit.InvFirstItem; item != nil; item = item.InvNextItem {
			if quest && !playerInventoryQuestSaveable41AC30(item, glyph) {
				continue
			}
			if err := playerInventoryWriteItem41AC30(cf, item); err != nil {
				return err
			}
		}
	}

	equipped := byte(0)
	for item := unit.InvFirstItem; item != nil; item = item.InvNextItem {
		if item.Flags().Has(object.FlagEquipped) {
			equipped++
		}
	}
	if err := cf.WriteU8(equipped); err != nil {
		return err
	}
	for item := unit.InvFirstItem; item != nil; item = item.InvNextItem {
		if item.Flags().Has(object.FlagEquipped) {
			if err := cf.WriteU32(uint32(item.ScriptIDVal)); err != nil {
				return err
			}
		}
	}

	secondary, quiver := playerInventorySelectedCodes41AC30()
	for _, code := range []uint32{secondary, quiver} {
		var scriptID uint32
		if item := playerInventoryFindNetCode41AC30(unit, code); item != nil && code != 0 {
			scriptID = uint32(item.ScriptIDVal)
		}
		if err := cf.WriteU32(scriptID); err != nil {
			return err
		}
	}
	if err := cf.WriteU8(byte(ud.CurTraps)); err != nil {
		return err
	}
	if quest && GetServer().S().QuestInventoryLimits4F2C30(unit) == 0 {
		return fmt.Errorf("quest inventory exceeds GAME.EXE item limits")
	}
	playerInventoryNotifyLoaded41AC30(pl.PlayerInd)
	return nil
}

type playerInventoryReadHooks41AC30 struct {
	coopMode             func() bool
	questMode            func() bool
	syncLevel            func(*server.Object)
	protectGold          func(uint32, int32)
	delayedDelete        func(*server.Object)
	newObject            func(string) *server.Object
	transferItem         func(*server.Object) error
	questItemAllowed     func(*server.Object) bool
	placeWorld           func(item, owner *server.Object)
	addPending           func()
	placeInventory       func(owner, item *server.Object) bool
	tryDequip            func(owner, item *server.Object) bool
	tryEquip             func(owner, item *server.Object) bool
	clearClientSelection func()
	reportSecondary      func(byte, *server.Object)
	reportQuiver         func(byte, *server.Object)
	nextScriptID         func() uint32
	questLimits          func(*server.Object) bool
	notifyLoaded         func(byte)
}

func playerInventorySubGoldNative41AC30(unit *server.Object, amount uint32, protect func(uint32, int32)) error {
	if unit == nil || unit.UpdateData == nil {
		return fmt.Errorf("player has no live UpdateData for gold subtraction")
	}
	update := (*server.PlayerUpdateData)(unit.UpdateData)
	player := update.Player
	if player == nil {
		return fmt.Errorf("player update has no live Player link for gold subtraction")
	}
	if player.GoldVal >= amount {
		player.GoldVal -= amount
	} else {
		player.GoldVal = 0
	}
	player = update.Player
	if player == nil {
		return fmt.Errorf("player link disappeared during gold subtraction")
	}
	protect(player.ProtPlayerGold, int32(uint32(0)-amount))
	return nil
}

func playerInventoryAddGoldNative41AC30(unit *server.Object, amount uint32, protect func(uint32, int32)) error {
	if unit == nil || unit.UpdateData == nil {
		return fmt.Errorf("player has no live UpdateData for gold addition")
	}
	update := (*server.PlayerUpdateData)(unit.UpdateData)
	player := update.Player
	if player == nil {
		return fmt.Errorf("player update has no live Player link for gold addition")
	}
	player.GoldVal += amount
	player = update.Player
	if player == nil {
		return fmt.Errorf("player link disappeared during gold addition")
	}
	protect(player.ProtPlayerGold, int32(amount))
	return nil
}

func playerInventoryFindScriptID41AC30(unit *server.Object, id uint32, each func(*server.Object) bool) {
	for item := unit.InvFirstItem; item != nil; item = item.InvNextItem {
		if uint32(item.ScriptIDVal) == id && !each(item) {
			return
		}
	}
}

// playerInventoryReadNative41AC30 follows the GAME.EXE version-3 read
// branch. Stream integers keep their PE32 widths, while the cached
// PlayerUpdateData and every inventory Object link retain the host pointer
// width. The entry-cached update record is intentionally reused for traps and
// the final notification; gold helpers reload unit.UpdateData as the original
// 004FA5D0/004FA590 callees do.
func playerInventoryReadNative41AC30(cf *cryptfile.CryptFile, unit *server.Object, h playerInventoryReadHooks41AC30) error {
	if cf == nil || !cf.ReadOnly() || unit == nil || unit.UpdateData == nil {
		return fmt.Errorf("missing inventory load state")
	}
	entryUpdate := (*server.PlayerUpdateData)(unit.UpdateData)
	h.syncLevel(unit)

	rawVersion, err := cf.ReadU16()
	if err != nil {
		return err
	}
	version := int16(rawVersion)
	if version > 3 {
		return fmt.Errorf("unsupported inventory version %d", version)
	}
	present, err := cf.ReadU8()
	if err != nil {
		return err
	}
	if present != 0 {
		coop, quest := h.coopMode(), h.questMode()
		if !coop && !quest {
			return fmt.Errorf("inventory section requires cooperative or quest mode")
		}

		gold, err := cf.ReadU32()
		if err != nil {
			return err
		}
		if entryUpdate.Player == nil {
			return fmt.Errorf("player update has no live Player link")
		}
		if err := playerInventorySubGoldNative41AC30(unit, entryUpdate.Player.GoldVal, h.protectGold); err != nil {
			return err
		}
		if err := playerInventoryAddGoldNative41AC30(unit, gold, h.protectGold); err != nil {
			return err
		}

		for item := unit.InvFirstItem; item != nil; {
			next := item.InvNextItem
			h.delayedDelete(item)
			item = next
		}
		rawCount, err := cf.ReadU32()
		if err != nil {
			return err
		}
		count := int32(rawCount)
		if quest && count > 2560 {
			return fmt.Errorf("quest inventory count %d exceeds 2560", count)
		}
		for index := int32(0); index < count; index++ {
			nameLen, err := cf.ReadU8()
			if err != nil {
				return fmt.Errorf("inventory[%d] name length: %w", index, err)
			}
			name := make([]byte, int(nameLen))
			if err := playerAttribReadExact41A590(cf, name); err != nil {
				return fmt.Errorf("inventory[%d] name: %w", index, err)
			}
			item := h.newObject(string(name))
			if item == nil {
				return fmt.Errorf("inventory[%d] has unknown type %q", index, name)
			}
			if err := h.transferItem(item); err != nil {
				return fmt.Errorf("inventory[%d] %q: %w", index, name, err)
			}
			item.PosVec = types.Pointf{X: 2944, Y: 2944}
			if quest && !h.questItemAllowed(item) {
				return fmt.Errorf("inventory[%d] %q is not valid in quest mode", index, name)
			}
			h.placeWorld(item, unit)
			h.addPending()
			if !h.placeInventory(unit, item) {
				if !quest {
					return fmt.Errorf("inventory[%d] %q could not be placed", index, name)
				}
				h.delayedDelete(item)
			}
			if !item.Flags().Has(object.FlagPending) && item.Flags().Has(object.FlagEquipped) {
				h.tryDequip(unit, item)
			}
		}

		equippedCount, err := cf.ReadU8()
		if err != nil {
			return err
		}
		for index := 0; index < int(equippedCount); index++ {
			scriptID, err := cf.ReadU32()
			if err != nil {
				return err
			}
			playerInventoryFindScriptID41AC30(unit, scriptID, func(item *server.Object) bool {
				h.tryEquip(unit, item)
				return true
			})
		}
		if coop {
			h.clearClientSelection()
		}

		secondaryID, err := cf.ReadU32()
		if err != nil {
			return err
		}
		if secondaryID != 0 {
			playerInventoryFindScriptID41AC30(unit, secondaryID, func(item *server.Object) bool {
				player := entryUpdate.Player
				if player != nil {
					h.reportSecondary(player.PlayerInd, item)
				}
				return false
			})
		}
		if version >= 2 {
			quiverID, err := cf.ReadU32()
			if err != nil {
				return err
			}
			if quiverID != 0 {
				playerInventoryFindScriptID41AC30(unit, quiverID, func(item *server.Object) bool {
					player := entryUpdate.Player
					if player != nil {
						h.reportQuiver(player.PlayerInd, item)
					}
					return false
				})
			}
		}
		if quest {
			for item := unit.InvFirstItem; item != nil; item = item.InvNextItem {
				item.ScriptIDVal = int32(h.nextScriptID())
				item.Extent = item.NetCode
			}
		}

		traps := byte(0)
		if version >= 3 {
			traps, err = cf.ReadU8()
			if err != nil {
				return err
			}
		}
		if quest {
			traps = 0
		}
		entryUpdate.CurTraps = entryUpdate.CurTraps&^uint32(0xff) | uint32(traps)
	}

	if h.questMode() && !h.questLimits(unit) {
		return fmt.Errorf("quest inventory exceeds GAME.EXE item limits")
	}
	player := entryUpdate.Player
	if player == nil {
		return fmt.Errorf("player update has no live Player link for inventory notification")
	}
	h.notifyLoaded(player.PlayerInd)
	return nil
}

func playerInventoryReadRuntime41AC30(cf *cryptfile.CryptFile, unit *server.Object) error {
	outer := GetServer()
	srv := outer.S()
	return playerInventoryReadNative41AC30(cf, unit, playerInventoryReadHooks41AC30{
		coopMode:  func() bool { return noxflags.HasGame(noxflags.GameModeCoop) },
		questMode: func() bool { return noxflags.HasGame(noxflags.GameModeQuest) },
		syncLevel: func(unit *server.Object) {
			_ = playerSyncLevelCall4EF140(unit)
		},
		protectGold:   Nox_xxx_protectGoldDelta_56F920,
		delayedDelete: outer.DelayedDelete,
		newObject:     srv.NewObjectByTypeID,
		transferItem: func(item *server.Object) error {
			return item.CallXfer(nil)
		},
		questItemAllowed: func(item *server.Object) bool {
			return srv.QuestItemEligible4F2590(item) != 0
		},
		placeWorld: func(item, owner *server.Object) {
			outer.CreateObjectAt(item, owner, types.Pointf{X: 2944, Y: 2944})
		},
		addPending: outer.ObjectsAddPending,
		placeInventory: func(owner, item *server.Object) bool {
			return Nox_xxx_inventoryServPlace_4F36F0(owner, item, 1, 1)
		},
		tryDequip: Nox_xxx_playerTryDequip_4F2FB0,
		tryEquip:  Nox_xxx_playerTryEquip_4F2F70,
		clearClientSelection: func() {
			C.sub_467750(0, 0)
			C.sub_467740(0)
		},
		reportSecondary: func(ind byte, item *server.Object) {
			C.nox_xxx_netSendSecondaryWeapon_4D9670(C.int(ind), (*C.uint)(item.CObj()), 0)
		},
		reportQuiver: func(ind byte, item *server.Object) {
			C.nox_xxx_netMsgLastQuiver_4D96B0(C.int(ind), (*C.uint)(item.CObj()))
		},
		nextScriptID: func() uint32 {
			return uint32(srv.Objs.NextObjectScriptID())
		},
		questLimits: func(unit *server.Object) bool {
			return srv.QuestInventoryLimits4F2C30(unit) != 0
		},
		notifyLoaded: playerInventoryNotifyLoaded41AC30,
	})
}

func playerSavePresentExceptQuest41B420() bool {
	return !(noxflags.HasGame(noxflags.GameOnline) && !noxflags.HasGame(noxflags.GameModeQuest))
}

func playerSaveWriteName41B420(cf *cryptfile.CryptFile, name string) error {
	if len(name) > 0xff {
		return fmt.Errorf("save entry name is too long: %q", name)
	}
	if err := cf.WriteU8(byte(len(name))); err != nil {
		return err
	}
	_, err := cf.Write([]byte(name))
	return err
}

func playerGuideName41B420(ind int) string {
	return GoString(C.nox_xxx_guideNameByN_427230(C.int(ind)))
}

func playerFieldbookWriteNative41B420(cf *cryptfile.CryptFile, unit *server.Object, info *server.PlayerInfo) error {
	if cf == nil || unit == nil || info == nil || unit.UpdateData == nil {
		return fmt.Errorf("missing field-guide save state")
	}
	pl := (*server.PlayerUpdateData)(unit.UpdateData).Player
	if pl == nil {
		return fmt.Errorf("player update has no Player link")
	}
	if err := cf.WriteU16(1); err != nil {
		return err
	}
	present := playerSavePresentExceptQuest41B420()
	if err := cf.WriteU8(byte(bool2int(present))); err != nil || !present {
		return err
	}
	if !noxflags.HasGame(noxflags.GameModeCoop) && !noxflags.HasGame(noxflags.GameModeQuest) {
		return fmt.Errorf("field-guide section requires cooperative or quest mode")
	}
	count := byte(0)
	for ind := 1; ind < len(pl.BeastScrollLvl); ind++ {
		if pl.BeastScrollLvl[ind] != 0 {
			count++
		}
	}
	if err := cf.WriteU8(count); err != nil {
		return err
	}
	for ind := 1; ind < len(pl.BeastScrollLvl); ind++ {
		if pl.BeastScrollLvl[ind] == 0 {
			continue
		}
		if err := playerSaveWriteName41B420(cf, playerGuideName41B420(ind)); err != nil {
			return err
		}
	}
	return nil
}

type playerFieldbookReadHooks41B420 struct {
	playerExists      func(uint32) bool
	coopMode          func() bool
	questMode         func() bool
	guideByName       func(string) int
	questGuideAllowed func(int) bool
	awardGuide        func(*server.Object, int)
}

// playerFieldbookReadNative41B420 preserves GAME.EXE's version-1 byte stream
// while keeping Object, PlayerUpdateData, and Player links at the host pointer
// width. Unknown guide names retain the original index-zero award attempt;
// that attempt is ignored by the award routine outside Quest validation.
func playerFieldbookReadNative41B420(cf *cryptfile.CryptFile, unit *server.Object, h playerFieldbookReadHooks41B420) error {
	if cf == nil || !cf.ReadOnly() || unit == nil {
		return fmt.Errorf("missing field-guide load state")
	}
	if !h.playerExists(unit.NetCode) {
		return fmt.Errorf("player %d is not registered", unit.NetCode)
	}
	rawVersion, err := cf.ReadU16()
	if err != nil {
		return err
	}
	version := int16(rawVersion)
	if version > 1 {
		return fmt.Errorf("unsupported field-guide version %d", version)
	}
	present, err := cf.ReadU8()
	if err != nil {
		return err
	}
	if present == 0 {
		return nil
	}
	coop, quest := h.coopMode(), h.questMode()
	if !coop && !quest {
		return fmt.Errorf("field-guide section requires cooperative or quest mode")
	}
	count, err := cf.ReadU8()
	if err != nil {
		return err
	}
	if count > 41 {
		return fmt.Errorf("field-guide count %d exceeds 41", count)
	}
	for index := 0; index < int(count); index++ {
		nameLen, err := cf.ReadU8()
		if err != nil {
			return fmt.Errorf("field-guide[%d] name length: %w", index, err)
		}
		name := make([]byte, int(nameLen))
		if err := playerAttribReadExact41A590(cf, name); err != nil {
			return fmt.Errorf("field-guide[%d] name: %w", index, err)
		}
		guide := h.guideByName(string(name))
		if quest && !h.questGuideAllowed(guide) {
			return fmt.Errorf("field-guide[%d] %q is not valid in quest mode", index, name)
		}
		h.awardGuide(unit, guide)
	}
	return nil
}

func playerGuideByName41B420(name string) int {
	for guide := 0; guide < 41; guide++ {
		if playerGuideName41B420(guide) == name {
			return guide
		}
	}
	return 0
}

func playerQuestGuideAllowedNative4F2EF0(guide int) bool {
	const (
		blob       = uintptr(0x587000)
		entryTable = uintptr(207796)
		groupTable = uintptr(207032)
	)
	allowed := false
	for index := uintptr(0); index < 256; index++ {
		off := entryTable + 12*index
		entryGuide := int(memmap.Uint32(blob, off))
		if entryGuide == 0 {
			break
		}
		if entryGuide == guide && memmap.Uint32(blob, off+4) != 0 {
			allowed = true
		}
	}
	for groupIndex := uintptr(0); groupIndex < 41; groupIndex++ {
		group := *memmap.PtrPtr(blob, groupTable+4*groupIndex)
		if group == nil {
			break
		}
		groupBlob, groupOff := memmap.BlobByPtr(group)
		if groupBlob == nil {
			continue
		}
		if memmap.Uint32(groupBlob.Addr, groupOff) == 0 {
			continue
		}
		for memberIndex := uintptr(1); memberIndex < 41; memberIndex++ {
			member := int(memmap.Uint32(groupBlob.Addr, groupOff+4*memberIndex))
			if member == 0 {
				break
			}
			if member == guide {
				allowed = true
				break
			}
		}
	}
	return allowed
}

func playerGuideRelationsNative4FAE80(guide int) []int {
	const (
		blob       = uintptr(0x587000)
		groupTable = uintptr(216292)
	)
	var related []int
	for groupIndex := uintptr(0); groupIndex < 41; groupIndex++ {
		group := *memmap.PtrPtr(blob, groupTable+4*groupIndex)
		if group == nil {
			break
		}
		groupBlob, groupOff := memmap.BlobByPtr(group)
		if groupBlob == nil || int(memmap.Uint32(groupBlob.Addr, groupOff)) != guide {
			continue
		}
		for memberIndex := uintptr(1); memberIndex < 41; memberIndex++ {
			member := int(memmap.Uint32(groupBlob.Addr, groupOff+4*memberIndex))
			if member == 0 {
				break
			}
			related = append(related, member)
		}
	}
	return related
}

type playerFieldbookAwardHooks41B420 struct {
	awardProtection func(uint32, int, int)
	relatedGuides   func(int) []int
	reportAward     func(*server.Object, *server.Player, int)
}

// playerFieldbookAwardLoadNative41B420 is the a3 == 0 path of GAME.EXE
// 004FAE80 used by save loading. Audio and reward broadcasts belong only to
// the live-award path and therefore cannot occur here.
func playerFieldbookAwardLoadNative41B420(unit *server.Object, guide int, h playerFieldbookAwardHooks41B420) bool {
	if unit == nil || !unit.Class().Has(object.ClassPlayer) || guide <= 0 || guide >= 41 || unit.UpdateData == nil {
		return false
	}
	update := (*server.PlayerUpdateData)(unit.UpdateData)
	player := update.Player
	if player == nil || player.BeastScrollLvl[guide] != 0 {
		return false
	}
	player.BeastScrollLvl[guide] = 1
	h.awardProtection(player.Prot4640, guide, int(player.BeastScrollLvl[guide]))
	for _, related := range h.relatedGuides(guide) {
		if related <= 0 || related >= len(player.BeastScrollLvl) {
			continue
		}
		player.BeastScrollLvl[related] = 1
		h.awardProtection(player.Prot4640, related, int(player.BeastScrollLvl[related]))
	}
	h.reportAward(unit, player, guide)
	return true
}

func playerFieldbookReadRuntime41B420(cf *cryptfile.CryptFile, unit *server.Object) error {
	srv := GetServer().S()
	return playerFieldbookReadNative41B420(cf, unit, playerFieldbookReadHooks41B420{
		playerExists: func(netCode uint32) bool {
			return srv.Players.ByID(int(netCode)) != nil
		},
		coopMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeCoop)
		},
		questMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeQuest)
		},
		guideByName:       playerGuideByName41B420,
		questGuideAllowed: playerQuestGuideAllowedNative4F2EF0,
		awardGuide: func(unit *server.Object, guide int) {
			playerFieldbookAwardLoadNative41B420(unit, guide, playerFieldbookAwardHooks41B420{
				awardProtection: Nox_xxx_playerAwardSpellProtectionCRC_56FCE0,
				relatedGuides:   playerGuideRelationsNative4FAE80,
				reportAward: func(unit *server.Object, player *server.Player, guide int) {
					packet := [3]byte{byte(netmsg.MSG_REPORT_GUIDE_AWARD), byte(guide), 0}
					srv.NetSendPacketXxx1(int(player.PlayerInd), packet[:], nil, 1)
				},
			})
		},
	})
}

func playerSpellName41B660(ind int) string {
	return GoString(C.nox_xxx_spellNameByN_424870(C.int(ind)))
}

func playerAbilityName41B660(ind int) string {
	return GoString(C.nox_xxx_abilityGetName_425250(C.int(ind)))
}

func playerSpellbookWriteNative41B660(cf *cryptfile.CryptFile, unit *server.Object, info *server.PlayerInfo) error {
	if cf == nil || unit == nil || info == nil || unit.UpdateData == nil {
		return fmt.Errorf("missing spellbook save state")
	}
	pl := (*server.PlayerUpdateData)(unit.UpdateData).Player
	if pl == nil {
		return fmt.Errorf("player update has no Player link")
	}
	if err := cf.WriteU16(3); err != nil {
		return err
	}
	present := playerSavePresentExceptQuest41B420()
	if err := cf.WriteU8(byte(bool2int(present))); err != nil || !present {
		return err
	}
	if !noxflags.HasGame(noxflags.GameModeCoop) && !noxflags.HasGame(noxflags.GameModeQuest) {
		return fmt.Errorf("spellbook section requires cooperative or quest mode")
	}
	limit := 6
	spellbook := info.PlayerClass() != 0
	if spellbook {
		limit = len(pl.SpellLvl)
	}
	count := byte(0)
	for ind := 1; ind < limit; ind++ {
		if pl.SpellLvl[ind] != 0 {
			count++
		}
	}
	if err := cf.WriteU8(count); err != nil {
		return err
	}
	for ind := 1; ind < limit; ind++ {
		level := pl.SpellLvl[ind]
		if level == 0 {
			continue
		}
		name := playerAbilityName41B660(ind)
		if spellbook {
			name = playerSpellName41B660(ind)
		}
		if err := playerSaveWriteName41B420(cf, name); err != nil {
			return err
		}
		if err := cf.WriteU32(level); err != nil {
			return err
		}
	}
	return nil
}

type playerSpellbookReadHooks41B660 struct {
	playerByID        func(uint32) *server.Player
	coopMode          func() bool
	questMode         func() bool
	spellByName       func(string) int
	abilityByName     func(string) int
	questSpellAllowed func(int) bool
	grantSpell        func(*server.Object, int, int32)
	awardAbility      func(*server.Object, int)
}

// playerSpellbookReadNative41B660 keeps the version-3 stream scalar widths
// while resolving the entry-cached Player and the unit passed to award paths
// at native pointer width. PlayerClass remains a live field of that cached
// Player, matching each original branch read rather than freezing the class.
func playerSpellbookReadNative41B660(cf *cryptfile.CryptFile, unit *server.Object, h playerSpellbookReadHooks41B660) error {
	if cf == nil || !cf.ReadOnly() || unit == nil {
		return fmt.Errorf("missing spellbook load state")
	}
	player := h.playerByID(unit.NetCode)
	if player == nil {
		return fmt.Errorf("player %d is not registered", unit.NetCode)
	}
	rawVersion, err := cf.ReadU16()
	if err != nil {
		return err
	}
	version := int16(rawVersion)
	if version > 3 {
		return fmt.Errorf("unsupported spellbook version %d", version)
	}
	present, err := cf.ReadU8()
	if err != nil {
		return err
	}
	if present == 0 {
		return nil
	}
	coop, quest := h.coopMode(), h.questMode()
	if !coop && !quest {
		return fmt.Errorf("spellbook section requires cooperative or quest mode")
	}
	count, err := cf.ReadU8()
	if err != nil {
		return err
	}
	if count > 137 {
		return fmt.Errorf("spellbook count %d exceeds 137", count)
	}
	for index := 0; index < int(count); index++ {
		nameLen, err := cf.ReadU8()
		if err != nil {
			return fmt.Errorf("spellbook[%d] name length: %w", index, err)
		}
		name := make([]byte, int(nameLen))
		if err := playerAttribReadExact41A590(cf, name); err != nil {
			return fmt.Errorf("spellbook[%d] name: %w", index, err)
		}
		level := int32(3)
		if version >= 2 {
			rawLevel, err := cf.ReadU32()
			if err != nil {
				return fmt.Errorf("spellbook[%d] level: %w", index, err)
			}
			level = int32(rawLevel)
		}
		if quest && level > 3 && byte(player.PlayerClass()) != 0 {
			return fmt.Errorf("spellbook[%d] %q level %d exceeds Quest level 3", index, name, level)
		}
		if quest && byte(player.PlayerClass()) != 0 {
			spellID := h.spellByName(string(name))
			if !h.questSpellAllowed(spellID) {
				return fmt.Errorf("spellbook[%d] %q is not valid in quest mode", index, name)
			}
		}
		if version < 3 || byte(player.PlayerClass()) != 0 {
			h.grantSpell(unit, h.spellByName(string(name)), level)
		} else {
			h.awardAbility(unit, h.abilityByName(string(name)))
		}
	}
	return nil
}

func playerSpellByName41B660(name string) int {
	return int(spell.ParseID(name))
}

func playerAbilityByName41B660(name string) int {
	for ability := server.AbilityInvalid; ability < server.AbilityMax; ability++ {
		if ability.String() == name {
			return int(ability)
		}
	}
	return 0
}

func playerQuestSpellAllowedNative4F2E70(spellID int) bool {
	return server.QuestSpellAllowed4F2E70(int32(spellID)) != 0
}

func playerSpellQuestSingleLevel41B660(spellID int) bool {
	switch spellID {
	case 19, 34, 45, 46, 47, 48, 49, 117, 118, 119, 120, 121, 122, 123, 124, 125, 134:
		return true
	default:
		return false
	}
}

type playerSpellGrantLoadHooks41B660 struct {
	coopOrQuest     func() bool
	questMode       func() bool
	hasFlags        func(int, things.SpellFlags) bool
	validSpell      func(int) bool
	awardProtection func(uint32, int, int)
	reportAward     func(*server.Object, *server.Player, int)
}

// playerSpellGrantLoadNative41B660 is the no-audio, no-shop path selected by
// 0041B660 when it calls GAME.EXE 004FB550 with a3/a4 == 0. The saved level is
// signed for comparisons but is stored with its original 32-bit bit pattern.
func playerSpellGrantLoadNative41B660(unit *server.Object, spellID int, level int32, h playerSpellGrantLoadHooks41B660) bool {
	if unit == nil || !unit.Class().Has(object.ClassPlayer) || spellID <= 0 || spellID >= 137 || unit.UpdateData == nil {
		return false
	}
	update := (*server.PlayerUpdateData)(unit.UpdateData)
	player := update.Player
	if player == nil {
		return false
	}
	current := player.SpellLvl[spellID]
	if (h.coopOrQuest() && current == 3) || current == 5 {
		return false
	}
	quest := h.questMode()
	if quest && playerSpellQuestSingleLevel41B660(spellID) && current != 0 {
		return false
	}
	current++
	if current > 5 {
		current = 5
	}
	if quest && current > 3 {
		current = 3
	}
	if level != 0 {
		current = uint32(level)
	}
	player.SpellLvl[spellID] = current
	h.awardProtection(player.Prot4636, spellID, int(current))

	var family things.SpellFlags
	switch {
	case h.hasFlags(spellID, things.SpellFlags(0x1000)):
		family = things.SpellFlags(0x2000)
	case h.hasFlags(spellID, things.SpellFlags(0x4000)):
		family = things.SpellFlags(0x8000)
	case h.hasFlags(spellID, things.SpellFlags(0x10000)):
		family = things.SpellFlags(0x20000)
	}
	if family != 0 {
		for related := 1; related < 137; related++ {
			if !h.hasFlags(related, family) || !h.validSpell(related) {
				continue
			}
			relatedLevel := player.SpellLvl[related]
			if level != 0 {
				relatedLevel = uint32(level)
			} else {
				relatedLevel++
			}
			if relatedLevel > 5 {
				relatedLevel = 5
			}
			player.SpellLvl[related] = relatedLevel
			// GAME.EXE checks and protects the selected spell index here,
			// even while storing a related spell's level.
			if quest && player.SpellLvl[spellID] > 3 {
				player.SpellLvl[spellID] = 3
			}
			h.awardProtection(player.Prot4636, spellID, int(relatedLevel))
		}
	}
	player = update.Player
	if player == nil {
		return false
	}
	h.reportAward(unit, player, spellID)
	return true
}

func playerSpellbookReadRuntime41B660(cf *cryptfile.CryptFile, unit *server.Object) error {
	srv := GetServer().S()
	return playerSpellbookReadNative41B660(cf, unit, playerSpellbookReadHooks41B660{
		playerByID: func(netCode uint32) *server.Player {
			return srv.Players.ByID(int(netCode))
		},
		coopMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeCoop)
		},
		questMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeQuest)
		},
		spellByName:       playerSpellByName41B660,
		abilityByName:     playerAbilityByName41B660,
		questSpellAllowed: playerQuestSpellAllowedNative4F2E70,
		grantSpell: func(unit *server.Object, spellID int, level int32) {
			playerSpellGrantLoadNative41B660(unit, spellID, level, playerSpellGrantLoadHooks41B660{
				coopOrQuest: func() bool {
					return noxflags.HasGame(noxflags.GameModeCoop | noxflags.GameModeQuest)
				},
				questMode: func() bool {
					return noxflags.HasGame(noxflags.GameModeQuest)
				},
				hasFlags: func(spellID int, flags things.SpellFlags) bool {
					return srv.Spells.HasFlags(spell.ID(spellID), flags)
				},
				validSpell: func(spellID int) bool {
					def := srv.Spells.DefByInd(spell.ID(spellID))
					return def != nil && def.IsValid()
				},
				awardProtection: Nox_xxx_playerAwardSpellProtectionCRC_56FCE0,
				reportAward: func(unit *server.Object, player *server.Player, spellID int) {
					packet := [4]byte{
						byte(netmsg.MSG_REPORT_SPELL_AWARD), byte(spellID),
						byte(player.SpellLvl[spellID]), 0,
					}
					srv.NetSendPacketXxx1(int(player.PlayerInd), packet[:], nil, 1)
				},
			})
		},
		awardAbility: func(unit *server.Object, ability int) {
			C.nox_xxx_abilityRewardServ_4FB9C0_ability(asObjectC(unit), C.int(ability), 0)
		},
	})
}

type playerEnchantmentReadHooks41B9C0 struct {
	coopMode            func() bool
	clearPlayerSpells   func()
	clearDurationSpells func()
	parseEnchant        func(string) (server.EnchantID, bool)
	applyEnchant        func(*server.Object, server.EnchantID, byte)
	gameFPS             func() uint32
	setShieldHealth     func(*server.Object, uint32)
	stopBerserker       func()
	executeAbility      func(*server.Object, server.Ability)
	setAbilityDuration  func(*server.Object, server.Ability, int)
	setCooldown         func(*server.Object, server.Ability, int)
	reportCooldown      func(*server.Object, server.Ability)
}

func playerEnchantmentReadAbilityState41B9C0(cf *cryptfile.CryptFile, unit *server.Object, h playerEnchantmentReadHooks41B9C0) error {
	activeBerserker, err := cf.ReadU8()
	if err != nil {
		return fmt.Errorf("Berserker state: %w", err)
	}
	if activeBerserker == 1 {
		h.stopBerserker()
	}
	activeHarpoon, err := cf.ReadU8()
	if err != nil {
		return fmt.Errorf("Harpoon state: %w", err)
	}
	rawDuration, err := cf.ReadU32()
	if err != nil {
		return fmt.Errorf("Harpoon duration: %w", err)
	}
	if activeHarpoon == 1 {
		h.executeAbility(unit, server.Ability(4))
		h.setAbilityDuration(unit, server.Ability(4), int(int32(rawDuration)))
	}
	for ind := playerAbilityCooldownStart41B9C0(activeBerserker == 1); ind < 6; ind++ {
		rawCooldown, err := cf.ReadU32()
		if err != nil {
			return fmt.Errorf("ability %d cooldown: %w", ind, err)
		}
		ability := server.Ability(ind)
		cooldown := int(int32(rawCooldown))
		h.setCooldown(unit, ability, cooldown)
		if cooldown != 0 {
			h.reportCooldown(unit, ability)
		}
	}
	return nil
}

// playerEnchantmentReadNative41B9C0 preserves GAME.EXE's version-5 scalar
// stream while resolving Object.UpdateData and Player at the host pointer
// width. The update-data pointer is cached at entry; its live Player link is
// read only at the original Warrior ability gates.
func playerEnchantmentReadNative41B9C0(cf *cryptfile.CryptFile, unit *server.Object, h playerEnchantmentReadHooks41B9C0) error {
	if cf == nil || !cf.ReadOnly() || unit == nil {
		return fmt.Errorf("missing enchantment load state")
	}
	var update *server.PlayerUpdateData
	if unit.UpdateData != nil {
		update = (*server.PlayerUpdateData)(unit.UpdateData)
	}
	rawVersion, err := cf.ReadU16()
	if err != nil {
		return err
	}
	version := int16(rawVersion)
	if version > 5 {
		return fmt.Errorf("unsupported enchantment version %d", version)
	}
	if h.coopMode() {
		h.clearPlayerSpells()
		h.clearDurationSpells()
	}
	present, err := cf.ReadU8()
	if err != nil {
		return err
	}
	if present != 0 {
		if !h.coopMode() {
			return fmt.Errorf("enchantment section requires cooperative mode")
		}
		count, err := cf.ReadU8()
		if err != nil {
			return err
		}
		for index := 0; index < int(count); index++ {
			nameLen, err := cf.ReadU8()
			if err != nil {
				return fmt.Errorf("enchantment[%d] name length: %w", index, err)
			}
			name := make([]byte, int(nameLen))
			if err := playerAttribReadExact41A590(cf, name); err != nil {
				return fmt.Errorf("enchantment[%d] name: %w", index, err)
			}
			enchant, ok := h.parseEnchant(string(name))
			if !ok || enchant >= 32 {
				return fmt.Errorf("enchantment[%d] %q is unknown", index, name)
			}
			duration, err := cf.ReadU16()
			if err != nil {
				return fmt.Errorf("enchantment[%d] duration: %w", index, err)
			}
			power := byte(2)
			if version >= 2 {
				power, err = cf.ReadU8()
				if err != nil {
					return fmt.Errorf("enchantment[%d] power: %w", index, err)
				}
			}
			h.applyEnchant(unit, enchant, power)
			if duration == 0 {
				duration = uint16(h.gameFPS())
				_ = h.gameFPS() // GAME.EXE performs the second, otherwise-unused call.
			}
			unit.BuffsDur[enchant] = duration
			if enchant == server.ENCHANT_SHIELD && version >= 3 {
				health, err := cf.ReadU32()
				if err != nil {
					return fmt.Errorf("enchantment[%d] shield health: %w", index, err)
				}
				h.setShieldHealth(unit, health)
			}
		}
		if version >= 5 {
			if update == nil || update.Player == nil {
				return fmt.Errorf("player update has no live Player link")
			}
			if update.Player.PlayerClass() == 0 {
				if err := playerEnchantmentReadAbilityState41B9C0(cf, unit, h); err != nil {
					return err
				}
			}
		}
	}
	if rawVersion == 4 {
		if update == nil || update.Player == nil {
			return fmt.Errorf("player update has no live Player link")
		}
		if update.Player.PlayerClass() == 0 {
			return playerEnchantmentReadAbilityState41B9C0(cf, unit, h)
		}
	}
	return nil
}

func playerEnchantmentReadRuntime41B9C0(cf *cryptfile.CryptFile, unit *server.Object) error {
	srv := GetServer().S()
	return playerEnchantmentReadNative41B9C0(cf, unit, playerEnchantmentReadHooks41B9C0{
		coopMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeCoop)
		},
		clearPlayerSpells: Nox_xxx_spellCastByPlayer_4FEEF0,
		clearDurationSpells: func() {
			srv.Spells.Dur.Sub4FE8A0(0)
		},
		parseEnchant: server.ParseEnchant,
		applyEnchant: func(unit *server.Object, enchant server.EnchantID, power byte) {
			arg := server.SpellAcceptArg{Obj: unit, Pos: unit.Pos()}
			GetServer().Nox_xxx_spellAccept4FD400(enchant.Spell(), unit, unit, unit, &arg, int(power))
		},
		gameFPS: gameFPS,
		setShieldHealth: func(unit *server.Object, health uint32) {
			for it := srv.Spells.Dur.List; it != nil; it = it.Next {
				if it.Flags88&1 == 0 && it.Spell == 51 && it.Target48 == unit {
					it.Field72 = int32(health)
					break
				}
			}
		},
		stopBerserker: func() {
			Sub_4FC670(1)
		},
		executeAbility: func(unit *server.Object, ability server.Ability) {
			Nox_xxx_playerExecuteAbil_4FBB70(unit, int(ability))
		},
		setAbilityDuration: func(unit *server.Object, ability server.Ability, duration int) {
			srv.Abils.Sub4FC070(unit, ability, duration)
		},
		setCooldown: func(unit *server.Object, ability server.Ability, cooldown int) {
			srv.Abils.SetCooldownForUnit(unit, ability, cooldown)
		},
		reportCooldown: func(unit *server.Object, ability server.Ability) {
			Nox_xxx_netAbilRepotState_4D8100(unit, ability, 0)
		},
	})
}

func playerEnchantIDs41B9C0() []server.EnchantID {
	var out []server.EnchantID
	for ind := int(C.sub_424D00()); ind != -1; ind = int(C.sub_424D20(C.int(ind))) {
		if ind >= 0 && ind < 32 {
			out = append(out, server.EnchantID(ind))
		}
	}
	return out
}

func playerEnchantName41B9C0(ind server.EnchantID) string {
	return GoString(C.nox_xxx_getEnchantName_4248F0(C.int(ind)))
}

func playerShieldHealth41B9C0(unit *server.Object) uint32 {
	for it := unit.Server().Spells.Dur.List; it != nil; it = it.Next {
		if it.Flags88&1 == 0 && it.Spell == 51 && it.Target48 == unit {
			return uint32(it.Field72)
		}
	}
	return 100
}

func playerAbilityCooldownStart41B9C0(activeBerserker bool) int {
	if activeBerserker {
		return 2
	}
	return 1
}

func playerEnchantmentWriteNative41B9C0(cf *cryptfile.CryptFile, unit *server.Object, info *server.PlayerInfo) error {
	if cf == nil || unit == nil || info == nil || unit.UpdateData == nil {
		return fmt.Errorf("missing enchantment save state")
	}
	if err := cf.WriteU16(5); err != nil {
		return err
	}
	present := !noxflags.HasGame(noxflags.GameOnline)
	if err := cf.WriteU8(byte(bool2int(present))); err != nil || !present {
		return err
	}
	if !noxflags.HasGame(noxflags.GameModeCoop) {
		return fmt.Errorf("enchantment section requires cooperative mode")
	}
	ids := playerEnchantIDs41B9C0()
	count := byte(0)
	for _, ind := range ids {
		if unit.HasEnchant(ind) {
			count++
		}
	}
	if err := cf.WriteU8(count); err != nil {
		return err
	}
	for _, ind := range ids {
		if !unit.HasEnchant(ind) {
			continue
		}
		if err := playerSaveWriteName41B420(cf, playerEnchantName41B9C0(ind)); err != nil {
			return err
		}
		if err := cf.WriteU16(unit.BuffsDur[ind]); err != nil {
			return err
		}
		if err := cf.WriteU8(unit.BuffsPower[ind]); err != nil {
			return err
		}
		if ind == server.ENCHANT_SHIELD {
			if err := cf.WriteU32(playerShieldHealth41B9C0(unit)); err != nil {
				return err
			}
		}
	}

	pl := (*server.PlayerUpdateData)(unit.UpdateData).Player
	if pl == nil {
		return fmt.Errorf("player update has no Player link")
	}
	if pl.PlayerClass() == 0 {
		abil := &GetServer().S().Abils
		active1 := abil.IsActive(unit, server.Ability(1))
		if err := cf.WriteU8(byte(bool2int(active1))); err != nil {
			return err
		}
		if err := cf.WriteU8(byte(bool2int(abil.IsActive(unit, server.Ability(4))))); err != nil {
			return err
		}
		if err := cf.WriteU32(uint32(abil.Sub4FC030(unit, server.Ability(4)))); err != nil {
			return err
		}
		first := playerAbilityCooldownStart41B9C0(active1)
		for ind := first; ind < 6; ind++ {
			cooldown := abil.GetCooldownForUnit(unit, server.Ability(ind))
			if err := cf.WriteU32(uint32(cooldown)); err != nil {
				return err
			}
		}
	}
	return nil
}

type playerJournalReadHooks41BEC0 struct {
	coopMode      func() bool
	removeEntries func(*server.Object, uint16)
	addEntry      func(*server.Object, string, uint16)
}

// playerJournalReadNative41BEC0 keeps the version-1 journal stream fixed-width
// while caching Object.UpdateData.Player at native pointer width at entry. Save
// order is oldest-first; prepending each restored entry rebuilds the original
// newest-first Player.Journal list.
func playerJournalReadNative41BEC0(cf *cryptfile.CryptFile, unit *server.Object, h playerJournalReadHooks41BEC0) error {
	if cf == nil || !cf.ReadOnly() || unit == nil || unit.UpdateData == nil {
		return fmt.Errorf("missing journal load state")
	}
	update := (*server.PlayerUpdateData)(unit.UpdateData)
	player := update.Player
	if player == nil {
		return fmt.Errorf("player update has no Player link")
	}
	rawVersion, err := cf.ReadU16()
	if err != nil {
		return err
	}
	if version := int16(rawVersion); version > 1 {
		return fmt.Errorf("unsupported journal version %d", version)
	}
	present, err := cf.ReadU8()
	if err != nil {
		return err
	}
	if present == 0 {
		return nil
	}
	if !h.coopMode() {
		return fmt.Errorf("journal section requires cooperative mode")
	}
	// GAME.EXE traverses the cached list to initialize the writer count before
	// the read overwrites its low word. Keep the traversal and native links.
	existingCount := uint32(0)
	for entry := player.Journal; entry != nil; entry = entry.Next {
		existingCount++
	}
	_ = existingCount
	count, err := cf.ReadU16()
	if err != nil {
		return err
	}
	h.removeEntries(unit, 0xffff)
	for index := 0; index < int(count); index++ {
		nameLen, err := cf.ReadU8()
		if err != nil {
			return fmt.Errorf("journal[%d] name length: %w", index, err)
		}
		if nameLen >= 64 {
			return fmt.Errorf("journal[%d] name length %d exceeds 63", index, nameLen)
		}
		name := make([]byte, int(nameLen))
		if err := playerAttribReadExact41A590(cf, name); err != nil {
			return fmt.Errorf("journal[%d] name: %w", index, err)
		}
		entryType, err := cf.ReadU16()
		if err != nil {
			return fmt.Errorf("journal[%d] type: %w", index, err)
		}
		h.addEntry(unit, string(name), entryType)
	}
	return nil
}

func playerJournalReadRuntime41BEC0(cf *cryptfile.CryptFile, unit *server.Object) error {
	srv := GetServer().S()
	return playerJournalReadNative41BEC0(cf, unit, playerJournalReadHooks41BEC0{
		coopMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeCoop)
		},
		removeEntries: func(unit *server.Object, mask uint16) {
			C.sub_4277B0(asObjectC(unit), C.ushort(mask))
		},
		addEntry: func(unit *server.Object, name string, entryType uint16) {
			srv.JournalEntryAdd427500(unit, name, entryType)
		},
	})
}

func playerJournalWriteNative41BEC0(cf *cryptfile.CryptFile, unit *server.Object, info *server.PlayerInfo) error {
	if cf == nil || unit == nil || info == nil || unit.UpdateData == nil {
		return fmt.Errorf("missing journal save state")
	}
	pl := (*server.PlayerUpdateData)(unit.UpdateData).Player
	if pl == nil {
		return fmt.Errorf("player update has no Player link")
	}
	if err := cf.WriteU16(1); err != nil {
		return err
	}
	present := !noxflags.HasGame(noxflags.GameOnline)
	if err := cf.WriteU8(byte(bool2int(present))); err != nil || !present {
		return err
	}
	if !noxflags.HasGame(noxflags.GameModeCoop) {
		return fmt.Errorf("journal section requires cooperative mode")
	}
	count := uint16(0)
	var oldest *server.PlayerJournal
	for it := pl.Journal; it != nil; it = it.Next {
		count++
		oldest = it
	}
	if err := cf.WriteU16(count); err != nil {
		return err
	}
	for it := oldest; it != nil; it = it.Prev {
		name := it.EntryBuf[:]
		if ind := bytes.IndexByte(name, 0); ind >= 0 {
			name = name[:ind]
		}
		if len(name) > 0xff {
			return fmt.Errorf("journal entry is too long")
		}
		if err := cf.WriteU8(byte(len(name))); err != nil {
			return err
		}
		if _, err := cf.Write(name); err != nil {
			return err
		}
		if err := cf.WriteU16(it.Field3); err != nil {
			return err
		}
	}
	return nil
}

func playerMapNamePayload41C080(pl *server.Player, mapName string) ([]byte, error) {
	if pl == nil {
		return nil, fmt.Errorf("missing player")
	}
	if len(mapName)*2 > len(pl.SaveNameBuf) {
		return nil, fmt.Errorf("map name is too long for GAME.EXE save buffer: %q", mapName)
	}
	copy(pl.SaveNameBuf[:], mapName)
	pl.SaveNameBuf[len(mapName)] = 0
	return pl.SaveNameBuf[:2*len(mapName)], nil
}

type playerGameReadHooks41C080 struct {
	onlineMode       func() bool
	readQuestJournal func(*cryptfile.CryptFile) error
	transferWalls    func(*cryptfile.CryptFile, *server.Object) bool
	setStage         func(byte)
}

// playerGameReadNative41C080 preserves the version-5 game-state stream while
// caching Object.UpdateData.Player at native pointer width. The map-name field
// deliberately remains GAME.EXE's 32-byte buffer and its unusual 2*length
// payload followed by a terminator at the undoubled length.
func playerGameReadNative41C080(cf *cryptfile.CryptFile, unit *server.Object, h playerGameReadHooks41C080) error {
	if cf == nil || !cf.ReadOnly() || unit == nil || unit.UpdateData == nil {
		return fmt.Errorf("missing game-state load state")
	}
	update := (*server.PlayerUpdateData)(unit.UpdateData)
	player := update.Player
	if player == nil {
		return fmt.Errorf("player update has no Player link")
	}
	if h.onlineMode() {
		return nil
	}
	rawVersion, err := cf.ReadU16()
	if err != nil {
		return err
	}
	version := int16(rawVersion)
	if version > 5 {
		return fmt.Errorf("unsupported game-state version %d", version)
	}
	if version >= 5 {
		if _, err := cf.ReadU32(); err != nil {
			return fmt.Errorf("last object script ID: %w", err)
		}
	}
	nameLen, err := cf.ReadU16()
	if err != nil {
		return fmt.Errorf("map name length: %w", err)
	}
	payloadLen := 2 * int(nameLen)
	if payloadLen > len(player.SaveNameBuf) {
		return fmt.Errorf("map name length %d exceeds GAME.EXE save buffer", nameLen)
	}
	if err := playerAttribReadExact41A590(cf, player.SaveNameBuf[:payloadLen]); err != nil {
		return fmt.Errorf("map name payload: %w", err)
	}
	player.SaveNameBuf[int(nameLen)] = 0
	if version >= 2 {
		if err := h.readQuestJournal(cf); err != nil {
			return fmt.Errorf("quest journal transfer: %w", err)
		}
	}
	if version >= 3 && !h.transferWalls(cf, unit) {
		return fmt.Errorf("wall transfer failed")
	}
	if version >= 4 {
		stage, err := cf.ReadU8()
		if err != nil {
			return fmt.Errorf("stage: %w", err)
		}
		h.setStage(stage)
	} else {
		h.setStage(0)
	}
	return nil
}

func playerGameReadRuntime41C080(cf *cryptfile.CryptFile, unit *server.Object) error {
	return playerGameReadNative41C080(cf, unit, playerGameReadHooks41C080{
		onlineMode: func() bool {
			return noxflags.HasGame(noxflags.GameOnline)
		},
		readQuestJournal: questJournalReadNative500B70,
		transferWalls: func(_ *cryptfile.CryptFile, unit *server.Object) bool {
			return Sub_5000B0(unit) != 0
		},
		setStage: func(stage byte) {
			*memmap.PtrUint8(0x5D4594, 831252) = stage
		},
	})
}

func playerGameWriteNative41C080(cf *cryptfile.CryptFile, unit *server.Object, info *server.PlayerInfo) error {
	if cf == nil || unit == nil || info == nil || unit.UpdateData == nil {
		return fmt.Errorf("missing game save state")
	}
	if noxflags.HasGame(noxflags.GameOnline) {
		return nil
	}
	pl := (*server.PlayerUpdateData)(unit.UpdateData).Player
	if pl == nil {
		return fmt.Errorf("player update has no Player link")
	}
	if err := cf.WriteU16(5); err != nil {
		return err
	}
	lastID := uint32(GetServer().S().Objs.LastObjectScriptID())
	if err := cf.WriteU32(lastID); err != nil {
		return err
	}
	GetServer().S().Objs.SetLastObjectScriptID(server.ObjectScriptID(lastID))

	mapName := Nox_xxx_mapGetMapName_409B40()
	mapPayload, err := playerMapNamePayload41C080(pl, mapName)
	if err != nil {
		return err
	}
	if err := cf.WriteU16(uint16(len(mapName))); err != nil {
		return err
	}
	if _, err := cf.Write(mapPayload); err != nil {
		return err
	}
	if err := questJournalWriteNative500A60(cf); err != nil {
		return fmt.Errorf("quest journal transfer: %w", err)
	}
	if Sub_5000B0(unit) == 0 {
		return fmt.Errorf("wall transfer failed")
	}
	stage := memmap.Uint8(0x5D4594, 831252)
	if err := cf.WriteU8(stage); err != nil {
		return err
	}
	*memmap.PtrUint8(0x5D4594, 831252) = stage
	return nil
}
