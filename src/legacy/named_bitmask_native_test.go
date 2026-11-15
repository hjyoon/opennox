package legacy

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestNamedBitmaskNativeLayoutCABI(t *testing.T) {
	ptr, word := namedBitmaskSizesC()
	if ptr != unsafe.Sizeof(uintptr(0)) || word != 4 {
		t.Fatal("native pointer array stride/packed mask ABI", ptr, word)
	}
}

func TestNamedBitmaskNativeOriginalCABI(t *testing.T) {
	cases := namedBitmaskCases(t)
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			v := namedBitmaskOwnersNew(t, placement)
			for i, c := range cases {
				t.Run(namedBitmaskCaseName(i, c), func(t *testing.T) {
					v.reset(t, c)
					check := v.expect(t, c)
					input, bits, names := v.pointers()
					trace := make([]map[string]any, 0)
					var cursor unsafe.Pointer
					appendTrace := func(service string, fields map[string]any) {
						fields["service"] = service
						trace = append(trace, fields)
					}
					o := &namedBitmaskObserver423930{
						Compare: func(a, b unsafe.Pointer) int32 {
							left, right := alloc.GoString((*byte)(a)), alloc.GoString((*byte)(b))
							result := int32(bytes.Compare(bytes.ToLower([]byte(left)), bytes.ToLower([]byte(right))))
							appendTrace("stricmp_ASCII", map[string]any{"left_hex": hex.EncodeToString([]byte(left)), "right_hex": hex.EncodeToString([]byte(right)), "result": result})
							return result
						},
						Prefix: func(p unsafe.Pointer) int32 {
							left := []byte(alloc.GoString((*byte)(p)))
							if len(left) > 4 {
								left = left[:4]
							}
							right := []byte("NULL")
							result := int32(bytes.Compare(left, right))
							if p == input {
								t.Fatal("tokenizer must own an input copy")
							}
							appendTrace("strncmp4", map[string]any{"left_hex": hex.EncodeToString(left), "right_hex": hex.EncodeToString(right), "count": 4, "result": result})
							return result
						},
						Token: func(p unsafe.Pointer) unsafe.Pointer {
							if p != nil {
								cursor = p
							}
							for cursor != nil && *(*byte)(cursor) == '+' {
								cursor = unsafe.Add(cursor, 1)
							}
							var token unsafe.Pointer
							var text any
							if cursor != nil && *(*byte)(cursor) != 0 {
								token = cursor
								for *(*byte)(cursor) != 0 && *(*byte)(cursor) != '+' {
									cursor = unsafe.Add(cursor, 1)
								}
								if *(*byte)(cursor) == '+' {
									*(*byte)(cursor) = 0
									cursor = unsafe.Add(cursor, 1)
								} else {
									cursor = nil
								}
								text = namedBitmaskHexCString(token)
							} else {
								cursor = nil
							}
							appendTrace("strtok_plus", map[string]any{"initial": p != nil, "token_hex": text, "result": token != nil})
							return token
						},
					}
					restore := namedBitmaskObserve423930(o)
					defer restore()
					if c.Kind == "setter" {
						if c.Result == nil {
							t.Fatal("missing original setter status")
						}
						if got := namedBitmaskSetC(input, bits, names); got != *c.Result {
							t.Fatal("actual C setter status", got, *c.Result)
						}
					} else {
						if c.Result != nil {
							t.Fatal("void parser status must not be inferred")
						}
						namedBitmaskParseC(input, bits, names)
					}
					got, _ := json.Marshal(trace)
					want, _ := json.Marshal(c.Trace)
					if !bytes.Equal(got, want) {
						t.Fatalf("declared libc transcript differs\ngot %s\nwant%s", got, want)
					}
					check()
				})
			}
		})
	}
	t.Log("unchanged original named-bitmask cases=1600, instruction bytes=223; original PE image immutable; declared libc boundaries=3; generic native pointer arrays and32bit indices")
}

func TestNamedBitmaskNativeDefaultCABI(t *testing.T) {
	cases := namedBitmaskCases(t)
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			v := namedBitmaskOwnersNew(t, placement)
			for i, c := range cases {
				t.Run(namedBitmaskCaseName(i, c), func(t *testing.T) {
					v.reset(t, c)
					check := v.expect(t, c)
					input, bits, names := v.pointers()
					if c.Kind == "setter" {
						if got := namedBitmaskSetC(input, bits, names); got != *c.Result {
							t.Fatal("unmocked C comparison status", got, *c.Result)
						}
					} else {
						namedBitmaskParseC(input, bits, names)
					}
					check()
				})
			}
		})
	}
}

func TestNamedBitmaskNativePackedMonsterOwnerCABI(t *testing.T) {
	namedBitmaskStockOwner(t)
	cases := namedBitmaskCases(t)
	const base = uintptr(247536)
	var stock []string
	for i := uintptr(0); ; i++ {
		p := *memmap.PtrPtr(0x587000, base+4*i)
		if p == nil {
			break
		}
		stock = append(stock, namedBitmaskHexCString(p))
		if i > 64 {
			t.Fatal("stock table unterminated")
		}
	}
	if len(stock) != 22 {
		t.Fatal("original stock table size", len(stock))
	}
	before := bytes.Clone(memmap.Slice(0x587000, base-16)[:672])
	var ptrs [23]unsafe.Pointer
	for i := range ptrs {
		ptrs[i] = *memmap.PtrPtr(0x587000, base+4*uintptr(i))
	}
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			v := namedBitmaskOwnersNew(t, placement)
			count := 0
			for i, c := range cases {
				if c.Kind != "parser" || !reflect.DeepEqual(c.Names, stock) {
					continue
				}
				count++
				t.Run(namedBitmaskCaseName(i, c), func(t *testing.T) {
					v.reset(t, c)
					check := v.expect(t, c)
					input, bits, _ := v.pointers()
					namedBitmaskMonsterC(input, bits)
					check()
					if got := namedBitmaskMonsterLow16C(input, c.BeforeMask); got != uint16(c.AfterMask) {
						t.Fatal("actual packed caller uint16narrowing", got, uint16(c.AfterMask))
					}
					if !bytes.Equal(before, memmap.Slice(0x587000, base-16)[:len(before)]) {
						t.Fatal("packed stock pointer table/strings changed")
					}
					for j, want := range ptrs {
						if *memmap.PtrPtr(0x587000, base+4*uintptr(j)) != want {
							t.Fatal("stock native name sidecar changed")
						}
					}
				})
			}
			if count != 85 {
				t.Fatal("missing original stock parser cases", count)
			}
		})
	}
}
