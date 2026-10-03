package legacy

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestItemEnchantmentTooltip413480AllBytesAndReadOrder(t *testing.T) {
	flags := [6]byte{8, 16, 1, 4, 2, 32}
	for query := 0; query < 256; query++ {
		var keys, text [6]byte
		var trace []int
		h := itemEnchantmentTooltipHooks413480{
			loadFlag: func(i int) byte { trace = append(trace, i); return flags[i] },
			loadKey:  func(i int) *byte { trace = append(trace, 10+i); return &keys[i] },
			loadText: func(key *byte, source string, line int32) unsafe.Pointer {
				if source != `C:\NoxPost\src\common\Object\Modifier.c` || line != 2087 {
					t.Fatalf("source=%q line=%d", source, line)
				}
				for i := range keys {
					if key == &keys[i] {
						trace = append(trace, 20+i)
						return unsafe.Pointer(&text[i])
					}
				}
				t.Fatal("unknown key pointer")
				return nil
			},
		}
		got := itemEnchantmentTooltip413480(byte(query), h)
		var want unsafe.Pointer
		var wantTrace []int
		for i, flag := range flags {
			wantTrace = append(wantTrace, i)
			if byte(query) == flag {
				wantTrace = append(wantTrace, 10+i, 20+i)
				want = unsafe.Pointer(&text[i])
				break
			}
		}
		if got != want || !reflect.DeepEqual(trace, wantTrace) {
			t.Fatalf("byte=%x got=%p want=%p trace=%v want=%v", query, got, want, trace, wantTrace)
		}
	}
}

func TestItemEnchantmentTooltip413480ReadsLiveFlagAndLateKey(t *testing.T) {
	var oldKey, newKey, text byte
	flags := [6]byte{8, 16, 1, 4, 2, 32}
	key := &oldKey
	reads, loads := 0, 0
	h := itemEnchantmentTooltipHooks413480{
		loadFlag: func(i int) byte {
			if i == 0 {
				flags[3] = 0xfe
			}
			if i == 3 {
				key = &newKey
			}
			return flags[i]
		},
		loadKey: func(i int) *byte {
			reads++
			if i != 3 {
				t.Fatal("key read before a complete match")
			}
			return key
		},
		loadText: func(got *byte, _ string, _ int32) unsafe.Pointer {
			loads++
			if got != &newKey {
				t.Fatal("key was prefetched before the matching flag")
			}
			return unsafe.Pointer(&text)
		},
	}
	if got := itemEnchantmentTooltip413480(0xfe, h); got != unsafe.Pointer(&text) || reads != 1 || loads != 1 {
		t.Fatalf("got=%p key reads=%d text loads=%d", got, reads, loads)
	}
}

func TestItemEnchantmentTooltip413480NilKeyAndNilResultStillLoad(t *testing.T) {
	loads := 0
	h := itemEnchantmentTooltipHooks413480{
		loadFlag: func(int) byte { return 0xff },
		loadKey: func(i int) *byte {
			if i != 0 {
				t.Fatal("first match must win")
			}
			return nil
		},
		loadText: func(key *byte, _ string, _ int32) unsafe.Pointer {
			loads++
			if key != nil {
				t.Fatal("nil key changed")
			}
			return nil
		},
	}
	for i := 0; i < 2; i++ {
		if got := itemEnchantmentTooltip413480(0xff, h); got != nil {
			t.Fatal("nil return changed")
		}
	}
	if loads != 2 {
		t.Fatalf("tooltip must not cache strings: loads=%d", loads)
	}
}
