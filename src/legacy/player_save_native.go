package legacy

/*
#include "GAME1_1.h"
#include "GAME2_1.h"
#include "GAME3_2.h"
#include "GAME5_2.h"
#include "common__magic__speltree.h"
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"

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

var playerInventoryQuestStackNames41AC30 = [...]string{
	"RedPotion",
	"BluePotion",
	"CurePoisonPotion",
	"HastePotion",
	"InvisibilityPotion",
	"ShieldPotion",
	"VampirismPotion",
	"FireProtectPotion",
	"ShockProtectPotion",
	"PoisonProtectPotion",
	"InvulnerabilityPotion",
	"InfravisionPotion",
}

func playerInventoryQuestLimits41AC30(isPlayer bool, count func(string) int32, staffLimit int32) bool {
	if !isPlayer {
		return true
	}
	for _, name := range playerInventoryQuestStackNames41AC30 {
		if count(name) > 9 {
			return false
		}
	}
	return count("InfinitePainWand") <= staffLimit
}

func playerInventoryQuestLimitsNative41AC30(unit *server.Object) bool {
	if unit == nil {
		return true
	}
	s := GetServer().S()
	return playerInventoryQuestLimits41AC30(
		unit.Class().Has(object.ClassPlayer),
		func(name string) int32 {
			return unit.CountInventoryWithType(int32(s.Types.IndByID(name)))
		},
		int32(s.Balance.Float("ForceOfNatureStaffLimit")),
	)
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
	if quest && !playerInventoryQuestLimitsNative41AC30(unit) {
		return fmt.Errorf("quest inventory exceeds GAME.EXE item limits")
	}
	playerInventoryNotifyLoaded41AC30(pl.PlayerInd)
	return nil
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
	if info.PlayerClass() == 0 {
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
