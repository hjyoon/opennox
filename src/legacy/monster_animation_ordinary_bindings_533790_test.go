package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func TestAnimationSelectorOrdinaryBindings533790(t *testing.T) {
	path := os.Getenv("NOX_ANIMATION_ORIGINAL_REFERENCE")
	if path == "" {
		path = "../server/testdata/monster_animation_selector_original_533790.json.gz"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != "1138dde79ead136120908feb14740281833366b7ef95bd30b4ada4b73e1298a9" {
		t.Fatal("unsealed original binding reference")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Rows []animationRow533790 `json:"rows"`
	}
	if err = json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if len(ref.Rows) != 14840 {
		t.Fatal("original binding dimensions")
	}
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	for index := 1; index <= 44; index++ {
		name := fmt.Sprintf("AnimationBindingType%d", index)
		for i, v := range []int{11, 22, 33, 44} {
			if v == index {
				name = animationNames533790[i]
			}
		}
		if err := srv.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(srv.FreeObjectTypes)
	if srv.Types.MimicID() != 11 {
		t.Fatal("ordinary Go Mimic type cache")
	}
	view := srv.TypePredicateRuntime534A10()
	view.Write(0, 22)
	view.Write(1, 33)
	view.Write(2, 44)
	oldCache := statePredicateOriginalSwapMimicCache534840(11)
	t.Cleanup(func() { statePredicateOriginalSwapMimicCache534840(oldCache) })
	unit, freeU := alloc.New(server.Object{})
	update, freeD := alloc.New(server.MonsterUpdateData{})
	enemy, freeE := alloc.New(server.Object{})
	t.Cleanup(freeU)
	t.Cleanup(freeD)
	t.Cleanup(freeE)
	*enemy = server.Object{}
	cases := []struct {
		Name           string
		Head           int8
		Action, Status uint32
		Type           uint16
		Enemy          bool
		Class          uint32
		NoServer       bool
	}{
		{"cached Mimic idle without server", 0, 0, 0, 11, false, 0, true},
		{"cached morphed Mimic without server", 0, 0, 0xc4000, 11, false, 0, true},
		{"ordinary plant idle", -1, 0, 0, 22, false, 2, false},
		{"ordinary plant enemy", -1, 0, 0, 22, true, 2, false},
		{"ordinary zombie dying", 0, 30, 0, 33, false, 2, false},
		{"ordinary burning zombie", 0, 30, 0x80000, 33, false, 2, false},
		{"ordinary burning vile zombie", 0, 30, 0x80000, 44, false, 2, false},
		{"ordinary running movement", 7, 7, 0x4000, 117, false, 2, false},
		{"ordinary final-slot missile", 23, 17, 0, 117, false, 2, false},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var original *animationRow533790
			for i := range ref.Rows {
				r := &ref.Rows[i]
				x := r.Input
				if x.Head == c.Head && x.Action == c.Action && x.Status == c.Status && x.Type == c.Type && x.Enemy == c.Enemy && x.Class == c.Class && x.Caches == [4]uint32{11, 22, 33, 44} && x.Repeat == 1 && x.Mutation == nil && x.Found == [4]bool{true, true, true, true} {
					original = r
					break
				}
			}
			if original == nil || len(original.Output.Returns) != 1 {
				t.Fatal("binding input missing from frozen original")
			}
			*update = server.MonsterUpdateData{AIStackInd: c.Head, StatusFlags: object.MonsterStatus(c.Status)}
			if c.Head >= 0 {
				update.AIStack[c.Head].Action = c.Action
			}
			if c.Enemy {
				update.CurrentEnemy = enemy
			}
			*unit = server.Object{TypeInd: c.Type, ObjClass: object.Class(c.Class), UpdateData: unsafe.Pointer(update)}
			oldServer := GetServer
			if c.NoServer {
				GetServer = func() Server { panic("cached Mimic idle unexpectedly resolved server or remaining predicate") }
			}
			defer func() { GetServer = oldServer }()
			beforeU, beforeD := *unit, *update
			want := original.Output.Returns[0]
			if got := uint32(ccall.CallIntPtr(animationSelectorPublicC533790(), unit.CObj())); got != want {
				t.Errorf("ordinary public C=%d original=%d", got, want)
			}
			if got := uint32(srv.MonsterAnimationIndex533790(unit)); got != want {
				t.Errorf("ordinary public Go=%d original=%d", got, want)
			}
			if *unit != beforeU || *update != beforeD || view.Read(0) != 22 || view.Read(1) != 33 || view.Read(2) != 44 || *(*uint32)(statePredicateOriginalMimicCachePointer534840()) != 11 {
				t.Fatal("ordinary cached binding modified native inputs or caches")
			}
		})
	}
	t.Log("ordinary animation bindings modes=9 actual-C-calls=9 public-Go-calls=9 original-words=true cached-Mimic-no-server=true no-provider-or-predicate-substitution=true")
}
