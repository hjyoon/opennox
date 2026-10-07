package opennox

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestE2ELockInvariantRejectsAndReportsActualDrift(t *testing.T) {
	for _, kind := range []string{"none", "owner", "expiry", "HP", "mana", "NPC-HP"} {
		t.Run(kind, func(t *testing.T) {
			hostUpdate := &server.PlayerUpdateData{ManaCur: 150}
			host := &server.Object{ObjClass: object.ClassPlayer, HealthData: &server.HealthData{Cur: 75}, UpdateData: unsafe.Pointer(hostUpdate)}
			npc := &server.Object{HealthData: &server.HealthData{Cur: 100}}
			owner, changedOwner := &server.Object{}, &server.Object{}
			outside := &server.Object{ObjOwner: owner, Field34: 512}
			f := &e2eLockFixture{host: host, npc: npc, doors: [3]*server.Object{nil, nil, outside}, outsideOwner: owner, outsideExpiry: 512, health: 75, mana: 150, npcHP: 100}
			switch kind {
			case "owner":
				outside.ObjOwner = changedOwner
			case "expiry":
				outside.Field34++
			case "HP":
				host.HealthData.Cur--
			case "mana":
				hostUpdate.ManaCur++
			case "NPC-HP":
				npc.HealthData.Cur--
			}
			beforeHost, beforeNPC, beforeOutside := *host, *npc, *outside
			beforeHostHP, beforeNPCHealth, beforeUpdate := *host.HealthData, *npc.HealthData, *hostUpdate
			var caught any
			func() {
				defer func() { caught = recover() }()
				f.unchangedOutsideAndUnits()
			}()
			if kind == "none" {
				if caught != nil {
					t.Fatalf("unchanged live state was rejected: %v", caught)
				}
			} else {
				if caught == nil {
					t.Fatal("original drift assertion no longer rejects")
				}
				want := fmt.Sprintf("outside=%p owner=%p want=%p expiry=%d want=%d HP=%d want=%d mana=%d want=%d NPC-HP=%d want=%d",
					outside, outside.ObjOwner, owner, outside.Field34, f.outsideExpiry, host.HealthData.Cur, f.health, hostUpdate.ManaCur, f.mana, npc.HealthData.Cur, f.npcHP)
				if !strings.Contains(fmt.Sprint(caught), want) {
					t.Fatalf("actual invariant values absent: got=%v want=%s", caught, want)
				}
			}
			if !reflect.DeepEqual(*host, beforeHost) || !reflect.DeepEqual(*npc, beforeNPC) || !reflect.DeepEqual(*outside, beforeOutside) ||
				*host.HealthData != beforeHostHP || *npc.HealthData != beforeNPCHealth || !reflect.DeepEqual(*hostUpdate, beforeUpdate) {
				t.Fatal("observation changed live invariant data")
			}
		})
	}
}
