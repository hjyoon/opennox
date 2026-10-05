package opennox

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestE2EPlayerFlameMetadata(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*server.Object, *server.PlayerUpdateData)
		want   bool
	}{
		{"raw damage type", func(*server.Object, *server.PlayerUpdateData) {}, true},
		{"float-encoded type rejected", func(_ *server.Object, u *server.PlayerUpdateData) {
			u.Field75 = math.Float32bits(float32(object.DamageFlame))
		}, false},
		{"wrong marker state", func(_ *server.Object, u *server.PlayerUpdateData) { u.Field76 = 1 }, false},
		{"wrong attribution", func(p *server.Object, _ *server.PlayerUpdateData) { p.Obj130 = &server.Object{} }, false},
		{"wrong damage type", func(p *server.Object, _ *server.PlayerUpdateData) { p.Field131 = uint32(object.DamageImpale) }, false},
		{"wrong hit position", func(p *server.Object, _ *server.PlayerUpdateData) { p.Pos132 = types.Ptf(1, 2) }, false},
		{"nil update", func(p *server.Object, _ *server.PlayerUpdateData) { p.UpdateData = nil }, false},
		{"non-player layout", func(p *server.Object, _ *server.PlayerUpdateData) { p.ObjClass = object.ClassMonster }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			flame := &server.Object{ObjClass: object.ClassFire}
			update := &server.PlayerUpdateData{Field75: uint32(object.DamageFlame), Field76: 2}
			player := &server.Object{
				ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update),
				Obj130: flame, Field131: uint32(object.DamageFlame),
				HealthData: &server.HealthData{Cur: 147, Max: 150},
			}
			tc.mutate(player, update)
			beforePlayer, beforeUpdate, beforeFlame, beforeHP := *player, *update, *flame, *player.HealthData
			if got := e2ePlayerFlameMetadata(player, flame); got != tc.want {
				t.Fatalf("metadata=%t want=%t", got, tc.want)
			}
			if !reflect.DeepEqual(*player, beforePlayer) || !reflect.DeepEqual(*update, beforeUpdate) ||
				!reflect.DeepEqual(*flame, beforeFlame) || *player.HealthData != beforeHP {
				t.Fatal("read-only metadata observer changed a record")
			}
		})
	}
	for _, tc := range []struct {
		name          string
		player, flame *server.Object
	}{
		{"nil player", nil, &server.Object{}},
		{"nil flame", &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&server.PlayerUpdateData{})}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if e2ePlayerFlameMetadata(tc.player, tc.flame) {
				t.Fatal("nil metadata record accepted")
			}
		})
	}
}
