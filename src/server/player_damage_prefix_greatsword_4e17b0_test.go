package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerDamagePiercePrefixGreatSword4E17B0(t *testing.T) {
	for _, livePlayer := range []bool{false, true} {
		t.Run(map[bool]string{false: "live player nil", true: "live player still observing"}[livePlayer], func(t *testing.T) {
			target, source, weapon, missile, _, cached, r, events := damageGreatSwordFixture4E17B0(t, true, false)
			entry := target.UpdateDataPlayer()
			entry.Player.Field3680, entry.Player.CameraFollowObj = 2, source
			entry.Field76, entry.Field75 = 1, uint32(missile.TypeInd)
			beforeEntry := *entry
			live := &PlayerUpdateData{State: PlayerState16, Field76: 31, Field75: 33}
			if livePlayer {
				live.Player = &Player{Field3680: 2, CameraFollowObj: missile}
			}
			target.UpdateData = unsafe.Pointer(live)
			r.playerPrefix = &playerDamagePrefix4E17B0{update: entry, weaponFlags: 0x400}
			r.ObserveClear = func(*Object) { t.Fatal("GreatSword repeated ObserveClear") }
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) }
			r.BlockSourceExcluded = func(w *Object) bool {
				if w != missile || *entry != beforeEntry {
					t.Fatal("GreatSword lost prefix before exclusion")
				}
				missile.TypeInd++
				return false
			}
			r.ProjectileReflect = func(w, v *Object) {
				if w != missile || v != target || *entry != beforeEntry || live.Field76 != 31 || live.Field75 != 33 {
					t.Fatal("GreatSword repeated marker stores or read live state")
				}
				*events = append(*events, "reflect")
			}
			applicable, handled, result := playerDamageGreatSwordMissileBlock4E17B0(target, source, weapon, cached, 9, object.DamageImpale, r)
			want := []string{"reflect", "clear-owner", "set-owner", "audio", "rng", "state-19", "balance", "wear"}
			if !applicable || !handled || result || !reflect.DeepEqual(*events, want) || *entry != beforeEntry || live.State != PlayerState19 || live.Field76 != 31 || live.Field75 != 33 || target.HealthData.Cur != 200 {
				t.Fatalf("GreatSword=%t/%t/%t events=%v cached/live state=%d/%d", applicable, handled, result, *events, entry.State, live.State)
			}
		})
	}
}
