package legacy

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestItemEnchantmentIcon413420LoadStoreAndLookupOrder(t *testing.T) {
	flags := [6]byte{8, 16, 1, 4, 2, 32}
	for _, initial := range []uint32{0, 1, 2, 0xffffffff} {
		for query := 0; query < 256; query++ {
			loaded := initial
			var markers [6]byte
			var cached [6]unsafe.Pointer
			for i := range cached {
				cached[i] = unsafe.Pointer(&markers[i])
			}
			var trace []int
			h := itemEnchantmentIconHooks413420{
				loadLoaded: func() uint32 { trace = append(trace, 100); return loaded },
				loadName:   func(i int) *byte { trace = append(trace, 200+i); return &markers[i] },
				loadImage: func(name *byte) unsafe.Pointer {
					for i := range markers {
						if name == &markers[i] {
							trace = append(trace, 300+i)
							return unsafe.Pointer(name)
						}
					}
					t.Fatal("unknown name pointer")
					return nil
				},
				storeImage: func(i int, image unsafe.Pointer) { trace = append(trace, 400+i); cached[i] = image },
				storeLoaded: func(v uint32) {
					trace = append(trace, 500)
					if v != 1 {
						t.Fatal("ready value")
					}
					loaded = v
				},
				loadFlag:        func(i int) byte { trace = append(trace, 600+i); return flags[i] },
				loadCachedImage: func(i int) unsafe.Pointer { trace = append(trace, 700+i); return cached[i] },
			}
			got := itemEnchantmentIcon413420(byte(query), h)
			wantTrace := []int{100}
			if initial == 0 {
				for i := 0; i < 6; i++ {
					wantTrace = append(wantTrace, 200+i, 300+i, 400+i)
				}
				wantTrace = append(wantTrace, 500)
			}
			var want unsafe.Pointer
			for i, flag := range flags {
				wantTrace = append(wantTrace, 600+i)
				if byte(query) == flag {
					wantTrace = append(wantTrace, 700+i)
					want = cached[i]
					break
				}
			}
			wantLoaded := initial
			if initial == 0 {
				wantLoaded = 1
			}
			if got != want || loaded != wantLoaded || !reflect.DeepEqual(trace, wantTrace) {
				t.Fatalf("initial=%x query=%x got=%p want=%p loaded=%x trace=%v want=%v", initial, query, got, want, loaded, trace, wantTrace)
			}
		}
	}
}

func TestItemEnchantmentIcon413420ReloadsAllImagesAfterResetEvenWhenNil(t *testing.T) {
	var loaded uint32
	var cached [6]unsafe.Pointer
	var loads []int
	var names [6]byte
	flags := [6]byte{8, 16, 1, 4, 2, 32}
	h := itemEnchantmentIconHooks413420{
		loadLoaded:      func() uint32 { return loaded },
		storeLoaded:     func(v uint32) { loaded = v },
		loadName:        func(i int) *byte { loads = append(loads, i); return &names[i] },
		loadImage:       func(*byte) unsafe.Pointer { return nil },
		storeImage:      func(i int, image unsafe.Pointer) { cached[i] = image },
		loadFlag:        func(i int) byte { return flags[i] },
		loadCachedImage: func(i int) unsafe.Pointer { return cached[i] },
	}
	for cycle := 0; cycle < 2; cycle++ {
		loaded, loads = 0, nil
		if got := itemEnchantmentIcon413420(0xff, h); got != nil || loaded != 1 || !reflect.DeepEqual(loads, []int{0, 1, 2, 3, 4, 5}) {
			t.Fatalf("invalid first query: got=%p loaded=%d loads=%v", got, loaded, loads)
		}
		loads = nil
		for _, flag := range flags {
			if got := itemEnchantmentIcon413420(flag, h); got != nil {
				t.Fatal("nil image changed")
			}
		}
		if len(loads) != 0 {
			t.Fatalf("nil images were loaded again without reset: %v", loads)
		}
	}
}

func TestItemEnchantmentIcon413420ReadsLiveFlagsAfterLoading(t *testing.T) {
	var marker byte
	flags := [6]byte{8, 16, 1, 4, 2, 32}
	h := itemEnchantmentIconHooks413420{
		loadLoaded:  func() uint32 { return 0 },
		loadName:    func(int) *byte { return &marker },
		loadImage:   func(*byte) unsafe.Pointer { flags[0] = 0xfe; return nil },
		storeImage:  func(int, unsafe.Pointer) {},
		storeLoaded: func(uint32) {},
		loadFlag:    func(i int) byte { return flags[i] },
		loadCachedImage: func(i int) unsafe.Pointer {
			if i != 0 {
				t.Fatal("lookup index")
			}
			return unsafe.Pointer(&marker)
		},
	}
	if got := itemEnchantmentIcon413420(0xfe, h); got != unsafe.Pointer(&marker) {
		t.Fatal("lookup used a pre-load flag snapshot")
	}
}
