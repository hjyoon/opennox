package legacy

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerAttackThrown538960InitializesPlayerAnimation(t *testing.T) {
	for _, tc := range []struct {
		name string
		flag object.WeaponClass
	}{
		{"chakram", object.WeaponChakram},
		{"shuriken", object.WeaponShuriken},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := installPlayerAttackProjectileServer538960(t)
			oldAnimation := playerAnimFrames4F9F90
			playerAnimFrames4F9F90 = func(action int) (int, int) {
				if action != 44 {
					t.Fatalf("throw animation = %d, want 44", action)
				}
				return 4, 0
			}
			t.Cleanup(func() { playerAnimFrames4F9F90 = oldAnimation })

			unit := &server.Object{ObjClass: object.ClassPlayer}
			player := &server.Player{WeaponEquip: uint32(tc.flag)}
			player.Info().SetField2239(37)
			weapon := &server.Object{
				TypeInd:     0x3210,
				ObjClass:    object.ClassWeapon,
				ObjSubClass: object.SubClass(tc.flag),
			}
			update := &server.PlayerUpdateData{
				Player: player, EquippedWeapon: weapon,
				Field59_1: 0x65, Field59_2: 0x9876,
			}
			unit.UpdateData = unsafe.Pointer(update)
			modifier := &server.Modifier{TypeInd: uint32(weapon.TypeInd)}
			srv.Modif.Dword_5d4594_251600 = modifier

			var pin runtime.Pinner
			defer pin.Unpin()
			pinPlayerAttackProjectilePointers538960(t, &pin,
				unsafe.Pointer(unit), unsafe.Pointer(player), unsafe.Pointer(weapon),
				unsafe.Pointer(update), unsafe.Pointer(modifier),
			)
			if got := playerAttackNativeEntry538960(unit); got != 1 {
				t.Fatalf("native throw start = %d, want 1", got)
			}
			if update.Field0 != 4 || update.Field59_0 != 0 {
				t.Fatalf("throw deadline/frame = %d/%d, want 4/0", update.Field0, update.Field59_0)
			}
			srv.SetFrame(4)
			if got := playerAttackNativeEntry538960(unit); got != 0 {
				t.Fatalf("native throw finish = %d, want 0", got)
			}
			if update.Field0 != 4 || update.Field59_0 != 3 || update.Field59_1 != 0x65 || update.Field59_2 != 0x9876 {
				t.Fatalf("throw finish changed deadline/frame/neighbors: %+v", *update)
			}
		})
	}
}

type playerAttackThrownFixture538960 struct {
	owner, weapon, projectile *server.Object
	update                    *server.PlayerUpdateData
	player                    *server.Player
	attrs                     *server.ModifierInitData
	ammo                      *server.AmmoUseData
	chakram                   *server.ChakramUpdateData
	arrow                     *server.ArrowCollideData
	flag                      object.WeaponClass
	deps                      playerAttackThrownRuntime538960
	order                     []string
	pin                       runtime.Pinner
}

func newPlayerAttackThrownFixture538960(t *testing.T, flag object.WeaponClass) *playerAttackThrownFixture538960 {
	t.Helper()
	srv := installPlayerAttackProjectileServer538960(t)
	f := &playerAttackThrownFixture538960{
		owner: &server.Object{
			ObjClass: object.ClassPlayer, PosVec: types.Ptf(100, 200), Direction1: 17,
		},
		weapon: &server.Object{
			TypeInd: 0x3211, ObjClass: object.ClassWeapon, ObjSubClass: object.SubClass(flag),
		},
		projectile: &server.Object{SpeedCur: 7},
		update:     &server.PlayerUpdateData{Field59_0: 1, Field59_1: 0x65, Field59_2: 0x9876},
		player:     &server.Player{WeaponEquip: uint32(flag), PlayerInd: 3},
		attrs:      &server.ModifierInitData{},
		ammo:       &server.AmmoUseData{Charge0: 40, Charge1: 20},
		chakram:    &server.ChakramUpdateData{},
		arrow:      &server.ArrowCollideData{},
		flag:       flag,
	}
	f.owner.Shape.Circle.R = 6
	f.owner.UpdateData = unsafe.Pointer(f.update)
	f.player.Info().SetField2239(37)
	f.update.Player, f.update.EquippedWeapon = f.player, f.weapon
	f.weapon.InitData, f.weapon.InvHolder = unsafe.Pointer(f.attrs), f.owner
	if flag == object.WeaponChakram {
		f.projectile.UpdateData = unsafe.Pointer(f.chakram)
	} else {
		f.weapon.UseData.Ptr = unsafe.Pointer(f.ammo)
		f.projectile.CollideData = unsafe.Pointer(f.arrow)
	}
	modifier := &server.Modifier{TypeInd: uint32(f.weapon.TypeInd)}
	srv.Modif.Dword_5d4594_251600 = modifier
	pinPlayerAttackProjectilePointers538960(t, &f.pin,
		unsafe.Pointer(f.owner), unsafe.Pointer(f.weapon), unsafe.Pointer(f.projectile),
		unsafe.Pointer(f.update), unsafe.Pointer(f.player), unsafe.Pointer(f.attrs),
		unsafe.Pointer(f.ammo), unsafe.Pointer(f.chakram), unsafe.Pointer(f.arrow), unsafe.Pointer(modifier),
	)
	t.Cleanup(f.pin.Unpin)
	cosine, sine := server.SinCosDir(17)
	wantSpawn := types.Ptf(float32(100+10*float64(cosine)), float32(200+10*float64(sine)))
	f.deps = playerAttackThrownRuntime538960{
		Frame: func() uint32 { return 2 },
		AnimFrames: func(action int) (int, int) {
			if action != 44 {
				t.Fatalf("throw action = %d, want 44", action)
			}
			return 4, 0
		},
		Direction: func(direction server.Dir16) types.Pointf {
			cosine, sine := server.SinCosDir(byte(direction))
			return types.Ptf(cosine, sine)
		},
		Trace: func(from, to types.Pointf, flags server.MapTraceFlags) bool {
			f.order = append(f.order, "trace")
			wantFlags := server.MapTraceFlags(4)
			if flag == object.WeaponChakram {
				wantFlags = 5
			}
			if from != types.Ptf(100, 200) || to != wantSpawn || flags != wantFlags {
				t.Fatalf("throw trace = %v/%v/%d, want (100,200)/%v/%d", from, to, flags, wantSpawn, wantFlags)
			}
			return true
		},
		NewObject: func(name string) *server.Object {
			f.order = append(f.order, "new")
			want := "FanChakramInMotion"
			if flag == object.WeaponChakram {
				want = "RoundChakramInMotion"
			}
			if name != want {
				t.Fatalf("throw projectile = %q, want %q", name, want)
			}
			return f.projectile
		},
		CreateAt: func(projectile, owner *server.Object, pos types.Pointf) {
			f.order = append(f.order, "create")
			if projectile != f.projectile || owner != f.owner || pos != wantSpawn {
				t.Fatalf("throw creation = %p/%p/%v", projectile, owner, pos)
			}
			if flag == object.WeaponShuriken && f.arrow.Owner != owner {
				t.Fatal("shuriken collision owner was not stored before creation")
			}
			projectile.ObjOwner, projectile.PosVec = owner, pos
		},
		ApplyModifierAttrs: func(projectile *server.Object, attrs *server.ModifierInitData) {
			f.order = append(f.order, "attrs")
			if projectile != f.projectile || attrs != f.attrs {
				t.Fatalf("throw modifiers = %p/%p, want %p/%p", projectile, attrs, f.projectile, f.attrs)
			}
		},
		DetachInventory: func(owner, weapon *server.Object) {
			f.order = append(f.order, "detach")
			if owner != f.owner || weapon != f.weapon {
				t.Fatalf("throw detach = %p/%p", owner, weapon)
			}
			weapon.InvHolder = nil
		},
		InventoryPut: func(projectile, weapon *server.Object, report bool) {
			f.order = append(f.order, "put")
			if projectile != f.projectile || weapon != f.weapon || !report || weapon.InvHolder != nil {
				t.Fatalf("chakram inventory transfer = %p/%p/%t", projectile, weapon, report)
			}
			weapon.InvHolder = projectile
		},
		AudioEvent: func(id sound.ID, owner *server.Object) {
			f.order = append(f.order, "audio")
			if id != sound.ID(891) || owner != f.owner {
				t.Fatalf("throw sound = %d/%p", id, owner)
			}
		},
		DelayedDelete: func(weapon *server.Object) {
			f.order = append(f.order, "delete")
			if weapon != f.weapon {
				t.Fatalf("deleted shuriken = %p, want %p", weapon, f.weapon)
			}
		},
		ReportCharges: func(index uint8, weapon *server.Object, charge1, charge0 uint8) {
			f.order = append(f.order, "charges")
			if index != f.update.Player.PlayerInd || weapon != f.weapon || charge1 != f.ammo.Charge1 || charge0 != f.ammo.Charge0 {
				t.Fatalf("shuriken charge packet = %d/%p/%d/%d", index, weapon, charge1, charge0)
			}
		},
		AutoEquipShuriken: func(owner *server.Object) {
			f.order = append(f.order, "auto-equip")
			if owner != f.owner {
				t.Fatalf("shuriken auto-equip owner = %p", owner)
			}
		},
	}
	return f
}

func (f *playerAttackThrownFixture538960) call() int {
	return playerAttackThrownNative538960(f.owner, f.weapon, f.update, uint32(f.flag), f.update.Field59_0, f.deps)
}

func installPlayerAttackThrownRuntime538960(t *testing.T, f *playerAttackThrownFixture538960) {
	t.Helper()
	old := playerAttackThrownRuntimeFactory538960
	playerAttackThrownRuntimeFactory538960 = func() playerAttackThrownRuntime538960 { return f.deps }
	t.Cleanup(func() { playerAttackThrownRuntimeFactory538960 = old })
}

func TestPlayerAttackThrown538960NativeChakramCachesUpdateAndReadsFreshFields(t *testing.T) {
	f := newPlayerAttackThrownFixture538960(t, object.WeaponChakram)
	installPlayerAttackThrownRuntime538960(t, f)
	replacementData := &server.ChakramUpdateData{Reflections: 99, ReturnState: 77}
	replacementAttrs := &server.ModifierInitData{}
	pinPlayerAttackProjectilePointers538960(t, &f.pin, unsafe.Pointer(replacementData), unsafe.Pointer(replacementAttrs))
	detach := f.deps.DetachInventory
	f.deps.DetachInventory = func(owner, weapon *server.Object) {
		detach(owner, weapon)
		f.projectile.UpdateData = unsafe.Pointer(replacementData)
		f.owner.PosVec = types.Ptf(300, 400)
	}
	put := f.deps.InventoryPut
	f.deps.InventoryPut = func(projectile, weapon *server.Object, report bool) {
		put(projectile, weapon, report)
		f.attrs = replacementAttrs
		weapon.InitData = unsafe.Pointer(replacementAttrs)
	}
	attrs := f.deps.ApplyModifierAttrs
	f.deps.ApplyModifierAttrs = func(projectile *server.Object, data *server.ModifierInitData) {
		attrs(projectile, data)
		f.owner.Direction1 = 0xab34
		projectile.SpeedCur = 10.5
	}
	direction := f.deps.Direction
	f.deps.Direction = func(dir server.Dir16) types.Pointf {
		if dir == 17 {
			return direction(dir)
		}
		if dir != 0xab34 {
			t.Fatalf("fresh direction lost its WORD: %#x", dir)
		}
		return types.Ptf(0.25, -0.5)
	}
	if got := playerAttackNativeEntry538960(f.owner); got != 1 {
		t.Fatalf("native chakram launch = %d, want 1", got)
	}
	if f.chakram.Reflections != 4 || f.chakram.OwnerPos != types.Ptf(300, 400) || f.chakram.ReturnState != 2 ||
		replacementData.Reflections != 99 || replacementData.ReturnState != 77 {
		t.Fatalf("cached/fresh chakram data = %+v / %+v", *f.chakram, *replacementData)
	}
	if f.projectile.VelVec != types.Ptf(2.625, -5.25) || f.projectile.Direction1 != 0xab34 || f.projectile.Direction2 != 0xab34 ||
		f.weapon.InvHolder != f.projectile {
		t.Fatalf("chakram motion/holder = %v/%#x/%#x/%p", f.projectile.VelVec, f.projectile.Direction1, f.projectile.Direction2, f.weapon.InvHolder)
	}
	if f.update.Field0 != 6 || f.update.Field59_0 != 2 || f.update.Field59_1 != 0x65 || f.update.Field59_2 != 0x9876 {
		t.Fatalf("chakram deadline/frame/neighbors = %d/%d/%#x/%#x", f.update.Field0, f.update.Field59_0, f.update.Field59_1, f.update.Field59_2)
	}
	if want := []string{"trace", "new", "detach", "create", "put", "attrs", "audio"}; !reflect.DeepEqual(f.order, want) {
		t.Fatalf("chakram order = %v, want %v", f.order, want)
	}
	before := len(f.order)
	if got := playerAttackNativeEntry538960(f.owner); got != 1 || len(f.order) != before {
		t.Fatalf("duplicate midpoint: result=%d calls=%v", got, f.order)
	}
}

func TestPlayerAttackThrown538960NativeShurikenCachesAmmoAndUpdate(t *testing.T) {
	f := newPlayerAttackThrownFixture538960(t, object.WeaponShuriken)
	installPlayerAttackThrownRuntime538960(t, f)
	replacementAmmo := &server.AmmoUseData{Charge0: 77, Charge1: 88}
	replacementAttrs := &server.ModifierInitData{}
	replacementPlayer := &server.Player{PlayerInd: 9}
	decoyPlayer := &server.Player{PlayerInd: 13}
	decoyUpdate := &server.PlayerUpdateData{Player: decoyPlayer, Field0: 123, Field59_0: 42}
	pinPlayerAttackProjectilePointers538960(t, &f.pin, unsafe.Pointer(replacementAmmo), unsafe.Pointer(replacementAttrs),
		unsafe.Pointer(replacementPlayer), unsafe.Pointer(decoyPlayer), unsafe.Pointer(decoyUpdate))
	trace := f.deps.Trace
	f.deps.Trace = func(from, to types.Pointf, flags server.MapTraceFlags) bool {
		f.weapon.UseData.Ptr = unsafe.Pointer(replacementAmmo)
		return trace(from, to, flags)
	}
	create := f.deps.CreateAt
	f.deps.CreateAt = func(projectile, owner *server.Object, pos types.Pointf) {
		create(projectile, owner, pos)
		f.attrs = replacementAttrs
		f.weapon.InitData = unsafe.Pointer(replacementAttrs)
	}
	audio := f.deps.AudioEvent
	f.deps.AudioEvent = func(id sound.ID, owner *server.Object) {
		audio(id, owner)
		f.ammo.Charge0 = 39
		f.update.Player = replacementPlayer
		owner.UpdateData = unsafe.Pointer(decoyUpdate)
	}
	if got := playerAttackNativeEntry538960(f.owner); got != 1 {
		t.Fatalf("native shuriken launch = %d, want 1", got)
	}
	if f.ammo.Charge1 != 19 || replacementAmmo.Charge1 != 88 || f.update.Field59_0 != 2 ||
		decoyUpdate.Field59_0 != 42 || decoyUpdate.Field0 != 123 || f.update.Field0 != 6 || f.arrow.Owner != f.owner {
		t.Fatalf("cached shuriken ammo/frame = %+v / %+v / %d / %d", *f.ammo, *replacementAmmo, f.update.Field59_0, decoyUpdate.Field59_0)
	}
	if want := []string{"trace", "new", "create", "attrs", "audio", "charges"}; !reflect.DeepEqual(f.order, want) {
		t.Fatalf("shuriken order = %v, want %v", f.order, want)
	}
}

func TestPlayerAttackThrown538960FrameGatesAndDeadline(t *testing.T) {
	for _, tc := range []struct {
		name                          string
		frame, started, deadline      uint32
		frames, duration              int
		previous, stored              uint8
		wantDeadline                  uint32
		wantResult, wantTrace, clocks int
	}{
		{"start", 100, 100, 0, 4, 0, 0, 0, 104, 1, 0, 2},
		{"midpoint", 102, 100, 0, 4, 0, 1, 2, 106, 1, 1, 2},
		{"same midpoint", 102, 100, 0, 4, 0, 2, 2, 106, 1, 0, 2},
		{"skipped midpoint", 103, 100, 0, 4, 0, 1, 3, 107, 1, 0, 2},
		{"finish clamps byte", 104, 100, 0, 4, 0, 2, 3, 108, 0, 0, 2},
		{"keep deadline", 100, 100, 77, 4, 0, 0, 0, 77, 1, 0, 1},
		{"DWORD wrap", math.MaxUint32, math.MaxUint32 - 2, 0, 4, 0, 1, 2, 3, 1, 1, 2},
		{"duration", 106, 100, 0, 4, 2, 1, 2, 118, 1, 1, 2},
		{"BYTE wrap", 258, 0, 0, 4, 0, 1, 2, 262, 1, 1, 2},
		{"do not truncate midpoint", 2, 0, 0, 516, 0, 1, 2, 518, 1, 0, 2},
		{"zero frames", 100, 100, 0, 0, 0, 0, 255, 100, 0, 0, 2},
		{"negative frames", 100, 100, 0, -4, 0, 0, 251, 96, 0, 0, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newPlayerAttackThrownFixture538960(t, object.WeaponShuriken)
			f.owner.Field34 = tc.started
			f.update.Field0, f.update.Field59_0 = tc.deadline, tc.previous
			var clocks, traces int
			f.deps.Frame = func() uint32 { clocks++; return tc.frame }
			f.deps.AnimFrames = func(action int) (int, int) { return tc.frames, tc.duration }
			f.deps.Trace = func(from, to types.Pointf, flags server.MapTraceFlags) bool { traces++; return false }
			f.deps.AudioEvent = func(id sound.ID, owner *server.Object) {
				if id != 323 || owner != f.owner {
					t.Fatalf("blocked throw sound = %d/%p", id, owner)
				}
			}
			if got := f.call(); got != tc.wantResult || f.update.Field0 != tc.wantDeadline ||
				f.update.Field59_0 != tc.stored || traces != tc.wantTrace || clocks != tc.clocks {
				t.Fatalf("frame gate result/deadline/frame/trace/clocks = %d/%d/%d/%d/%d, want %d/%d/%d/%d/%d",
					got, f.update.Field0, f.update.Field59_0, traces, clocks, tc.wantResult, tc.wantDeadline, tc.stored, tc.wantTrace, tc.clocks)
			}
		})
	}
}

func TestPlayerAttackThrown538960FailurePathsAndChakramPriority(t *testing.T) {
	for _, flag := range []object.WeaponClass{object.WeaponChakram, object.WeaponShuriken} {
		t.Run(flag.String(), func(t *testing.T) {
			f := newPlayerAttackThrownFixture538960(t, flag)
			f.deps.NewObject = func(name string) *server.Object { f.order = append(f.order, "new"); return nil }
			if got := f.call(); got != 0 || f.update.Field59_0 != 1 || f.update.Field0 != 6 ||
				f.ammo.Charge1 != 20 || !reflect.DeepEqual(f.order, []string{"trace", "new"}) {
				t.Fatalf("allocation failure = %d, frame/deadline=%d/%d, calls=%v", got, f.update.Field59_0, f.update.Field0, f.order)
			}
		})
	}
	f := newPlayerAttackThrownFixture538960(t, object.WeaponChakram)
	f.flag |= object.WeaponShuriken | object.WeaponSword
	if got := f.call(); got != 1 || f.ammo.Charge1 != 20 || f.weapon.InvHolder != f.projectile {
		t.Fatalf("Round priority = %d, ammo=%d holder=%p", got, f.ammo.Charge1, f.weapon.InvHolder)
	}
}

func TestPlayerAttackThrown538960ShurikenConsumption(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		charge, infinite, want uint8
		after                  []string
	}{
		{"remaining", 20, 0, 19, []string{"charges"}},
		{"last shuriken", 1, 0, 0, []string{"detach", "delete", "auto-equip"}},
		{"zero wraps", 0, 0, 255, []string{"charges"}},
		{"infinite one", 20, 1, 20, nil},
		{"any nonzero infinite", 20, 255, 20, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newPlayerAttackThrownFixture538960(t, object.WeaponShuriken)
			f.ammo.Charge1, f.ammo.Field2 = tc.charge, tc.infinite
			if got := f.call(); got != 1 || f.ammo.Charge1 != tc.want {
				t.Fatalf("shuriken consumption = %d/%d, want 1/%d", got, f.ammo.Charge1, tc.want)
			}
			want := append([]string{"trace", "new", "create", "attrs", "audio"}, tc.after...)
			if !reflect.DeepEqual(f.order, want) {
				t.Fatalf("consumption order = %v, want %v", f.order, want)
			}
		})
	}
	f := newPlayerAttackThrownFixture538960(t, object.WeaponShuriken)
	audio := f.deps.AudioEvent
	f.deps.AudioEvent = func(id sound.ID, owner *server.Object) { audio(id, owner); f.ammo.Field2 = 2 }
	if got := f.call(); got != 1 || f.ammo.Charge1 != 20 || len(f.order) != 5 {
		t.Fatalf("infinite flag was not read after audio: result=%d ammo=%+v calls=%v", got, *f.ammo, f.order)
	}
}

func TestPlayerAttackThrown538960RetainsFinalCoordinateSpillAndFullDirectionIndex(t *testing.T) {
	f := newPlayerAttackThrownFixture538960(t, object.WeaponShuriken)
	// This cancellation distinguishes radius/product binary32 spills from
	// the original x87 expression's final coordinate store.
	f.owner.Shape.Circle.R = 16777218
	f.owner.PosVec = types.Ptf(-16777216, -16777216)
	f.deps.Direction = func(dir server.Dir16) types.Pointf { return types.Ptf(0.99999994, 0.99999994) }
	f.deps.Trace = func(from, to types.Pointf, flags server.MapTraceFlags) bool {
		if math.Float32bits(to.X) != 0x409fffff || math.Float32bits(to.Y) != 0x409fffff {
			t.Fatalf("final binary32 coordinate = %08x/%08x, want 409fffff/409fffff", math.Float32bits(to.X), math.Float32bits(to.Y))
		}
		return false
	}
	f.deps.AudioEvent = func(sound.ID, *server.Object) {}
	f.call()

	// Exercise the real table binding, including indices that a byte cast
	// would incorrectly map to 255 and 0. These are isolated memmap slots.
	direction := playerAttackThrownRuntimeFactory538960().Direction
	for _, dir := range []server.Dir16{0xffff, 0x100} {
		offset := uintptr(int(int16(dir)) * 8)
		x, y := memmap.PtrFloat32(0x587000, 194136+offset), memmap.PtrFloat32(0x587000, 194140+offset)
		oldX, oldY := *x, *y
		*x, *y = 0.125, -0.375
		got := direction(dir)
		*x, *y = oldX, oldY
		if got != types.Ptf(0.125, -0.375) {
			t.Fatalf("signed WORD table index %#x = %v", dir, got)
		}
	}
}

func TestPlayerAttackThrown538960KeepsOriginalFaultsAndLiveClassStores(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*playerAttackThrownFixture538960)
	}{
		{"missing trace binding", func(f *playerAttackThrownFixture538960) { f.deps.Trace = nil }},
		{"missing creation binding", func(f *playerAttackThrownFixture538960) { f.deps.CreateAt = nil }},
		{"missing modifier binding", func(f *playerAttackThrownFixture538960) { f.deps.ApplyModifierAttrs = nil }},
		{"missing charge binding", func(f *playerAttackThrownFixture538960) { f.deps.ReportCharges = nil }},
		{"missing ammo", func(f *playerAttackThrownFixture538960) { f.weapon.UseData.Ptr = nil }},
		{"missing collision data", func(f *playerAttackThrownFixture538960) { f.projectile.CollideData = nil }},
		{"missing live player", func(f *playerAttackThrownFixture538960) { f.update.Player = nil }},
		{"zero frame divisor", func(f *playerAttackThrownFixture538960) { f.deps.AnimFrames = func(int) (int, int) { return 4, -1 } }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newPlayerAttackThrownFixture538960(t, object.WeaponShuriken)
			tc.edit(f)
			defer func() {
				if recover() == nil {
					t.Fatal("original fault was silently converted into a no-op")
				}
			}()
			f.call()
		})
	}
	f := newPlayerAttackThrownFixture538960(t, object.WeaponShuriken)
	audio := f.deps.AudioEvent
	f.deps.AudioEvent = func(id sound.ID, owner *server.Object) { audio(id, owner); owner.ObjClass = 0 }
	if got := f.call(); got != 1 || f.update.Field59_0 != 1 || len(f.order) != 5 || f.ammo.Charge1 != 19 {
		t.Fatalf("non-player live class store/report = %d/%d/%v/%d", got, f.update.Field59_0, f.order, f.ammo.Charge1)
	}
	f = newPlayerAttackThrownFixture538960(t, object.WeaponChakram)
	// A callback can change the live class while the original cached update
	// pointer survives. Give that alias enough storage for either native
	// layout; a stand-alone PlayerUpdateData is smaller than the NPC layout.
	backing := &struct {
		player server.PlayerUpdateData
		rest   [unsafe.Sizeof(server.MonsterUpdateData{})]byte
	}{player: *f.update}
	f.update = &backing.player
	f.owner.UpdateData = unsafe.Pointer(f.update)
	f.pin.Pin(backing)
	audio = f.deps.AudioEvent
	f.deps.AudioEvent = func(id sound.ID, owner *server.Object) {
		audio(id, owner)
		owner.ObjClass, owner.ObjSubClass = object.ClassMonster, object.SubClass(object.MonsterNPC)
	}
	monster := (*server.MonsterUpdateData)(unsafe.Pointer(f.update))
	monster.Field481 = 0xdeadbe01
	monster.Field120_1 = 1
	if got := f.call(); got != 1 || f.update.Field59_0 != 1 || monster.Field120_1 != 2 || monster.Field481 != 0xdeadbe01 {
		t.Fatalf("live NPC frame store = %d/player-%d/NPC-%d/unrelated-%#x", got, f.update.Field59_0, monster.Field120_1, monster.Field481)
	}
}
