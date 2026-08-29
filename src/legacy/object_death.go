package legacy

/*
#include "server__object__die__die.h"
#include "GAME4_3.h"
#include "GAME5.h"

void nox_xxx_diePlayer_54D2B0_go(nox_object_t* unit);
void nox_xxx_dieCreateObject_54E010_go(nox_object_t* source);
void nox_xxx_dieSpawnObject_54E070_go(nox_object_t* source);

static int nox_call_objectType_parseDeath_go(int (*fnc)(char*, void*), char* arg1, void* arg2) { return fnc(arg1, arg2); }
*/
import "C"
import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"strings"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

var playerDieAnkhType54D2B0 uint32

func init() {
	server.RegisterObjectDeath("PlayerDie", C.nox_xxx_diePlayer_54D2B0_go, 0)
	server.RegisterObjectDeath("PotionDie", C.nox_xxx_diePotion_54CBB0, 0)
	server.RegisterObjectDeath("ImpEggDie", C.nox_xxx_dieImpEgg_54CAE0, 0)
	server.RegisterObjectDeath("GlyphDie", C.nox_xxx_dieGlyph_54DF30, 0)
	server.RegisterObjectDeath("BarrelDie", C.nox_xxx_dieBarrel_54DFA0, 0)
	server.RegisterObjectDeathGo("CreateObjectDie", C.nox_xxx_dieCreateObject_54E010_go, func(source *server.Object) {
		createObjectDieCall54E010(source)
	}, unsafe.Sizeof(server.CreateSpawnObjectDeathData54E010{}))
	server.RegisterObjectDeathGo("SpawnObjectDie", C.nox_xxx_dieSpawnObject_54E070_go, func(source *server.Object) {
		spawnObjectDieCall54E070(source)
	}, unsafe.Sizeof(server.CreateSpawnObjectDeathData54E010{}))
	server.RegisterObjectDeath("PolypDie", C.nox_xxx_diePolyp_54CB10, 0)
	server.RegisterObjectDeath("MarkerDie", C.nox_xxx_dieMarker_54E460, 0)
	server.RegisterObjectDeath("WeaponDie", C.nox_xxx_dieWeapon_54E370_obj_die, 0)
	server.RegisterObjectDeath("ArmorDie", C.nox_xxx_dieArmor_54E170_obj_die, 0)
	server.RegisterObjectDeath("BoulderDie", C.nox_xxx_dieBoulder_54E4B0, 0)
	server.RegisterObjectDeath("GameBallDie", C.nox_xxx_dieGameBall_54E620, 0)
	server.RegisterObjectDeath("MonsterGeneratorDie", C.nox_xxx_dieMonsterGen_54E630, 0)

	server.RegisterObjectDeathParse("CreateObjectDie", wrapObjectDeathParseC(C.sub_536B40))
	server.RegisterObjectDeathParse("SpawnObjectDie", wrapObjectDeathParseC(C.sub_536B40))
}

func createSpawnObjectDeathRuntime54E010() server.CreateSpawnObjectDeathRuntime54E010 {
	outer := GetServer()
	s := outer.S()
	return server.CreateSpawnObjectDeathRuntime54E010{
		NewObjectByTypeID: s.NewObjectByTypeID,
		CreateAt: func(obj *server.Object, pos types.Pointf) {
			outer.CreateObjectAt(obj, nil, pos)
		},
		Audio: func(id uint32, obj *server.Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		DelayedDelete: outer.DelayedDelete,
	}
}

var createObjectDieCall54E010 = func(source *server.Object) {
	server.CreateObjectDieNative54E010(source, createSpawnObjectDeathRuntime54E010())
}

var spawnObjectDieCall54E070 = func(source *server.Object) {
	server.SpawnObjectDieNative54E070(source, createSpawnObjectDeathRuntime54E010())
}

func createObjectDieExportCall54E010(source *server.Object) {
	C.nox_xxx_dieCreateObject_54E010_go(asObjectC(source))
}

func spawnObjectDieExportCall54E070(source *server.Object) {
	C.nox_xxx_dieSpawnObject_54E070_go(asObjectC(source))
}

//export nox_xxx_dieCreateObject_54E010_go
func nox_xxx_dieCreateObject_54E010_go(source *nox_object_t) {
	createObjectDieCall54E010(asObjectS(source))
}

//export nox_xxx_dieSpawnObject_54E070_go
func nox_xxx_dieSpawnObject_54E070_go(source *nox_object_t) {
	spawnObjectDieCall54E070(asObjectS(source))
}

//export nox_xxx_diePlayer_54D2B0_go
func nox_xxx_diePlayer_54D2B0_go(unitp *nox_object_t) {
	unit := asObjectS(unitp)
	s := GetServer().S()
	handled := server.PlayerDieNative54D2B0(unit, server.PlayerDieRuntime54D2B0{
		GameFlag: func(flag uint32) bool {
			return noxflags.HasGame(noxflags.GameFlag(flag))
		},
		PrepareAnkhType: func() {
			if playerDieAnkhType54D2B0 == 0 {
				playerDieAnkhType54D2B0 = uint32(s.Types.IndByID("AnkhTradable"))
			}
		},
		CancelPendingSave: func() {
			Sub_4DB170(false, nil, 0)
		},
		Audio: func(id int, obj *server.Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		SetPlayerState: Nox_xxx_playerSetState_4FA020,
		RemoveActionShadow: func(obj *server.Object) {
			Nox_xxx_action_4DA9F0(obj)
		},
		DropAllItems: dropAllItemsCall4EDA40,
		NotifyPlayerDied: func(obj *server.Object) {
			var packet [3]byte
			packet[0] = byte(netmsg.MSG_PLAYER_DIED)
			binary.LittleEndian.PutUint16(packet[1:], uint16(obj.NetCode))
			s.NetSendPacketXxx1(255, packet[:], nil, 0)
		},
		ProtectMana: func(token uint32, delta int16) {
			Nox_xxx_protectMana_56F9E0(int(token), delta)
		},
		SetBuffFlags: func(obj *server.Object, flags uint32) {
			obj.SetBuffFlags(flags, func(player *server.Player, flags uint32) {
				Nox_xxx_playerResetProtectionCRC_56F7D0(player.ProtUnitBuffs, int(flags))
			})
		},
		CancelAbilities: Nox_xxx_playerCancelAbils_4FC180,
		CancelSpells:    Nox_xxx_playerCancelSpells_4FEAE0,
		Unsupported: func(reason string, obj *server.Object) {
			if s.Log != nil {
				s.Log.Error("PlayerDie native branch is not ported",
					slog.String("reason", reason),
					slog.Uint64("unit_ptr", uint64(uintptr(obj.CObj()))),
				)
			}
		},
	})
	if handled {
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 4 {
		C.nox_xxx_diePlayer_54D2B0(C.int(uintptr(unsafe.Pointer(unitp))))
	}
}

func wrapObjectDeathParseC(ptr unsafe.Pointer) server.ObjectParseFunc {
	return func(objt *server.ObjectType, args []string) error {
		if Nox_call_objectType_parseDeath_go(ptr, strings.Join(args, " "), objt.DeathData) == 0 {
			return fmt.Errorf("cannot parse death data for %q", objt.ID())
		}
		return nil
	}
}

func Nox_call_objectType_parseDeath_go(a1 unsafe.Pointer, a2 string, a3 unsafe.Pointer) int {
	cstr := CString(a2)
	defer StrFree(cstr)
	return int(C.nox_call_objectType_parseDeath_go((*[0]byte)(a1), cstr, a3))
}
