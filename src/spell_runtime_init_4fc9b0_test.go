package opennox

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSpellRuntimeInit4FC9B0FailureGates(t *testing.T) {
	tests := []struct {
		name       string
		durations  int32
		allocator  uintptr
		caster     uintptr
		want       int32
		wantEvents []string
	}{
		{
			name:       "duration-failure",
			durations:  0,
			allocator:  0x1111,
			caster:     0x2222,
			want:       0,
			wantEvents: []string{"durations"},
		},
		{
			name:       "magic-class-failure-is-stored",
			durations:  1,
			allocator:  0,
			caster:     0x2222,
			want:       0,
			wantEvents: []string{"durations", "new-magic", "store-magic:0x0"},
		},
		{
			name:       "caster-failure-is-stored",
			durations:  -1,
			allocator:  0x1111,
			caster:     0,
			want:       0,
			wantEvents: []string{"durations", "new-magic", "store-magic:0x1111", "new-object:ImaginaryCaster", "store-caster:0x0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			got := spellRuntimeInit4FC9B0(
				spellRuntimeMagicClassPE32Size4FC9B0,
				spellRuntimeInitHooks4FC9B0[uintptr, uintptr]{
					initDurations: func() int32 {
						events = append(events, "durations")
						return tc.durations
					},
					newMagicClass: func(name string, size uintptr, capacity int) uintptr {
						events = append(events, "new-magic")
						if name != spellRuntimeMagicClassName4FC9B0 || size != 60 || capacity != 64 {
							t.Fatalf("magic class request = (%q, %d, %d)", name, size, capacity)
						}
						return tc.allocator
					},
					storeMagicClass: func(value uintptr) {
						events = append(events, fmt.Sprintf("store-magic:%#x", value))
					},
					newObjectByTypeID: func(name string) uintptr {
						events = append(events, "new-object:"+name)
						return tc.caster
					},
					storeImaginaryCaster: func(value uintptr) {
						events = append(events, fmt.Sprintf("store-caster:%#x", value))
					},
					createObjectAt: func(uintptr, uintptr, float32, float32) {
						t.Fatal("failed gate reached object creation")
					},
					objectTypeIDByName: func(string) uint32 {
						t.Fatal("failed gate reached object-type lookup")
						return 0
					},
					storeSpellObjectTypeID: func(int, uint32) {
						t.Fatal("failed gate reached object-type store")
					},
				},
			)

			if got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %#v, want %#v", events, tc.wantEvents)
			}
		})
	}
}

func TestSpellRuntimeInit4FC9B0SuccessOrderConstantsAndRawIDs(t *testing.T) {
	const (
		allocator = uintptr(0x100001234)
		caster    = uintptr(0x200005678)
	)
	wantIDs := [...]uint32{
		0,
		0xffffffff,
		0x80000000,
		0x7fffffff,
		0x01234567,
		0x89abcdef,
		0xa5a55a5a,
	}
	var events []string
	storedIDs := [len(spellRuntimeObjectTypeNames4FC9B0)]uint32{}

	got := spellRuntimeInit4FC9B0(
		spellRuntimeMagicClassPE32Size4FC9B0,
		spellRuntimeInitHooks4FC9B0[uintptr, uintptr]{
			initDurations: func() int32 {
				events = append(events, "durations")
				return -7
			},
			newMagicClass: func(name string, size uintptr, capacity int) uintptr {
				events = append(events, fmt.Sprintf("new-magic:%s:%d:%d", name, size, capacity))
				return allocator
			},
			storeMagicClass: func(value uintptr) {
				events = append(events, fmt.Sprintf("store-magic:%#x", value))
			},
			newObjectByTypeID: func(name string) uintptr {
				events = append(events, "new-object:"+name)
				return caster
			},
			storeImaginaryCaster: func(value uintptr) {
				events = append(events, fmt.Sprintf("store-caster:%#x", value))
			},
			createObjectAt: func(object, owner uintptr, x, y float32) {
				events = append(events, fmt.Sprintf("create:%#x:%#x:%g:%g", object, owner, x, y))
			},
			objectTypeIDByName: func(name string) uint32 {
				for i, wantName := range spellRuntimeObjectTypeNames4FC9B0 {
					if name == wantName {
						events = append(events, "lookup:"+name)
						return wantIDs[i]
					}
				}
				t.Fatalf("unexpected object type %q", name)
				return 0
			},
			storeSpellObjectTypeID: func(index int, value uint32) {
				events = append(events, fmt.Sprintf("store-id:%d:%#x", index, value))
				storedIDs[index] = value
			},
		},
	)

	if got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if storedIDs != wantIDs {
		t.Fatalf("stored IDs = %#v, want raw %#v", storedIDs, wantIDs)
	}
	wantEvents := []string{
		"durations",
		"new-magic:magicEntityClass:60:64",
		"store-magic:0x100001234",
		"new-object:ImaginaryCaster",
		"store-caster:0x200005678",
		"create:0x200005678:0x0:2944:2944",
	}
	for i, name := range spellRuntimeObjectTypeNames4FC9B0 {
		wantEvents = append(wantEvents, "lookup:"+name)
		wantEvents = append(wantEvents, fmt.Sprintf("store-id:%d:%#x", i, wantIDs[i]))
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}
}

func TestSpellRuntimeInit4FC9B0FaultPrefixes(t *testing.T) {
	allEvents := []string{
		"durations",
		"new-magic",
		"store-magic",
		"new-object",
		"store-caster",
		"create-object",
	}
	for _, name := range spellRuntimeObjectTypeNames4FC9B0 {
		allEvents = append(allEvents, "lookup:"+name, "store:"+name)
	}
	stop := &struct{}{}

	for failAt := range allEvents {
		t.Run(fmt.Sprintf("fault-%02d-%s", failAt, allEvents[failAt]), func(t *testing.T) {
			events := make([]string, 0, len(allEvents))
			observe := func(event string) {
				if len(events) == failAt {
					panic(stop)
				}
				events = append(events, event)
			}
			lookupIndex := 0
			storeIndex := 0
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				spellRuntimeInit4FC9B0(
					80,
					spellRuntimeInitHooks4FC9B0[uintptr, uintptr]{
						initDurations: func() int32 {
							observe("durations")
							return 1
						},
						newMagicClass: func(string, uintptr, int) uintptr {
							observe("new-magic")
							return 0x100000001
						},
						storeMagicClass: func(uintptr) {
							observe("store-magic")
						},
						newObjectByTypeID: func(string) uintptr {
							observe("new-object")
							return 0x200000002
						},
						storeImaginaryCaster: func(uintptr) {
							observe("store-caster")
						},
						createObjectAt: func(uintptr, uintptr, float32, float32) {
							observe("create-object")
						},
						objectTypeIDByName: func(name string) uint32 {
							if name != spellRuntimeObjectTypeNames4FC9B0[lookupIndex] {
								t.Fatalf("lookup %d = %q", lookupIndex, name)
							}
							lookupIndex++
							observe("lookup:" + name)
							return uint32(lookupIndex)
						},
						storeSpellObjectTypeID: func(index int, value uint32) {
							name := spellRuntimeObjectTypeNames4FC9B0[storeIndex]
							if index != storeIndex || value != uint32(storeIndex+1) {
								t.Fatalf("store %d = (%d, %d)", storeIndex, index, value)
							}
							storeIndex++
							observe("store:" + name)
						},
					},
				)
			}()

			if recovered != stop {
				t.Fatalf("recovered = %#v, want sentinel", recovered)
			}
			if want := allEvents[:failAt]; !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %#v, want prefix %#v", events, want)
			}
		})
	}
}
