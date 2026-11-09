package legacy

import (
	"encoding/binary"
	"runtime"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

// Only external service boundaries are replaceable. Object fields and wire
// widths below always use the actual native layouts and original DWORDs.
type worldPlayerRuntime40A830 struct {
	PlayerFirst       func() *server.Player
	PlayerNext        func(*server.Player) *server.Player
	UnitFirst         func() *server.Object
	UnitNext          func(*server.Object) *server.Object
	TeamFirst         func() *server.Team
	TeamNext          func(*server.Team) *server.Team
	TeamContains      func(*server.ObjectTeam, server.TeamID) bool
	EngineNoRendering func() bool
	GameplayTeams     func() bool
	GameFlag          func(uint32) uint32
	GameFlags         func() uint32
	TimerStatus       func() int32
	TimerMode         func(int16) byte
	TimerStart        func()
	TimerSet          func(int32) int32
	Send              func(int32, []byte) int32
	Lesson            func(*server.Object)
	Minimap           func(mark bool, index byte, object *server.Object, flags uint32)
	WallCreate        func(int32, int32) *server.Wall
	WallAttach        func(*server.Wall, *server.Object) unsafe.Pointer
	ArmorDefinition   func(uint16) *server.Modifier
	CompareNames      func(*uint16, *uint16) int32
}

type worldPlayerRuntimeFrame40A830 struct {
	value    *worldPlayerRuntime40A830
	previous *worldPlayerRuntimeFrame40A830
}

var worldPlayerRuntimeStack40A830 *worldPlayerRuntimeFrame40A830

// Serial scopes, including nested callbacks. These are service hooks, not a
// substitute for the entry algorithms or a PE/native physical-fault model.
func pushWorldPlayerRuntime40A830(value *worldPlayerRuntime40A830) func() {
	frame := &worldPlayerRuntimeFrame40A830{value, worldPlayerRuntimeStack40A830}
	worldPlayerRuntimeStack40A830 = frame
	return func() {
		if worldPlayerRuntimeStack40A830 != frame {
			panic("world/player runtime LIFO")
		}
		worldPlayerRuntimeStack40A830 = frame.previous
	}
}
func worldPlayerRuntimeCurrent40A830() *worldPlayerRuntime40A830 {
	if f := worldPlayerRuntimeStack40A830; f != nil {
		return f.value
	}
	return &worldPlayerDefaults40A830
}

func worldPlayerForUnit40A830(unit *server.Object) *server.Player {
	return (*server.PlayerUpdateData)(unit.UpdateData).Player
}
func countNonEliminatedNative40A830(team *server.Team) int32 {
	r := worldPlayerRuntimeCurrent40A830()
	var count int32
	for unit := r.UnitFirst(); unit != nil; unit = r.UnitNext(unit) {
		if r.TeamContains(&unit.TeamVal, team.IDVal) {
			player := worldPlayerForUnit40A830(unit)
			if player.Field3680&1 == 0 && (player.PlayerInd != 31 || !r.EngineNoRendering()) {
				count++
			}
		}
	}
	runtime.KeepAlive(team)
	return count
}
func gamePlayAnyNative40A8A0() int32 {
	r := worldPlayerRuntimeCurrent40A830()
	count := int32(0)
	if !r.GameplayTeams() {
		for p := r.PlayerFirst(); p != nil; p = r.PlayerNext(p) {
			if p.PlayerUnit != nil {
				flags := p.Field3680
				if flags&1 == 0 || flags&0x20 != 0 {
					count++
				}
			}
		}
	} else {
		for team := r.TeamFirst(); team != nil; team = r.TeamNext(team) {
			unit := r.UnitFirst()
			for unit != nil {
				if r.TeamContains(&unit.TeamVal, team.IDVal) {
					flags := worldPlayerForUnit40A830(unit).Field3680
					if flags&1 == 0 || flags&0x20 != 0 {
						break
					}
				}
				unit = r.UnitNext(unit)
			}
			if unit != nil {
				count++
			}
		}
	}
	if count > 1 {
		return 1
	}
	return 0
}
func cliCanTalkNative4100F0(text *uint16) int32 {
	defer runtime.KeepAlive(text)
	for p := unsafe.Pointer(text); ; p = unsafe.Add(p, 2) {
		value := *(*uint16)(p)
		if value == 0 {
			return 1
		}
		if value&0xff00 != 0 {
			return 0
		}
	}
}
func doorAttachWallNative410360(object *server.Object, x, y int32) unsafe.Pointer {
	r := worldPlayerRuntimeCurrent40A830()
	wall := r.WallCreate(x, y)
	result := r.WallAttach(wall, object)
	runtime.KeepAlive(object)
	return result
}
func secretWallBlockNative410760(node *server.SecretWall) *server.SecretWall {
	cell := secretWallNativeHeadCell410550()
	node.Next = (*server.SecretWall)(*(*unsafe.Pointer)(cell))
	// Pointer-typed store preserves Go write barriers; C callers that retain Go
	// allocations must individually pin them for the complete C-head lifetime.
	*(*unsafe.Pointer)(cell) = unsafe.Pointer(node)
	return node
}
func secretWallFirstNative410780() *server.SecretWall {
	return (*server.SecretWall)(*(*unsafe.Pointer)(secretWallNativeHeadCell410550()))
}
func secretWallNextNative410790(node *server.SecretWall) *server.SecretWall {
	if node == nil {
		return nil
	}
	return node.Next
}
func modifierNextNative4133C0(modifier *server.ModifierEff) *server.ModifierEff {
	return modifier.Next()
}
func checkPauseNative413A50() int32 {
	return int32(worldPlayerRuntimeCurrent40A830().GameFlag(0x40000))
}
func itemDefendNative415C00(item *server.Object) float64 {
	if item == nil || uint32(item.ObjClass)&0x02000000 == 0 {
		return 0
	}
	definition := worldPlayerRuntimeCurrent40A830().ArmorDefinition(item.TypeInd)
	if definition == nil {
		return 0
	}
	value := definition.DamageCoeffOrArmor64
	var effect *server.ModifierEff
	if item.InitData != nil {
		effect = (*server.ModifierInitData)(item.InitData).Modifiers[0]
	}
	if effect != nil && effect.Defend76.Fnc != nil {
		worldPlayerCallDefend415C00(effect.Defend76.Fnc, effect, item, &value)
	}
	runtime.KeepAlive(definition)
	runtime.KeepAlive(item)
	runtime.KeepAlive(effect)
	return float64(value)
}
func cliGamedataNative416590(index int32) unsafe.Pointer {
	// Match the existing signed record accessor. Overflow beyond C's defined
	// signed-index domain is an explicit int32-wrap extension, not old-C credit.
	off := int32(uint32(371380) + 58*uint32(index))
	return memmap.PtrOff(0x5D4594, uintptr(int64(off)))
}
func settingsAcquiredNative4169D0(value int32) int32 {
	*(*uint32)(memmap.PtrOff(0x5D4594, 371704)) = uint32(value)
	return value
}
func playerForceLessonsNative416E50(send int32) unsafe.Pointer {
	r := worldPlayerRuntimeCurrent40A830()
	for p := r.PlayerFirst(); p != nil; p = r.PlayerNext(p) {
		p.Lessons, p.Field2140 = 0, 0
		if send != 0 && p.PlayerUnit != nil {
			r.Lesson(p.PlayerUnit)
		}
	}
	return nil
}
func playerByNameNative4170D0(name *uint16) *server.Player {
	if name == nil {
		return nil
	}
	r := worldPlayerRuntimeCurrent40A830()
	defer runtime.KeepAlive(name)
	for p := r.PlayerFirst(); p != nil; p = r.PlayerNext(p) {
		if r.CompareNames(&p.NameFinal[0], name) == 0 {
			return p
		}
	}
	return nil
}
func minimapAllNative417470(object *server.Object, flags uint32, mark bool) unsafe.Pointer {
	r := worldPlayerRuntimeCurrent40A830()
	defer runtime.KeepAlive(object)
	for p := r.PlayerFirst(); p != nil; p = r.PlayerNext(p) {
		r.Minimap(mark, p.PlayerInd, object, flags)
	}
	return nil
}
func playerUnsetNative417530(player *server.Player, mask uint32) byte {
	r := worldPlayerRuntimeCurrent40A830()
	player.Field3680 &^= mask // before every service
	result := r.GameFlag(1)
	if result != 0 {
		if mask&1 != 0 {
			result = r.GameFlag(128)
			if result == 0 {
				result = uint32(gamePlayAnyNative40A8A0())
				if result != 0 {
					result = uint32(r.TimerStatus())
					if result == 0 {
						mode := r.TimerMode(int16(uint16(r.GameFlags())))
						result = result&0xffffff00 | uint32(mode)
						if mode != 0 {
							r.TimerStart()
							result = result&0xffffff00 | uint32(byte(r.TimerSet(1)))
						}
					}
				}
			}
		}
		if mask&0x423 != 0 {
			result = result&0xffffff00 | uint32(byte(netReportStatusNative417630(player)))
		}
	}
	runtime.KeepAlive(player)
	return byte(result) // public C char only
}
func sendAllStatusNative4175C0(recipient *server.Player) unsafe.Pointer {
	r := worldPlayerRuntimeCurrent40A830()
	defer runtime.KeepAlive(recipient)
	for p := r.PlayerFirst(); p != nil; p = r.PlayerNext(p) {
		var packet [7]byte
		packet[0] = 106
		binary.LittleEndian.PutUint16(packet[1:], uint16(p.NetCodeVal))
		index := recipient.PlayerInd // live recipient read between NetCode and flags
		binary.LittleEndian.PutUint32(packet[3:], p.Field3680&0x423)
		r.Send(int32(index), packet[:])
	}
	return nil
}
func netReportStatusNative417630(player *server.Player) int32 {
	r := worldPlayerRuntimeCurrent40A830()
	defer runtime.KeepAlive(player)
	code := uint16(player.NetCodeVal)
	flags := player.Field3680 & 0x423
	var packet [7]byte
	packet[0] = 106
	binary.LittleEndian.PutUint16(packet[1:], code)
	binary.LittleEndian.PutUint32(packet[3:], flags)
	return r.Send(255, packet[:])
}

// Native row geometry comes from the actual C owners, never a packed PE
// pointer field or a Go shadow of the production row table.
type worldTileCell411160 struct {
	Flag, Reserved                 byte
	ReservedWord                   uint16
	Value1, Field2, Field3, Field4 int32
	List5                          *tileListNativeNode411350
	Value6, Field7, Field8, Field9 int32
	List10                         *tileListNativeNode411350
}

func tileAllocNative410F60() int32 {
	cap := worldTileCapacity411160()
	head := worldTileCalloc411160(uintptr(cap), unsafe.Sizeof(unsafe.Pointer(nil)))
	*(*unsafe.Pointer)(worldTileHeadCell411160()) = head
	if head == nil {
		return 0
	}
	for i := int32(0); i < cap; i++ {
		row := worldTileCalloc411160(uintptr(cap), unsafe.Sizeof(worldTileCell411160{}))
		*(*unsafe.Pointer)(unsafe.Add(head, uintptr(i)*unsafe.Sizeof(unsafe.Pointer(nil)))) = row
		if row == nil {
			return 0
		} // preserve partial allocation and no rollback
	}
	return 1
}
func tileFreeNative410FC0() {
	for i := int32(0); i < worldTileCapacity411160(); i++ {
		head := *(*unsafe.Pointer)(worldTileHeadCell411160())
		row := *(*unsafe.Pointer)(unsafe.Add(head, uintptr(i)*unsafe.Sizeof(unsafe.Pointer(nil))))
		if row != nil {
			worldTileFree411160(row)
		}
	}
	// Current C frees rows only; it neither frees nor clears the head/slots.
}
func worldTileRow411160(head unsafe.Pointer, index int32) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Add(head, uintptr(index)*unsafe.Sizeof(unsafe.Pointer(nil))))
}
func tileFromPointNative411160(point *types.Pointf) int32 {
	defer runtime.KeepAlive(point)
	// Unsuffixed C literals promote to double before the named float stores.
	x, y := point.X, point.Y
	i := worldTileFloatToInt411160(float32((float64(x) + 11.5) * 0.021739131))
	j := worldTileFloatToInt411160(float32((float64(y) + 11.5) * 0.021739131))
	dx := worldTileFloatToInt411160(float32(float64(point.X)+11.5)) % 46
	dy := worldTileFloatToInt411160(float32(float64(point.Y)+11.5)) % 46
	if int32(uint32(i)-1) <= 0 || i >= 127 || int32(uint32(j)-1) <= 0 || j >= 127 {
		return -1
	}
	head := *(*unsafe.Pointer)(worldTileHeadCell411160())
	if head == nil {
		return -1
	}
	row, previous := worldTileRow411160(head, i), worldTileRow411160(head, i-1)
	if row == nil || previous == nil {
		return -1
	} // current transition policy
	var cell *worldTileCell411160
	var initial int32
	var list *tileListNativeNode411350
	var xy [2]int32
	if dx <= dy {
		if 46-dx <= dy {
			cell = (*worldTileCell411160)(unsafe.Add(row, uintptr(j)*unsafe.Sizeof(worldTileCell411160{})))
			initial, list = cell.Value6, cell.List10
			xy = [2]int32{dx, dy - 23}
		} else {
			cell = (*worldTileCell411160)(unsafe.Add(previous, uintptr(j)*unsafe.Sizeof(worldTileCell411160{})))
			initial, list = cell.Value1, cell.List5
			xy = [2]int32{dx + 23, dy}
		}
	} else if 46-dx <= dy {
		cell = (*worldTileCell411160)(unsafe.Add(row, uintptr(j)*unsafe.Sizeof(worldTileCell411160{})))
		initial, list = cell.Value1, cell.List5
		xy = [2]int32{dx - 23, dy}
	} else {
		cell = (*worldTileCell411160)(unsafe.Add(row, uintptr(j-1)*unsafe.Sizeof(worldTileCell411160{})))
		initial, list = cell.Value6, cell.List10
		xy = [2]int32{dx, dy + 23}
	}
	if list != nil {
		return sub411350Native(unsafe.Pointer(list), unsafe.Pointer(&xy[0]), initial)
	}
	return initial
}
func tileAllowsTeleportNative411A90(point *types.Pointf) int32 {
	cache := unsafe.Slice((*int32)(memmap.PtrOff(0x587000, 26516)), 5)
	if cache[1] == -1 {
		names := [...]string{"WaterNoTeleport", "WaterDeepNoTeleport", "WaterShallowNoTeleport", "WaterSwampDeepNoTeleport", "WaterSwampShallowNoTeleport"}
		for i, def := range worldTileDefinitions411160() {
			name := def.Name()
			for k, expected := range names {
				if name == expected {
					cache[k] = int32(i)
					break
				}
			}
		}
	}
	value := tileFromPointNative411160(point)
	for _, cached := range cache {
		if value == cached {
			return 1
		}
	}
	return 0
}
