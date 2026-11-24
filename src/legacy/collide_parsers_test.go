package legacy

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func TestCollideParsersLoadNativeObjectRecords(t *testing.T) {
	srv := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	load := func(name string, args []string) (*server.ObjectType, error) {
		id := "ParserFixture" + name
		err := srv.Types.ReadObjectType(&things.Thing{
			Name: id, OnCollide: &things.ProcFunc{Name: name, Args: args},
		})
		return srv.Types.ByID(id), err
	}
	for _, test := range []struct {
		name string
		args []string
		read func(unsafe.Pointer) any
		want any
	}{
		{"ProjectileCollide", []string{"\t-42tail"}, func(p unsafe.Pointer) any {
			return *(*server.ProjectileCollideData)(p)
		}, server.ProjectileCollideData{Damage: -42}},
		{"WallReflectCollide", []string{"not-decimal"}, func(p unsafe.Pointer) any {
			return *(*int32)(p)
		}, int32(0)},
		{"AudioEventCollide", []string{"\tHasteOn", "ignored"}, func(p unsafe.Pointer) any {
			return *(*server.AudioEventCollideData)(p)
		}, server.AudioEventCollideData{Sound: uint32(sound.SoundHasteOn)}},
		{"SparkExplosionCollide", []string{"300particles"}, func(p unsafe.Pointer) any {
			return *(*server.SparkExplosionCollideData)(p)
		}, server.SparkExplosionCollideData{Power: 44}},
		{"DamageCollide", []string{"  +260tail", "DAMAGE_AIRBORNE_ELECTRIC", "ignored"}, func(p unsafe.Pointer) any {
			return *(*server.DamageCollideData)(p)
		}, server.DamageCollideData{Damage: 4, DamageType: 17}},
		{"ManaDrainCollide", []string{"-3junk", "ignored"}, func(p unsafe.Pointer) any {
			return *(*server.ManaDrainCollideData)(p)
		}, server.ManaDrainCollideData{Amount: 253}},
		{"MonsterArrowCollide", []string{"-1095suffix", "8"}, func(p unsafe.Pointer) any {
			return *(*server.MonsterArrowCollideData)(p)
		}, server.MonsterArrowCollideData{CoopDamage: -1095, OtherDamage: -1095}},
	} {
		t.Run(test.name, func(t *testing.T) {
			obj, err := load(test.name, test.args)
			if err != nil || obj == nil || obj.CollideData == nil {
				t.Fatalf("object %v parse error %v", obj, err)
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(obj.CollideData) <= math.MaxUint32 {
				t.Fatal("native collide data did not exercise a pointer above 4 GiB")
			}
			if got := test.read(obj.CollideData); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("collide record %+v, want %+v", got, test.want)
			}
		})
	}
	for _, args := range [][]string{{"1", "damage_bite"}, {"21"}, {"9", "DAMAGE_TRUE"}} {
		if _, err := load("DamageCollide", args); err == nil || !strings.Contains(err.Error(), "cannot parse collide data") {
			t.Errorf("invalid original damage type %q returned %v", args, err)
		}
	}
}

func TestCollideDamageNamesKeepOriginalIdentifiers(t *testing.T) {
	names := []string{
		"BLADE", "FLAME", "CRUSH", "IMPALE", "DRAIN", "POISON",
		"DISPEL_UNDEAD", "EXPLOSION", "BITE", "ELECTRIC", "CLAW", "IMPACT",
		"LAVA", "DEATH_MAGIC", "PLASMA", "MANA_BOMB", "ZAP_RAY", "AIRBORNE_ELECTRIC",
	}
	for index, name := range names {
		if got := damageTypeNamed4E0A00("DAMAGE_" + name); got != int32(index) {
			t.Errorf("DAMAGE_%s mapped to %d, want %d", name, got, index)
		}
		if got := damageTypeNamed4E0A00(name); got != 18 {
			t.Errorf("unprefixed %s mapped to %d, want the invalid code 18", name, got)
		}
	}
}
