package server

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

type attackModifierOriginalRow struct {
	Name, Entry, Mutation string
	HasItem               bool `json:"has_item"`
	OwnerClass            int  `json:"owner_class"`
	Gameplay, Enemy       uint32
	Buff23, Buff27        uint32
	Slots                 [4]int
	FinalSlots            [4]int `json:"final_slots"`
	Active                []int
	FinalActive           []int `json:"final_active"`
	Context               bool
	Result                uint32
	Trace                 [][]any
	ChangedBank           bool `json:"changed_bank"`
}

func TestAttackModifierFrozenOriginal(t *testing.T) {
	data, err := os.ReadFile("testdata/attack_modifier_original_v1.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != "9f673b4b3ab5cbc5fb79eac2065541e0d1fa5dcc380b797f8a9809e038a56fb3" {
		t.Fatalf("unexpected original corpus hash %s", got)
	}
	input, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	var original struct {
		Code       map[string]int `json:"executed_bytes"`
		Executions map[string]int
		Guards     bool `json:"guards_verified"`
		Rows       []attackModifierOriginalRow
	}
	if err := json.NewDecoder(input).Decode(&original); err != nil {
		t.Fatal(err)
	}
	if !original.Guards || original.Code["pre-hit"] != 159 || original.Code["apply"] != 73 ||
		original.Executions["pre-hit"] != 69 || original.Executions["apply"] != 20 || len(original.Rows) != 89 {
		t.Fatal("unexpected PE32 attack modifier corpus scope")
	}
	for _, row := range original.Rows {
		t.Run(row.Name, func(t *testing.T) { checkAttackModifierRow(t, row) })
	}
}

func checkAttackModifierRow(t *testing.T, row attackModifierOriginalRow) {
	t.Helper()
	var markers [5]byte
	var mods [6]*ModifierEff
	preHit := row.Entry == "pre-hit"
	callback := func(mod *ModifierEff) *ModifierEffFnc {
		if preHit {
			return &mod.AttackPreHit52
		}
		return &mod.Attack40
	}
	for i := 1; i < len(mods); i++ {
		mods[i] = new(ModifierEff)
		if slices.Contains(row.Active, i) {
			callback(mods[i]).Fnc = unsafe.Pointer(&markers[i-1])
		}
	}
	var bank, replacement ModifierInitData
	var final [4]int
	for i, index := range row.Slots {
		bank.Modifiers[i] = mods[index]
		replacement.Modifiers[i] = mods[5]
	}
	weapon := &Object{InitData: unsafe.Pointer(&bank)}
	if !row.HasItem {
		weapon = nil
	}
	target := &Object{}
	var owner *Object
	if row.OwnerClass >= 0 {
		owner = &Object{ObjClass: object.Class(row.OwnerClass)}
	}
	var ctxWord uint32
	var context unsafe.Pointer
	if row.Context {
		context = unsafe.Pointer(&ctxWord)
	}
	events := make([][]any, 0)
	mutated := false
	invoke := func(fn unsafe.Pointer, mod *ModifierEff, item, itemOwner, victim *Object, attack unsafe.Pointer) {
		index := slices.Index(mods[:], mod)
		if index <= 0 || fn != unsafe.Pointer(&markers[index-1]) || item != weapon || itemOwner != owner || attack != context ||
			(preHit && victim != target) || (!preHit && victim != nil) {
			t.Fatal("modifier or callback arguments changed")
		}
		events = append(events, []any{"callback", index})
		if mutated || row.Mutation == "" {
			return
		}
		mutated = true
		slot := slices.Index(bank.Modifiers[:], mod) + 1
		switch row.Mutation {
		case "replace-item-bank":
			weapon.InitData = unsafe.Pointer(&replacement)
		case "replace-next-slot":
			bank.Modifiers[slot] = mods[5]
		case "clear-next-callback":
			callback(bank.Modifiers[slot]).Fnc = nil
		case "erase-next-slot":
			bank.Modifiers[slot] = nil
		default:
			t.Fatal(row.Mutation)
		}
	}
	var result int32
	if preHit {
		result = playerPreAttackEffects538290(target, owner, weapon, context, preAttackEffectsRuntime538290{
			gameplay: func() bool {
				events = append(events, []any{"gameplay", 1})
				return row.Gameplay != 0
			},
			enemy: func(a, b *Object) bool {
				if a != owner || b != target {
					t.Fatal("allegiance used different objects")
				}
				events = append(events, []any{"enemy"})
				return row.Enemy != 0
			},
			buff: func(unit *Object, id int32) int32 {
				if unit != target {
					t.Fatal("buff used a different target")
				}
				events = append(events, []any{"buff", id})
				if id == 23 {
					return int32(row.Buff23)
				}
				return int32(row.Buff27)
			},
			invoke: invoke,
		})
	} else {
		result = itemApplyAttackEffect538840(weapon, owner, context, func(fn unsafe.Pointer, mod *ModifierEff, item, itemOwner *Object, attack unsafe.Pointer) {
			invoke(fn, mod, item, itemOwner, nil, attack)
		})
	}
	for i, mod := range bank.Modifiers {
		final[i] = slices.Index(mods[:], mod)
	}
	var active []int
	for i, mod := range mods[1:] {
		if callback(mod).Fnc != nil {
			active = append(active, i+1)
		}
	}
	want := make([][]any, len(row.Trace))
	for i, event := range row.Trace {
		want[i] = event
		if event[0] == "callback" {
			// The real pointers above cover the callback arguments. The slot
			// number comes from the original function's private stack counter.
			want[i] = []any{event[0], event[2]}
		}
	}
	actualTrace, _ := json.Marshal(events)
	originalTrace, _ := json.Marshal(want)
	changed := weapon != nil && weapon.InitData == unsafe.Pointer(&replacement)
	if uint32(result) != row.Result || !bytes.Equal(actualTrace, originalTrace) || final != row.FinalSlots ||
		changed != row.ChangedBank || !slices.Equal(active, row.FinalActive) {
		t.Fatalf("result/trace %d/%s, want %d/%s, slots %v/%v, changed bank %t/%t, callbacks %v/%v",
			uint32(result), actualTrace, row.Result, originalTrace, final, row.FinalSlots, changed, row.ChangedBank, active, row.FinalActive)
	}
}
