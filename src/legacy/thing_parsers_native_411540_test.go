package legacy

import (
	"bytes"
	"fmt"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func TestThingParserNativeLayout411540(t *testing.T) {
	got := thingFixtureLayout411540()
	f := nox_memfile{}
	p := server.TileDef{}
	want := []uintptr{unsafe.Sizeof(unsafe.Pointer(nil)), unsafe.Sizeof(f), unsafe.Offsetof(f.data), unsafe.Offsetof(f.size), unsafe.Offsetof(f.cur), unsafe.Offsetof(f.end), unsafe.Sizeof(p), unsafe.Offsetof(p.Data32), unsafe.Offsetof(p.Field36), unsafe.Offsetof(p.Color48), unsafe.Offsetof(p.Field58)}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("actual C/Go layout %v != %v", got, want)
	}
}

func TestThingParserPreservedC411540(t *testing.T) {
	_, restore := thingFixtureSave411540()
	defer restore()
	if !thingFixtureSignedChar411540() {
		t.Fatal("preserved sentinel RGB controls require signed char; no undefined old RGB reads admitted")
	}
	for _, x := range thingInputs411540() {
		t.Run(x.name, func(t *testing.T) {
			type result struct {
				ret                   int32
				offset                int
				state, scratch, input []byte
			}
			run := func(backend int) result {
				thingFixtureSeed411540(x.floor, x.edge)
				f, free := thingFixtureMemfile411540(x.data, x.start)
				defer free()
				scratch, drop := thingFixtureScratch411540()
				defer drop()
				p := unsafe.Pointer(&scratch[0])
				for _, v := range []unsafe.Pointer{unsafe.Pointer(f), unsafe.Pointer(f.data), p, unsafe.Pointer(thingTileOwner411540(0)), unsafe.Pointer(thingEdgeImageCell411850(0))} {
					if thingFixturePointer411540(v) != uintptr(v) {
						t.Fatal("C pointer identity")
					}
					if unsafe.Sizeof(v) == 8 && uintptr(v) <= 0xffffffff {
						t.Fatal("64-bit fixture must use actual high native addresses")
					}
				}
				var ret int32
				if backend == 0 {
					ret = thingFixtureCall411540(x.entry, true, f, p)
				} else if backend == 1 {
					ret = thingFixtureDirect411540(x.entry, f, p)
				} else {
					ret = thingFixtureCall411540(x.entry, false, f, p)
				}
				return result{ret, thingFixtureOffset411540(f), thingFixtureState411540(), append([]byte(nil), scratch...), thingFixtureInput411540(f, len(x.data))}
			}
			reference := run(0)
			for _, backend := range []int{1, 2} {
				got := run(backend)
				if !reflect.DeepEqual(reference, got) {
					t.Fatalf("backend%d return/cursor/native-owner/scratch/input+guards mismatch: ref%d/%d got%d/%d", backend, reference.ret, reference.offset, got.ret, got.offset)
				}
			}
		})
	}
}

func TestThingParserRGBPolicy411540(t *testing.T) {
	_, restore := thingFixtureSave411540()
	defer restore()
	// The old non-sentinel branch reads uninitialized high cbuf bytes. This is
	// a new bounded three-BYTE semantic policy, not an old-stack equivalence.
	for _, rgb := range [][3]byte{{19, 128, 251}, {0, 0, 0}, {255, 255, 255}, {128, 1, 254}} {
		for _, public := range []bool{false, true} {
			t.Run(fmt.Sprintf("%x-%t", rgb, public), func(t *testing.T) {
				var w thingWire411540
				w.word(0)
				w.str([]byte("RGB"))
				w.b(rgb[:]...)
				w.word(17)
				w.word(29)
				w.b(0, 0, 2, 1, 0)
				thingFixtureSeed411540(0, 0)
				f, free := thingFixtureMemfile411540(w, 0)
				defer free()
				var result int32
				if public {
					result = thingFixtureCall411540(0x411540, false, f, nil)
				} else {
					result = thingFixtureDirect411540(0x411540, f, nil)
				}
				if result != 1 || thingFixtureOffset411540(f) != len(w) || thingTileOwner411540(0).Color48 != thingFixtureRGB411540(rgb[0], rgb[1], rgb[2]) {
					t.Fatal("three-channel/color-endpoint policy")
				}
			})
		}
	}
}

func TestThingParserMalformedPolicies411540(t *testing.T) {
	_, restore := thingFixtureSave411540()
	defer restore()
	// No malformed input is sent to preserved C. Every proper-prefix cut is
	// bounded; a cut can still be a valid early-return domain (e.g. zero count).
	for _, x := range thingInputs411540() {
		if x.start != 0 {
			continue
		}
		for cut := 0; cut < len(x.data); cut++ {
			t.Run(fmt.Sprintf("%s/cut-%d", x.name, cut), func(t *testing.T) {
				for _, cgo := range []bool{false, true} {
					func() {
						thingFixtureSeed411540(0, 0)
						f, free := thingFixtureMemfile411540(x.data[:cut], 0)
						defer free()
						scratch, drop := thingFixtureScratch411540()
						defer drop()
						before := thingFixtureInput411540(f, cut)
						var ret int32
						if cgo {
							ret = thingFixtureCall411540(x.entry, false, f, unsafe.Pointer(&scratch[0]))
						} else {
							ret = thingFixtureDirect411540(x.entry, f, unsafe.Pointer(&scratch[0]))
						}
						off := thingFixtureOffset411540(f)
						if (ret != 0 && ret != 1) || off < 0 || off > cut || !bytes.Equal(before, thingFixtureInput411540(f, cut)) {
							t.Fatal("malformed read escaped bounded input or return domain")
						}
					}()
				}
			})
		}
	}
	for _, entry := range []uint32{0x411540, 0x411850} {
		for _, n := range []byte{32, 255} {
			t.Run(fmt.Sprintf("long-name-%x-%d", entry, n), func(t *testing.T) {
				thingFixtureSeed411540(0, 0)
				data := append([]byte{0, 0, 0, 0, n}, bytes.Repeat([]byte{'L'}, int(n))...)
				f, free := thingFixtureMemfile411540(data, 0)
				defer free()
				before := thingFixtureState411540()
				if thingFixtureCall411540(entry, false, f, nil) != 0 || thingFixtureOffset411540(f) != 5 || !bytes.Equal(before, thingFixtureState411540()) {
					t.Fatal("long-name policy must reject before owner write/payload read")
				}
			})
		}
	}
}

func TestThingParserMemfileAndPinned411540(t *testing.T) {
	_, restore := thingFixtureSave411540()
	defer restore()
	data, unusedFree := alloc.CloneSlice([]byte{0, 0, 0, 0})
	_ = unusedFree
	f := binfile.NewMemFile(unsafe.Pointer(&data[0]), len(data))
	defer f.Free()
	if thingFixtureCall411540(0x415240, false, (*nox_memfile)(f.C()), nil) != 1 {
		t.Fatal("actual binfile owner/API zero count")
	}
	off, _ := f.Seek(0, 1)
	if off != 4 {
		t.Fatalf("native cursor/Go API %d", off)
	}
	for _, public := range []bool{false, true} {
		t.Run(fmt.Sprintf("pinned-public-%t", public), func(t *testing.T) {
			input := append([]byte(nil), thingInputs411540()[0].data...)
			scratch := make([]byte, 256*1024)
			mf := thingFixtureGoMemfile411540(input)
			var pin runtime.Pinner
			pin.Pin(mf)
			pin.Pin(&input[0])
			pin.Pin(&scratch[0])
			defer pin.Unpin()
			thingFixtureSeed411540(0, 0)
			calls := 0
			badOwner := false
			pop := pushThingParserServices411540(&thingParserServices411540{Facade: func(p *server.TileDef) int32 {
				calls++
				runtime.GC()
				badOwner = p != thingTileOwner411540(0)
				return 1
			}})
			defer pop()
			var v int32
			if public {
				v = thingFixtureCall411540(0x411540, false, mf, unsafe.Pointer(&scratch[0]))
			} else {
				v = thingFixtureDirect411540(0x411540, mf, unsafe.Pointer(&scratch[0]))
			}
			if v != 1 || calls != 1 || badOwner || thingTileOwner411540(0).Field58 != 1 {
				t.Fatal("pinned owners/default facade service call")
			}
			runtime.KeepAlive(input)
			runtime.KeepAlive(mf)
			runtime.KeepAlive(scratch)
		})
	}
}

func TestThingParserLimitsAndServices411540(t *testing.T) {
	_, restore := thingFixtureSave411540()
	defer restore()
	thingFixtureSeed411540(176, 64)
	for _, entry := range []uint32{0x411540, 0x411850} {
		if thingFixtureCall411540(entry, true, nil, nil) != 0 || thingFixtureCall411540(entry, false, nil, nil) != 0 || thingFixtureDirect411540(entry, nil, nil) != 0 {
			t.Fatal("full owners must reject before nil file/buffer access")
		}
	}
	thingFixtureSeed411540(0, 0)
	for _, entry := range []uint32{0x411540, 0x411850, 0x414D40, 0x414DB0, 0x414E70, 0x414F60, 0x415100, 0x415240, 0x415320, 0x415660} {
		if thingFixtureCall411540(entry, false, nil, nil) != 0 {
			t.Fatal("nil native memfile safety policy")
		}
	}
	for _, entry := range []uint32{0x411540, 0x411850, 0x414D40, 0x414DB0, 0x414E70, 0x414F60, 0x415100, 0x415240, 0x415320, 0x415660} {
		for cut := 0; cut < 4; cut++ {
			f, free := thingFixtureMemfile411540(make([]byte, cut), 0)
			if thingFixtureCall411540(entry, false, f, nil) != 0 || thingFixtureOffset411540(f) != cut {
				free()
				t.Fatal("short first DWORD must fail at EOF")
			}
			free()
		}
	}
	for _, public := range []bool{false, true} {
		t.Run(fmt.Sprintf("audio-live-failure-%t", public), func(t *testing.T) {
			var input []byte
			for _, x := range thingInputs411540() {
				if x.entry == 0x415660 && x.start == 0 && len(x.data) > 4 {
					input = x.data
					break
				}
			}
			f, free := thingFixtureMemfile411540(input, 0)
			defer free()
			scratch, drop := thingFixtureScratch411540()
			defer drop()
			calls, afterFirst := 0, 0
			pop := pushThingParserServices411540(&thingParserServices411540{Audio: func(actual *nox_memfile, p unsafe.Pointer) int32 {
				calls++
				if actual != f || p != unsafe.Pointer(&scratch[0]) {
					return 0
				}
				if calls == 2 {
					return 0
				}
				v := thingFixtureAudioReal415660(actual, p)
				afterFirst = thingFixtureOffset411540(actual)
				return v
			}})
			defer pop()
			var ret int32
			if public {
				ret = thingFixtureCall411540(0x415660, false, f, unsafe.Pointer(&scratch[0]))
			} else {
				ret = thingFixtureDirect411540(0x415660, f, unsafe.Pointer(&scratch[0]))
			}
			if ret != 0 || calls != 2 || afterFirst <= 4 || thingFixtureOffset411540(f) != afterFirst {
				t.Fatal("outer audio caller must stop at actual helper zero and retain live cursor")
			}
		})
	}
	// A signed negative block skip is malformed; never run its old backward
	// cursor access. The new policy returns0 at EOF without touching input.
	for _, entry := range []uint32{0x415100, 0x415320} {
		var w thingWire411540
		w.word(1)
		w.str(nil)
		if entry == 0x415100 {
			w.b(0, 0, 0)
			w.str(nil)
			w.ref(false)
			w.ref(false)
			w.word(0)
		} else {
			w.b(0)
			for i := 0; i < 3; i++ {
				w.ref(false)
			}
		}
		w.str(nil)
		w.short(0xffff)
		w.b(0, 0, 0)
		f, free := thingFixtureMemfile411540(w, 0)
		if thingFixtureCall411540(entry, false, f, nil) != 0 || thingFixtureOffset411540(f) != len(w) {
			free()
			t.Fatal("negative skip safety policy")
		}
		free()
	}
}

// Test fixture direct-Go dispatch shares the exact public entry cores.
func thingFixtureDirect411540(entry uint32, f *nox_memfile, p unsafe.Pointer) int32 {
	switch entry {
	case 0x411540:
		return thingReadFloorNative411540(f, p)
	case 0x411850:
		return thingReadEdgeNative411850(f, p)
	case 0x414D40:
		return thingSkipAudioNative414D40(f)
	case 0x414DB0:
		return thingReadFloorNative414DB0(f)
	case 0x414E70:
		return thingReadEdgeNative414E70(f, p)
	case 0x414F60:
		return thingReadWallNative414F60(f, p)
	case 0x415100:
		return thingSkipSpellsNative415100(f)
	case 0x415240:
		return thingReadImageNative415240(f)
	case 0x415320:
		return thingReadAbilityNative415320(f)
	case 0x415660:
		return thingReadAudioNative415660(f, p)
	default:
		panic("thing fixture unknown entry")
	}
}
