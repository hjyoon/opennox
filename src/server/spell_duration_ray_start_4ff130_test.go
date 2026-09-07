package server

import (
	"fmt"
	"reflect"
	"testing"
)

func (w *durationRayWorld4FEF90) startHooks4FF130() DurationRayStartHooks4FF130[uint64, uint64] {
	return DurationRayStartHooks4FF130[uint64, uint64]{
		LoadSpell: func(record uint64) uint32 {
			value := w.records[record].spell
			w.observe(fmt.Sprintf("spell:%s=%d", durationRayRecordName4FEF90(record), value))
			return value
		},
		LoadLevelLow: func(record uint64) byte {
			value := byte(w.records[record].level)
			w.observe(fmt.Sprintf("level:%s=%02x", durationRayRecordName4FEF90(record), value))
			return value
		},
		LoadCaster: func(record uint64) uint64 {
			value := w.records[record].caster
			w.observe(fmt.Sprintf("caster:%s=%s", durationRayRecordName4FEF90(record), durationRayObjectName4FEF90(value)))
			return value
		},
		LoadDirectionLow: func(object uint64) byte {
			value := byte(w.objects[object].direction)
			w.observe(fmt.Sprintf("direction:%s=%02x", durationRayObjectName4FEF90(object), value))
			return value
		},
		LoadTarget: func(record uint64) uint64 {
			value := w.records[record].target
			w.observe(fmt.Sprintf("target:%s=%s", durationRayRecordName4FEF90(record), durationRayObjectName4FEF90(value)))
			return value
		},
		LoadSub108: func(record uint64) uint64 {
			value := w.records[record].sub108
			w.observe(fmt.Sprintf("sub108:%s=%s", durationRayRecordName4FEF90(record), durationRayRecordName4FEF90(value)))
			return value
		},
		LoadNext: func(record uint64) uint64 {
			value := w.records[record].next
			w.observe(fmt.Sprintf("next:%s=%s", durationRayRecordName4FEF90(record), durationRayRecordName4FEF90(value)))
			return value
		},
		UnitCode: func(object uint64) uint32 {
			value := w.objects[object].code
			w.observe(fmt.Sprintf("code:%s=%08x", durationRayObjectName4FEF90(object), value))
			return value
		},
		SendPacket: func(recipient int32, packet [7]byte, related uint64, remove int32) int32 {
			w.packets = append(w.packets, durationRayPacket4FEF90{recipient, packet, related, remove})
			w.observe(fmt.Sprintf("send:%d:%x:%s:%d", recipient, packet, durationRayObjectName4FEF90(related), remove))
			return w.sendResult
		},
		UnmarkMinimap: func(object uint64, flags uint32) {
			event := fmt.Sprintf("unmark:%s:%d", durationRayObjectName4FEF90(object), flags)
			w.unmarks = append(w.unmarks, event)
			w.observe(event)
		},
	}
}

func TestDurationRayStart4FF130DispatchAndPacket(t *testing.T) {
	tests := []struct {
		spell  uint32
		packet [7]byte
	}{
		{7, [7]byte{0x9e, 3, 0xab, 0x01, 0xef, 0x21, 0x43}},
		{9, [7]byte{0x9e, 2, 0xab, 0x01, 0xef, 0x21, 0x43}},
		{22, [7]byte{0x9e, 5, 0xab, 0x01, 0xef, 0x21, 0x43}},
		{24, [7]byte{0x9e, 4, 0xab, 0x01, 0xef, 0x21, 0x43}},
		{35, [7]byte{0x9e, 6, 0xab, 0x21, 0x43, 0x01, 0xef}},
		{59, [7]byte{0x9e, 1, 0xcd, 0x01, 0xef, 0x21, 0x43}},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("spell-%d", tc.spell), func(t *testing.T) {
			w := newDurationRayWorld4FEF90(tc.spell)
			w.sendResult = -123456789
			DurationRayStart4FF130(durationRayRecordA4FEF90, w.startHooks4FF130())
			wantPacket := durationRayPacket4FEF90{
				recipient: 255,
				packet:    tc.packet,
				remove:    1,
			}
			if !reflect.DeepEqual(w.packets, []durationRayPacket4FEF90{wantPacket}) {
				t.Fatalf("packets = %#v, want %#v", w.packets, []durationRayPacket4FEF90{wantPacket})
			}
			wantUnmarks := []string{"unmark:caster:2", "unmark:target:2"}
			if !reflect.DeepEqual(w.unmarks, wantUnmarks) {
				t.Fatalf("unmarks = %v, want %v", w.unmarks, wantUnmarks)
			}
		})
	}
}

func TestDurationRayStart4FF130FirstAccessDefaultsAndMissingValues(t *testing.T) {
	t.Run("zero-record-is-not-guarded", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(9)
		w.records[0] = &durationRayRecordState4FEF90{
			spell: 9, level: 0x7a, caster: durationRayCaster4FEF90, target: durationRayTarget4FEF90,
		}
		DurationRayStart4FF130(uint64(0), w.startHooks4FF130())
		if len(w.events) == 0 || w.events[0] != "spell:nil=9" {
			t.Fatalf("events = %v, want Spell as the first unguarded record access", w.events)
		}
	})

	for spellID := uint32(0); spellID <= 70; spellID++ {
		if spellID == 7 || spellID == 9 || spellID == 22 || spellID == 24 ||
			spellID == 35 || spellID == 43 || spellID == 59 {
			continue
		}
		w := newDurationRayWorld4FEF90(spellID)
		DurationRayStart4FF130(durationRayRecordA4FEF90, w.startHooks4FF130())
		want := []string{fmt.Sprintf("spell:A=%d", spellID)}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("spell %d events = %v, want %v", spellID, w.events, want)
		}
	}

	t.Run("ordinary-nil-target", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(24)
		w.records[durationRayRecordA4FEF90].target = 0
		DurationRayStart4FF130(durationRayRecordA4FEF90, w.startHooks4FF130())
		want := []string{"spell:A=24", "level:A=ab", "target:A=nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	t.Run("ordinary-nil-caster-is-not-guarded", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(24)
		w.records[durationRayRecordA4FEF90].caster = 0
		w.objects[0] = &durationRayObjectState4FEF90{code: 0x11223344}
		DurationRayStart4FF130(durationRayRecordA4FEF90, w.startHooks4FF130())
		wantPacket := [7]byte{0x9e, 4, 0xab, 0x44, 0x33, 0x21, 0x43}
		if len(w.packets) != 1 || w.packets[0].packet != wantPacket {
			t.Fatalf("packets = %#v, want %x", w.packets, wantPacket)
		}
		wantPrefix := []string{
			"spell:A=24", "level:A=ab", "target:A=target",
			"code:target=87654321", "caster:A=nil", "code:nil=11223344",
		}
		if !reflect.DeepEqual(w.events[:len(wantPrefix)], wantPrefix) {
			t.Fatalf("events = %v, want prefix %v", w.events, wantPrefix)
		}
	})

	t.Run("plasma-nil-caster-direction-is-not-guarded", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(59)
		w.records[durationRayRecordA4FEF90].caster = 0
		w.objects[0] = &durationRayObjectState4FEF90{direction: 0x77, code: 0x11223344}
		DurationRayStart4FF130(durationRayRecordA4FEF90, w.startHooks4FF130())
		wantPrefix := []string{
			"spell:A=59", "caster:A=nil", "direction:nil=77", "target:A=target",
		}
		if !reflect.DeepEqual(w.events[:len(wantPrefix)], wantPrefix) {
			t.Fatalf("events = %v, want prefix %v", w.events, wantPrefix)
		}
	})

	t.Run("greater-heal-both-nil-equality", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(35)
		w.records[durationRayRecordA4FEF90].target = 0
		w.records[durationRayRecordA4FEF90].caster = 0
		DurationRayStart4FF130(durationRayRecordA4FEF90, w.startHooks4FF130())
		want := []string{"spell:A=35", "target:A=nil", "caster:A=nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	t.Run("greater-heal-nil-target-is-not-guarded", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(35)
		w.records[durationRayRecordA4FEF90].target = 0
		w.objects[0] = &durationRayObjectState4FEF90{code: 0x11223344}
		DurationRayStart4FF130(durationRayRecordA4FEF90, w.startHooks4FF130())
		wantPrefix := []string{
			"spell:A=35", "target:A=nil", "caster:A=caster", "code:nil=11223344",
		}
		if !reflect.DeepEqual(w.events[:len(wantPrefix)], wantPrefix) {
			t.Fatalf("events = %v, want prefix %v", w.events, wantPrefix)
		}
	})
}

func TestDurationRayStart4FF130OrdinaryOrderAndLiveReloads(t *testing.T) {
	w := newDurationRayWorld4FEF90(24)
	w.after["level:A=ab"] = func() {
		w.records[durationRayRecordA4FEF90].target = durationRayTargetA4FEF90
	}
	w.after["code:targetA=aaaa1001"] = func() {
		w.records[durationRayRecordA4FEF90].caster = durationRayCaster24FEF90
	}
	w.after["send:255:9e04ab22220110:nil:1"] = func() {
		w.records[durationRayRecordA4FEF90].caster = durationRayCaster34FEF90
		w.records[durationRayRecordA4FEF90].target = durationRayTargetB4FEF90
	}
	DurationRayStart4FF130(durationRayRecordA4FEF90, w.startHooks4FF130())
	want := []string{
		"spell:A=24",
		"level:A=ab",
		"target:A=targetA",
		"code:targetA=aaaa1001",
		"caster:A=caster2",
		"code:caster2=11112222",
		"send:255:9e04ab22220110:nil:1",
		"caster:A=caster3",
		"unmark:caster3:2",
		"target:A=targetB",
		"unmark:targetB:2",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%v\nwant\n%v", w.events, want)
	}
}

func TestDurationRayStart4FF130CachesPlasmaCasterOnlyForDirection(t *testing.T) {
	w := newDurationRayWorld4FEF90(59)
	w.after["caster:A=caster"] = func() {
		w.records[durationRayRecordA4FEF90].caster = durationRayCaster24FEF90
	}
	w.after["direction:caster=cd"] = func() {
		w.records[durationRayRecordA4FEF90].target = durationRayTargetA4FEF90
	}
	w.after["code:targetA=aaaa1001"] = func() {
		w.records[durationRayRecordA4FEF90].caster = durationRayCaster34FEF90
	}
	w.after["send:255:9e01cd44440110:nil:1"] = func() {
		w.records[durationRayRecordA4FEF90].caster = durationRayCaster44FEF90
		w.records[durationRayRecordA4FEF90].target = durationRayTargetB4FEF90
	}
	DurationRayStart4FF130(durationRayRecordA4FEF90, w.startHooks4FF130())
	want := []string{
		"spell:A=59",
		"caster:A=caster",
		"direction:caster=cd",
		"target:A=targetA",
		"code:targetA=aaaa1001",
		"caster:A=caster3",
		"code:caster3=33334444",
		"send:255:9e01cd44440110:nil:1",
		"caster:A=caster4",
		"unmark:caster4:2",
		"target:A=targetB",
		"unmark:targetB:2",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%v\nwant\n%v", w.events, want)
	}
}

func TestDurationRayStart4FF130GreaterHealOrderAndFullIdentity(t *testing.T) {
	t.Run("same-full-identity", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(35)
		w.records[durationRayRecordA4FEF90].target = durationRayCaster4FEF90
		DurationRayStart4FF130(durationRayRecordA4FEF90, w.startHooks4FF130())
		want := []string{"spell:A=35", "target:A=caster", "caster:A=caster"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	t.Run("same-low-dword-is-distinct", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(35)
		w.records[durationRayRecordA4FEF90].target = durationRayCaster24FEF90
		DurationRayStart4FF130(durationRayRecordA4FEF90, w.startHooks4FF130())
		wantPacket := [7]byte{0x9e, 6, 0xab, 0x22, 0x22, 0x01, 0xef}
		if len(w.packets) != 1 || w.packets[0].packet != wantPacket {
			t.Fatalf("packets = %#v, want %x", w.packets, wantPacket)
		}
	})

	t.Run("callback-order-and-live-reloads", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(35)
		w.after["target:A=target"] = func() {
			w.records[durationRayRecordA4FEF90].caster = durationRayCaster24FEF90
		}
		w.after["code:target=87654321"] = func() {
			w.records[durationRayRecordA4FEF90].caster = durationRayCaster34FEF90
		}
		w.after["code:caster3=33334444"] = func() {
			w.records[durationRayRecordA4FEF90].level = 0xfeedface
		}
		w.after["send:255:9e06ce21434444:nil:1"] = func() {
			w.records[durationRayRecordA4FEF90].caster = durationRayCaster44FEF90
			w.records[durationRayRecordA4FEF90].target = durationRayTargetB4FEF90
		}
		DurationRayStart4FF130(durationRayRecordA4FEF90, w.startHooks4FF130())
		want := []string{
			"spell:A=35",
			"target:A=target",
			"caster:A=caster2",
			"code:target=87654321",
			"caster:A=caster3",
			"code:caster3=33334444",
			"level:A=ce",
			"send:255:9e06ce21434444:nil:1",
			"caster:A=caster4",
			"unmark:caster4:2",
			"target:A=targetB",
			"unmark:targetB:2",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events =\n%v\nwant\n%v", w.events, want)
		}
	})
}

func TestDurationRayStart4FF130ChainLightningUsesLiveNextAfterRecursion(t *testing.T) {
	w := newDurationRayWorld4FEF90(43)
	w.records = map[uint64]*durationRayRecordState4FEF90{
		durationRayRecordParent4FEF90: {
			spell: 43, caster: durationRayCaster4FEF90, sub108: durationRayRecordA4FEF90,
		},
		durationRayRecordA4FEF90: {
			spell: 7, level: 1, caster: durationRayCaster24FEF90,
			target: durationRayTargetA4FEF90, next: durationRayRecordB4FEF90,
		},
		durationRayRecordB4FEF90: {
			spell: 9, level: 2, caster: durationRayCaster34FEF90,
			target: durationRayTargetB4FEF90,
		},
		durationRayRecordC4FEF90: {
			spell: 59, caster: durationRayCaster44FEF90,
			target: durationRayTargetC4FEF90,
		},
	}
	w.after["send:255:9e030122220110:nil:1"] = func() {
		w.records[durationRayRecordA4FEF90].next = durationRayRecordC4FEF90
	}
	DurationRayStart4FF130(durationRayRecordParent4FEF90, w.startHooks4FF130())

	wantPackets := []durationRayPacket4FEF90{
		{255, [7]byte{0x9e, 3, 1, 0x22, 0x22, 0x01, 0x10}, 0, 1},
		{255, [7]byte{0x9e, 1, 0x77, 0x66, 0x66, 0x03, 0x30}, 0, 1},
	}
	if !reflect.DeepEqual(w.packets, wantPackets) {
		t.Fatalf("packets = %#v, want %#v", w.packets, wantPackets)
	}
	for _, event := range w.events {
		if event == "spell:B=9" {
			t.Fatalf("saved Next used instead of live post-recursion link: %v", w.events)
		}
	}
	wantOrder := []string{
		"spell:P=43", "sub108:P=A", "spell:A=7",
		"unmark:targetA:2", "next:A=C", "spell:C=59",
	}
	start := 0
	for _, want := range wantOrder {
		found := false
		for start < len(w.events) {
			if w.events[start] == want {
				found = true
				start++
				break
			}
			start++
		}
		if !found {
			t.Fatalf("event %q not found in order: %v", want, w.events)
		}
	}
}

func TestDurationRayStart4FF130FaultPrefixes(t *testing.T) {
	tests := []struct {
		name  string
		spell uint32
		want  []string
	}{
		{
			name:  "ordinary",
			spell: 24,
			want: []string{
				"spell:A=24", "level:A=ab", "target:A=target",
				"code:target=87654321", "caster:A=caster", "code:caster=abcdef01",
				"send:255:9e04ab01ef2143:nil:1", "caster:A=caster",
				"unmark:caster:2", "target:A=target", "unmark:target:2",
			},
		},
		{
			name:  "greater-heal",
			spell: 35,
			want: []string{
				"spell:A=35", "target:A=target", "caster:A=caster",
				"code:target=87654321", "caster:A=caster", "code:caster=abcdef01",
				"level:A=ab", "send:255:9e06ab214301ef:nil:1", "caster:A=caster",
				"unmark:caster:2", "target:A=target", "unmark:target:2",
			},
		},
		{
			name:  "plasma",
			spell: 59,
			want: []string{
				"spell:A=59", "caster:A=caster", "direction:caster=cd",
				"target:A=target", "code:target=87654321", "caster:A=caster",
				"code:caster=abcdef01", "send:255:9e01cd01ef2143:nil:1", "caster:A=caster",
				"unmark:caster:2", "target:A=target", "unmark:target:2",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for faultAt := 1; faultAt <= len(tc.want); faultAt++ {
				w := newDurationRayWorld4FEF90(tc.spell)
				w.faultAt = faultAt
				panicked := false
				func() {
					defer func() {
						panicked = recover() != nil
					}()
					DurationRayStart4FF130(durationRayRecordA4FEF90, w.startHooks4FF130())
				}()
				if !panicked {
					t.Fatalf("fault %d did not panic", faultAt)
				}
				if want := tc.want[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("fault %d events = %v, want %v", faultAt, w.events, want)
				}
			}
		})
	}
}
