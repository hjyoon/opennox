package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestCastFireballNative52C790KeepsNativeCasterPointer(t *testing.T) {
	caster := &server.Object{
		PosVec:     types.Ptf(100, 200),
		VelVec:     types.Ptf(2, 3),
		Direction1: 64,
	}
	caster.Shape.Circle.R = 4
	projectile := &server.Object{SpeedCur: 10}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(caster)) <= uintptr(math.MaxUint32) {
		t.Fatalf("caster pointer = %p, want native address above 4 GiB", caster)
	}

	var created, audio bool
	got := castFireballNative52C790(spell.SPELL_FIREBALL, caster, 1, fireballCastHooks52C790{
		newObject: func(typeID string) *server.Object {
			if typeID != "Fireball" {
				t.Fatalf("projectile type = %q, want Fireball", typeID)
			}
			return projectile
		},
		traceRay: func(from, to types.Pointf) bool {
			cosine, sine := server.SinCosDir(byte(caster.Direction1))
			want := caster.PosVec.Add(caster.VelVec).Add(types.Ptf(cosine, sine).Mul(8))
			if from != caster.PosVec || to != want {
				t.Fatalf("trace = %v -> %v, want %v -> %v", from, to, caster.PosVec, want)
			}
			return true
		},
		createAt: func(gotProjectile, owner *server.Object, position types.Pointf) {
			created = true
			if gotProjectile != projectile || owner != caster {
				t.Fatalf("create pointers = %p/%p, want %p/%p", gotProjectile, owner, projectile, caster)
			}
			projectile.PosVec = position
		},
		speedCoeff: func(levelIndex int) float64 {
			if levelIndex != 0 {
				t.Fatalf("speed level index = %d, want 0", levelIndex)
			}
			return 1.5
		},
		playCastAudio: func(id spell.ID, gotCaster *server.Object) {
			audio = true
			if id != spell.SPELL_FIREBALL || gotCaster != caster {
				t.Fatalf("audio = %d/%p, want %d/%p", id, gotCaster, spell.SPELL_FIREBALL, caster)
			}
		},
	})
	if got != 1 || !created || !audio {
		t.Fatalf("result/create/audio = %d/%t/%t, want 1/true/true", got, created, audio)
	}
	cosine, sine := server.SinCosDir(byte(caster.Direction1))
	direction := types.Ptf(cosine, sine)
	wantPosition := caster.PosVec.Add(caster.VelVec).Add(direction.Mul(8))
	wantVelocity := caster.VelVec.Add(direction.Mul(15))
	if projectile.PosVec != wantPosition || projectile.SpeedCur != 15 || projectile.VelVec != wantVelocity ||
		projectile.Direction1 != caster.Direction1 || projectile.Direction2 != caster.Direction1 {
		t.Fatalf("projectile = pos:%v speed:%g vel:%v dir:%d/%d", projectile.PosVec,
			projectile.SpeedCur, projectile.VelVec, projectile.Direction1, projectile.Direction2)
	}
}

func TestCastFireballNative52C790ProjectileLevels(t *testing.T) {
	want := []string{"Fireball", "StrongFireball", "TitanFireball", "TitanFireball", "TitanFireball"}
	for level, wantType := range want {
		level++
		t.Run(wantType+string(rune('0'+level)), func(t *testing.T) {
			caster := &server.Object{}
			projectile := &server.Object{SpeedCur: 2}
			var gotType string
			got := castFireballNative52C790(spell.SPELL_FIREBALL, caster, level, fireballCastHooks52C790{
				newObject: func(typeID string) *server.Object { gotType = typeID; return projectile },
				traceRay:  func(types.Pointf, types.Pointf) bool { return false },
				createAt: func(_ *server.Object, owner *server.Object, position types.Pointf) {
					if owner != caster || position != caster.PosVec {
						t.Fatalf("blocked spawn = %p/%v", owner, position)
					}
				},
				speedCoeff: func(levelIndex int) float64 {
					if levelIndex != level-1 {
						t.Fatalf("speed level index = %d, want %d", levelIndex, level-1)
					}
					return float64(level)
				},
				playCastAudio: func(spell.ID, *server.Object) {},
			})
			if got != 1 || gotType != wantType || projectile.SpeedCur != float32(2*level) {
				t.Fatalf("level %d = result:%d type:%q speed:%g", level, got, gotType, projectile.SpeedCur)
			}
		})
	}
}

func TestCastFireballNative52C790MissingInputs(t *testing.T) {
	called := false
	hooks := fireballCastHooks52C790{
		newObject: func(string) *server.Object { called = true; return nil },
	}
	if got := castFireballNative52C790(spell.SPELL_FIREBALL, nil, 1, hooks); got != 0 || called {
		t.Fatalf("nil caster = result:%d side effects:%t, want 0/false", got, called)
	}
	if got := castFireballNative52C790(spell.SPELL_FIREBALL, &server.Object{}, 0, hooks); got != 1 || called {
		t.Fatalf("invalid level = result:%d side effects:%t, want 1/false", got, called)
	}
	if got := castFireballNative52C790(spell.SPELL_FIREBALL, &server.Object{}, 1, hooks); got != 1 || !called {
		t.Fatalf("missing projectile = result:%d new-object:%t, want 1/true", got, called)
	}
}
