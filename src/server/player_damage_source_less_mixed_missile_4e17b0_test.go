package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Unchanged original callbacks retain a nil source and the actual missile.
// Signed/default rows exercise the ABI, not public root DoDamage.
func TestPlayerDamageSourceLessMixedMissileMatchesGAMEEXE4E17B0(t *testing.T) {
	data, err := os.ReadFile("testdata/player_damage_source_less_mixed_missile_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Player      bool              `json:"player"`
			SourceKind  string            `json:"source_kind"`
			WeaponKind  string            `json:"weapon_kind"`
			WeaponClass object.Class      `json:"weapon_class"`
			Type        object.DamageType `json:"damage_type"`
			Raw         int32             `json:"raw"`
			Forward     uint32            `json:"forward_damage"`
			Marker      uint32            `json:"marker"`
			Kind        uint32            `json:"marker_type"`
			Carry       uint32            `json:"carry_bits"`
			Wear        []uint32          `json:"wear"`
			Return      int               `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 1152 {
		t.Fatal("unexpected original nil-source missile fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("weapon-%x/player-%t/type-%d/raw-%d", uint32(row.WeaponClass), row.Player, row.Type, row.Raw), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, row.Player)
			armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
			var attacker *Object
			modifiers, free := alloc.New(ModifierInitData{})
			*modifiers = ModifierInitData{}
			t.Cleanup(free)
			weapon := &Object{TypeInd: 73, ObjClass: row.WeaponClass, ObjSubClass: 0x10, InitData: unsafe.Pointer(modifiers)}
			if row.Player {
				v.UpdateDataPlayer().Field76, v.UpdateDataPlayer().Field75 = 99, 77
			} else {
				v.UpdateDataMonster().Field547, v.UpdateDataMonster().Field546 = 99, 77
			}
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, .5)
			r.CoopMode = func() bool { return false }
			r.IsEnemy = func(*Object, *Object) bool { return true }
			r.QuestMode = func() bool { return false }
			r.GodMode = func() bool { return false }
			r.GameplayFlag1 = func() bool { return true }
			r.FireProtection = func(*Object) float64 { return 0 }
			r.ElectricArmorScale = func(*Object) float32 { return 1 }
			r.BuffOff = func(*Object, EnchantID) {}
			r.PlayerSetState = func(*Object, PlayerState) bool { return true }
			calls, forward, wear := 0, int32(123456), int32(0)
			r.DefaultDamage = func(target, a, w *Object, d int32, k object.DamageType) bool {
				if target != v || a != attacker || w != weapon || k != row.Type {
					t.Fatal("nil-source missile DefaultDamage arguments")
				}
				calls++
				forward = d
				return true
			}
			r.DamageClear = func(target *Object, d int32) {
				if target != v {
					t.Fatal("wrong HP target")
				}
				calls++
				forward = d
			}
			r.DamageArmor = func(item, a, w *Object, d int32, k object.DamageType) bool {
				if item != armor || a != attacker || w != weapon || k != row.Type {
					t.Fatal("nil-source missile armor arguments")
				}
				wear += d
				return true
			}
			reason := ""
			r.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
			h, ok := PlayerDamageNative4E17B0(v, attacker, weapon, row.Raw, row.Type, r)
			marker, kind, carry := damageMeleeMarker4E17B0(v)
			wantWear := int32(0)
			if len(row.Wear) != 0 {
				wantWear = max(int32(row.Wear[0]), 0)
			}
			if !h || ok != (row.Return != 0) || calls != 1 || uint32(forward) != row.Forward || marker != row.Marker || kind != row.Kind || math.Float32bits(carry) != row.Carry || wear != wantWear || reason != "" {
				t.Fatalf("handled/result=%t/%t forward=%d/%d calls=%d marker=%d/%d carry=%g/%g wear=%d/%d reason=%s", h, ok, forward, int32(row.Forward), calls, marker, kind, carry, math.Float32frombits(row.Carry), wear, wantWear, reason)
			}
		})
	}
}

func TestPlayerDamageSourceLessMixedMissileFrontReflectMatchesGAMEEXE4E17B0(t *testing.T) {
	blob, err := os.ReadFile("testdata/player_damage_source_less_mixed_missile_front_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			WeaponClass object.Class      `json:"weapon_class"`
			Type        object.DamageType `json:"damage_type"`
			Player      bool              `json:"player"`
			Mode        string            `json:"owner"`
			Forward     []uint32          `json:"forward_damage"`
			Marker      uint32            `json:"marker"`
			Kind        uint32            `json:"marker_type"`
			Carry       uint32            `json:"carry_bits"`
			Events      [][]any           `json:"events"`
			Return      int               `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 432 {
		t.Fatal("unexpected original raw front/Reflect fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("class-%x/player-%t/type-%d/%s", uint32(row.WeaponClass), row.Player, row.Type, row.Mode), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, row.Player)
			var a *Object
			w := damageMeleeUnitFixture4E17B0(t, false)
			w.ObjClass, w.ObjSubClass, w.UpdateData, w.HealthData = row.WeaponClass, 0x10, nil, nil
			modifiers, free := alloc.New(ModifierInitData{})
			*modifiers = ModifierInitData{}
			t.Cleanup(free)
			w.InitData = unsafe.Pointer(modifiers)
			w.TypeInd, w.ObjOwner = 73, a
			w.PosVec, w.PrevPos = types.Ptf(20, 0), types.Ptf(-20, 7)
			if row.Player {
				u := v.UpdateDataPlayer()
				u.Field76, u.Field75, u.Field21, u.Field57 = 99, 77, math.Float32bits(.25), math.Float32bits(.5)
			} else {
				u := v.UpdateDataMonster()
				u.Field547, u.Field546, u.Field1, u.Field518 = 99, 77, math.Float32bits(.25), math.Float32bits(.5)
			}
			reflect, front := strings.Contains(row.Mode, "reflect"), strings.HasPrefix(row.Mode, "front")
			if reflect {
				v.Buffs |= 1 << playerDamageReflectEnchant4E17B0
			}
			r := damageMeleeRuntimeFixture4E17B0(t)
			r.CoopMode = func() bool { return false }
			r.ElectricArmorScale = func(*Object) float32 { return 1 }
			r.QuestMode = func() bool { return false }
			var trace, wantTrace []string
			r.ProjectileReflect = func(attack, victim *Object) {
				if attack != w || victim != v {
					t.Fatal("missile reflection identity")
				}
				trace = append(trace, "reflect")
			}
			r.ClearOwner = func(attack *Object) {
				if attack != w {
					t.Fatal("missile clear-owner identity")
				}
				attack.ObjOwner = nil
				trace = append(trace, "clear-owner")
			}
			r.SetOwner = func(owner, attack *Object) {
				if owner != v || attack != w || attack.ObjOwner != nil {
					t.Fatal("missile set-owner identity/order")
				}
				attack.ObjOwner = owner
				trace = append(trace, "set-owner")
			}
			for _, e := range row.Events {
				switch e[0] {
				case "facing", "reflect", "clear-owner", "set-owner":
					wantTrace = append(wantTrace, e[0].(string))
				case "point-fx":
					wantTrace = append(wantTrace, fmt.Sprintf("fx:%d", int(e[1].(float64))))
				case "audio":
					wantTrace = append(wantTrace, fmt.Sprintf("audio:%d", int(e[1].(float64))))
				}
			}
			directions, exclusions := 0, 0
			r.BlockDirection = func(target *Object, pos types.Pointf) bool {
				want := w.PrevPos
				if reflect && directions == 0 {
					want = w.PosVec
				}
				if target != v || pos != want {
					t.Fatal("raw Reflect/ordinary direction snapshot")
				}
				directions++
				trace = append(trace, "facing")
				return front
			}
			r.BlockSourceExcluded = func(source *Object) bool {
				if source != w {
					t.Fatal("raw weapon exclusion source")
				}
				exclusions++
				return false
			}
			r.BlockSourceOnlyExcluded = func(*Object) bool { t.Fatal("nonnull actor weapon used source-only exclusions"); return false }
			r.PointFX = func(id int, pos types.Pointf) {
				if pos != v.PosVec {
					t.Fatal("raw Reflect FX position")
				}
				trace = append(trace, fmt.Sprintf("fx:%d", id))
			}
			r.Audio = func(id int, target *Object) {
				if target != v {
					t.Fatal("raw Reflect audio target")
				}
				trace = append(trace, fmt.Sprintf("audio:%d", id))
			}
			var forward []uint32
			r.DefaultDamage = func(target, source, weapon *Object, d int32, k object.DamageType) bool {
				if target != v || source != a || weapon != w || k != row.Type {
					t.Fatal("raw forward identity")
				}
				forward = append(forward, uint32(d))
				return true
			}
			h, ok := PlayerDamageNative4E17B0(v, a, w, 9, row.Type, r)
			marker, kind, carry := damageMeleeMarker4E17B0(v)
			wantExclusions := 0
			if !h || ok != (row.Return != 0) || fmt.Sprint(forward) != fmt.Sprint(row.Forward) || marker != row.Marker || kind != row.Kind || math.Float32bits(carry) != row.Carry || fmt.Sprint(trace) != fmt.Sprint(wantTrace) || exclusions != wantExclusions {
				t.Fatalf("raw front/Reflect=%t/%t forward=%v/%v marker=%d/%d trace=%v/%v exclusions=%d/%d", h, ok, forward, row.Forward, marker, kind, trace, wantTrace, exclusions, wantExclusions)
			}
		})
	}
}

// Original equipment rows declare block-item wear, state and balance as
// boundaries. Native prefix selection must retain the actual owner/weapon.
func TestPlayerDamageSourceLessMixedMissileBlocksMatchesGAMEEXE4E17B0(t *testing.T) {
	blob, err := os.ReadFile("testdata/player_damage_source_less_mixed_missile_blocks_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			WeaponClass object.Class      `json:"weapon_class"`
			Type        object.DamageType `json:"damage_type"`
			Player      bool              `json:"player"`
			Mode        string            `json:"owner"`
			Forward     []uint32          `json:"forward_damage"`
			Marker      uint32            `json:"marker"`
			Kind        uint32            `json:"marker_type"`
			Carry       uint32            `json:"carry_bits"`
			Blocks      [][]float64       `json:"blocks"`
			State       byte              `json:"final_player_state"`
			Events      [][]any           `json:"events"`
			Return      int               `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 324 {
		t.Fatal("unexpected original block fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("class-%x/player-%t/type-%d/%s", uint32(row.WeaponClass), row.Player, row.Type, row.Mode), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, row.Player)
			var owner *Object
			w := damageMeleeUnitFixture4E17B0(t, false)
			w.ObjClass, w.ObjSubClass, w.UpdateData, w.HealthData = row.WeaponClass, 0x10, nil, nil
			modifiers, free := alloc.New(ModifierInitData{})
			*modifiers = ModifierInitData{}
			t.Cleanup(free)
			w.InitData = unsafe.Pointer(modifiers)
			w.TypeInd, w.ObjOwner, w.PosVec, w.PrevPos = 73, owner, types.Ptf(20, 0), types.Ptf(-20, 7)
			mask, weaponFlags, armorFlags := uint32(2), uint32(0), uint32(0x1000000)
			state := PlayerState16
			if strings.HasSuffix(row.Mode, "sword") {
				mask, weaponFlags, armorFlags, state = 0x400, 0x400, 0, PlayerState13
			}
			if strings.HasSuffix(row.Mode, "staff") {
				mask, weaponFlags, armorFlags, state = 0x8000, 0x8000, 0, PlayerState13
			}
			item, freeItem := alloc.New(Object{})
			*item = Object{ObjClass: object.ClassSimple, ObjSubClass: object.SubClass(mask), ObjFlags: object.FlagEquipped}
			t.Cleanup(freeItem)
			v.InvFirstItem = item
			if row.Player {
				u := v.UpdateDataPlayer()
				u.Field76, u.Field75, u.Field21, u.Field57 = 99, 77, math.Float32bits(.25), math.Float32bits(.5)
				u.Player.WeaponEquip, u.Player.ArmorEquip, u.State = weaponFlags, armorFlags, state
			} else {
				u := v.UpdateDataMonster()
				u.Field547, u.Field546, u.Field1, u.Field518 = 99, 77, math.Float32bits(.25), math.Float32bits(.5)
				u.WeaponEquipFlags, u.ArmorEquipFlags, u.AIStackInd = weaponFlags, armorFlags, 0
				u.AIStack[0].Action = uint32(ai.ACTION_IDLE)
				if armorFlags != 0 {
					u.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
				}
			}
			r := damageMeleeRuntimeFixture4E17B0(t)
			r.CoopMode = func() bool { return false }
			r.QuestMode = func() bool { return false }
			r.GodMode = func() bool { return false }
			r.ElectricArmorScale = func(*Object) float32 { return 1 }
			var trace, wantTrace []string
			r.ProjectileReflect = func(attack, victim *Object) {
				if attack != w || victim != v {
					t.Fatal("missile reflection identity")
				}
				trace = append(trace, "reflect")
			}
			r.ClearOwner = func(attack *Object) {
				if attack != w {
					t.Fatal("missile clear-owner identity")
				}
				attack.ObjOwner = nil
				trace = append(trace, "clear-owner")
			}
			r.SetOwner = func(owner, attack *Object) {
				if owner != v || attack != w || attack.ObjOwner != nil {
					t.Fatal("missile set-owner identity/order")
				}
				attack.ObjOwner = owner
				trace = append(trace, "set-owner")
			}
			for _, e := range row.Events {
				switch e[0] {
				case "facing", "balance", "block", "reflect", "clear-owner", "set-owner":
					wantTrace = append(wantTrace, e[0].(string))
				case "audio", "player-state", "monster-action", "default":
					wantTrace = append(wantTrace, fmt.Sprintf("%s:%d", e[0].(string), int(e[1].(float64))))
				}
			}
			r.BlockSourceExcluded = func(attack *Object) bool {
				if attack != w {
					t.Fatal("wrong excluded attack")
				}
				return false
			}
			r.BlockSourceOnlyExcluded = func(*Object) bool { t.Fatal("nonnull weapon used source-only exclusions"); return false }
			r.BlockDirection = func(target *Object, pos types.Pointf) bool {
				marker, kind, _ := damageMeleeMarker4E17B0(v)
				if target != v || pos != w.PrevPos || marker != 1 || kind != 73 {
					t.Fatal("weapon facing/attribution order")
				}
				trace = append(trace, "facing")
				return true
			}
			r.Audio = func(id int, target *Object) {
				if target != v {
					t.Fatal("wrong block sound target")
				}
				trace = append(trace, fmt.Sprintf("audio:%d", id))
			}
			r.PlayerSetState = func(target *Object, state PlayerState) bool {
				if target != v {
					t.Fatal("wrong block state target")
				}
				trace = append(trace, fmt.Sprintf("player-state:%d", state))
				target.UpdateDataPlayer().State = state
				return true
			}
			r.Melee.RandomInt = func(lo, hi int) int {
				if lo != 18 || hi != 20 {
					t.Fatal("wrong block random interval")
				}
				return 18
			}
			r.Melee.MonsterBlockAction = func(target *Object) {
				if target != v {
					t.Fatal("wrong NPC block target")
				}
				trace = append(trace, "monster-action:23")
			}
			r.BlockDamagePercent = func() float64 { trace = append(trace, "balance"); return .25 }
			blocks := 0
			block := func(got, target, source, weapon *Object, amount float32, typ object.DamageType) bool {
				if got != item || target != v || source != owner || weapon != w || typ != row.Type || len(row.Blocks) != 1 || math.Float32bits(amount) != math.Float32bits(float32(row.Blocks[0][1])) || uint32(row.Blocks[0][0])&mask == 0 {
					t.Fatal("block item/owner/weapon/wear identity")
				}
				blocks++
				trace = append(trace, "block")
				return true
			}
			r.CanDamageBlockItem = func(got *Object) bool { return got == item }
			r.DamageBlockItem = block
			r.Melee.CanDamageBlockWeapon = r.CanDamageBlockItem
			r.Melee.DamageBlockWeapon = block
			var forward []uint32
			r.DefaultDamage = func(target, source, weapon *Object, d int32, typ object.DamageType) bool {
				if target != v || source != owner || weapon != w || typ != row.Type {
					t.Fatal("unblocked damage identity")
				}
				forward = append(forward, uint32(d))
				trace = append(trace, fmt.Sprintf("default:%d", d))
				return true
			}
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) }
			h, ok := PlayerDamageNative4E17B0(v, owner, w, 9, row.Type, r)
			marker, kind, carry := damageMeleeMarker4E17B0(v)
			if !h || ok != (row.Return != 0) || marker != row.Marker || kind != row.Kind || math.Float32bits(carry) != row.Carry || fmt.Sprint(forward) != fmt.Sprint(row.Forward) || blocks != len(row.Blocks) || fmt.Sprint(trace) != fmt.Sprint(wantTrace) {
				t.Fatalf("block result=%t/%t marker=%d/%d carry=%g forward=%v/%v blocks=%d trace=%v/%v", h, ok, marker, kind, carry, forward, row.Forward, blocks, trace, wantTrace)
			}
			if row.Player && byte(v.UpdateDataPlayer().State) != row.State {
				t.Fatal("original block state")
			}
		})
	}
}
