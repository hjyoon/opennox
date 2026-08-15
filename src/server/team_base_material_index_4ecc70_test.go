package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTeamBaseMaterialIndex4ECC70CachedMatchOrder(t *testing.T) {
	var events []string
	got := teamBaseMaterialIndex4ECC70(10, teamBaseMaterialIndexHooks4ECC70[int, int, int, uint32]{
		loadCachedType: func() uint32 {
			events = append(events, "cache")
			return 7
		},
		lookupType: func(string) uint32 {
			t.Fatal("nonzero cache performed type lookup")
			return 0
		},
		storeCachedType: func(uint32) {
			t.Fatal("nonzero cache was stored again")
		},
		loadTypeIndex: func(obj int) uint16 {
			events = append(events, "type")
			if obj != 10 {
				t.Fatalf("object = %d, want 10", obj)
			}
			return 7
		},
		loadInitData: func(obj int) int {
			events = append(events, "init")
			return obj + 10
		},
		loadSecondModifier: func(data int) int {
			events = append(events, "modifier")
			if data != 20 {
				t.Fatalf("init-data = %d, want 20", data)
			}
			return 30
		},
		lookupMaterial: func(modifier int) uint32 {
			events = append(events, "material")
			if modifier != 30 {
				t.Fatalf("modifier = %d, want 30", modifier)
			}
			return 9
		},
	})
	if got != 9 {
		t.Fatalf("result = %d, want 9", got)
	}
	wantEvents := []string{"cache", "type", "init", "modifier", "material"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestTeamBaseMaterialIndex4ECC70LookupStoreAndLiveResultOrder(t *testing.T) {
	var events []string
	cache := uint32(0)
	got := teamBaseMaterialIndex4ECC70(1, teamBaseMaterialIndexHooks4ECC70[int, int, int, uint32]{
		loadCachedType: func() uint32 {
			events = append(events, "cache")
			return cache
		},
		lookupType: func(name string) uint32 {
			events = append(events, "lookup:"+name)
			if name != teamBaseTypeName4ECC70 {
				t.Fatalf("lookup name = %q, want %q", name, teamBaseTypeName4ECC70)
			}
			return 7
		},
		storeCachedType: func(value uint32) {
			events = append(events, fmt.Sprintf("store:%d", value))
			if value != 7 {
				t.Fatalf("stored type = %d, want 7", value)
			}
			// The original compares the live lookup result in EAX and does not
			// reload the BSS cell after this store.
			cache = 99
		},
		loadTypeIndex: func(int) uint16 {
			events = append(events, "type")
			return 7
		},
		loadInitData: func(int) int {
			events = append(events, "init")
			return 2
		},
		loadSecondModifier: func(int) int {
			events = append(events, "modifier")
			return 3
		},
		lookupMaterial: func(int) uint32 {
			events = append(events, "material")
			return 5
		},
	})
	if got != 5 {
		t.Fatalf("result = %d, want 5", got)
	}
	wantEvents := []string{"cache", "lookup:TeamBase", "store:7", "type", "init", "modifier", "material"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestTeamBaseMaterialIndex4ECC70ZeroLookupCanMatchZeroType(t *testing.T) {
	var events []string
	got := teamBaseMaterialIndex4ECC70(1, teamBaseMaterialIndexHooks4ECC70[int, int, int, uint32]{
		loadCachedType: func() uint32 {
			events = append(events, "cache")
			return 0
		},
		lookupType: func(string) uint32 {
			events = append(events, "lookup")
			return 0
		},
		storeCachedType: func(value uint32) {
			events = append(events, "store")
			if value != 0 {
				t.Fatalf("stored type = %d, want 0", value)
			}
		},
		loadTypeIndex: func(int) uint16 {
			events = append(events, "type")
			return 0
		},
		loadInitData: func(int) int {
			events = append(events, "init")
			return 2
		},
		loadSecondModifier: func(int) int {
			events = append(events, "modifier")
			return 0
		},
		lookupMaterial: func(modifier int) uint32 {
			events = append(events, "material")
			if modifier != 0 {
				t.Fatalf("modifier = %d, want nil token 0", modifier)
			}
			return 4
		},
	})
	if got != 4 {
		t.Fatalf("result = %d, want 4", got)
	}
	wantEvents := []string{"cache", "lookup", "store", "type", "init", "modifier", "material"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestTeamBaseMaterialIndex4ECC70MismatchZeroExtendsTypeAndSkipsInit(t *testing.T) {
	var events []string
	got := teamBaseMaterialIndex4ECC70(1, teamBaseMaterialIndexHooks4ECC70[int, int, int, uint32]{
		loadCachedType: func() uint32 {
			events = append(events, "cache")
			return ^uint32(0)
		},
		lookupType:      func(string) uint32 { t.Fatal("unexpected lookup"); return 0 },
		storeCachedType: func(uint32) { t.Fatal("unexpected store") },
		loadTypeIndex: func(int) uint16 {
			events = append(events, "type")
			return ^uint16(0)
		},
		loadInitData: func(int) int { t.Fatal("mismatch loaded init-data"); return 0 },
		loadSecondModifier: func(int) int {
			t.Fatal("mismatch loaded modifier")
			return 0
		},
		lookupMaterial: func(int) uint32 { t.Fatal("mismatch performed material lookup"); return 0 },
	})
	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	wantEvents := []string{"cache", "type"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestTeamBaseMaterialIndex4ECC70ZeroCacheRetriesLookup(t *testing.T) {
	cache := uint32(0)
	lookups := 0
	hooks := teamBaseMaterialIndexHooks4ECC70[int, int, int, uint32]{
		loadCachedType: func() uint32 { return cache },
		lookupType: func(string) uint32 {
			lookups++
			return 0
		},
		storeCachedType: func(value uint32) { cache = value },
		loadTypeIndex:   func(int) uint16 { return 1 },
		loadInitData:    func(int) int { t.Fatal("mismatch loaded init-data"); return 0 },
		loadSecondModifier: func(int) int {
			t.Fatal("mismatch loaded modifier")
			return 0
		},
		lookupMaterial: func(int) uint32 { t.Fatal("mismatch performed material lookup"); return 0 },
	}
	if got := teamBaseMaterialIndex4ECC70(1, hooks); got != 0 {
		t.Fatalf("first result = %d, want 0", got)
	}
	if got := teamBaseMaterialIndex4ECC70(1, hooks); got != 0 {
		t.Fatalf("second result = %d, want 0", got)
	}
	if lookups != 2 {
		t.Fatalf("lookup count = %d, want 2", lookups)
	}
}

func TestTeamBaseMaterialIndex4ECC70FaultOrder(t *testing.T) {
	const fault = "ordered fault"
	stages := []string{"cache", "lookup", "store", "type", "init", "modifier", "material"}
	for _, fail := range stages {
		t.Run(fail, func(t *testing.T) {
			var events []string
			step := func(name string) {
				events = append(events, name)
				if name == fail {
					panic(fault)
				}
			}
			defer func() {
				if got := recover(); got != fault {
					t.Fatalf("panic = %v, want %q", got, fault)
				}
				want := make([]string, 0, len(stages))
				for _, stage := range stages {
					want = append(want, stage)
					if stage == fail {
						break
					}
				}
				if !reflect.DeepEqual(events, want) {
					t.Fatalf("events = %v, want %v", events, want)
				}
			}()
			teamBaseMaterialIndex4ECC70(1, teamBaseMaterialIndexHooks4ECC70[int, int, int, uint32]{
				loadCachedType: func() uint32 { step("cache"); return 0 },
				lookupType: func(string) uint32 {
					step("lookup")
					return 7
				},
				storeCachedType: func(uint32) { step("store") },
				loadTypeIndex: func(int) uint16 {
					step("type")
					return 7
				},
				loadInitData: func(int) int { step("init"); return 2 },
				loadSecondModifier: func(int) int {
					step("modifier")
					return 3
				},
				lookupMaterial: func(int) uint32 { step("material"); return 4 },
			})
		})
	}
}
