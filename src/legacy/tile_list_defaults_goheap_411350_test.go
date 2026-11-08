//go:build darwin && arm64

package legacy

import (
	"fmt"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func tlFindControl411350(t *testing.T, cases []tlCase411350, predicate func(*tlInput411350, map[uint32][]byte) bool) *tlCase411350 {
	t.Helper()
	for i := range cases {
		pages, e := tlInitialPages411350(&cases[i].Input)
		if e != nil {
			t.Fatal(e)
		}
		if predicate(&cases[i].Input, pages) {
			return &cases[i]
		}
	}
	t.Fatal("literal input control not found")
	return nil
}
func tlZeroXY411350(in *tlInput411350, pages map[uint32][]byte) bool {
	b, e := swOriginalBytes410550(pages, in.Roles["xy"], 8)
	return e == nil && swWord410550(b[:4]) == 0 && swWord410550(b[4:]) == 0
}
func tlNodeNext411350(in *tlInput411350, pages map[uint32][]byte, role string) uint32 {
	b, e := swOriginalBytes410550(pages, in.Roles[role]+16, 4)
	if e != nil {
		panic(e)
	}
	return swWord410550(b)
}
func tlWidthControl411350(t *testing.T, cases []tlCase411350) *tlCase411350 {
	return tlFindControl411350(t, cases, func(in *tlInput411350, p map[uint32][]byte) bool {
		return in.Family == "finite-node-chain" && in.Args[2] == 0 && tlZeroXY411350(in, p) && tlNodeNext411350(in, p, "node0") == 0 && tlNodeNext411350(in, p, "node1") == 0 && tlNodeNext411350(in, p, "node2") == 0
	})
}
func tlChainControl411350(t *testing.T, cases []tlCase411350) *tlCase411350 {
	return tlFindControl411350(t, cases, func(in *tlInput411350, p map[uint32][]byte) bool {
		return in.Family == "finite-node-chain" && in.Args[2] == 0 && tlZeroXY411350(in, p) && tlNodeNext411350(in, p, "node0") == in.Roles["node1"] && tlNodeNext411350(in, p, "node1") == in.Roles["node2"] && tlNodeNext411350(in, p, "node2") == 0
	})
}
func tlControlReturn411350(t *testing.T, item *tlCase411350, value int32) {
	t.Helper()
	for _, cw := range []uint16{0x0e3f, 0x023f} {
		row := item.References[cw]
		if row == nil || row.Status != "returned" || row.Result == nil || uint32(value) != *row.Result {
			t.Fatalf("actual control return vs originalCW%04x", cw)
		}
	}
}
func tlProductionEdge411350(t *testing.T, in *tlInput411350) func() {
	t.Helper()
	type restore struct {
		pointer *byte
		value   byte
	}
	var saved []restore
	for _, w := range in.InitialWrites {
		if w.Address < 0x862000 || w.Address >= 0x864000 {
			continue
		}
		data := swHex410550(w.Hex)
		if len(data) != 1 {
			t.Fatal("control BYTE initializer")
		}
		off := w.Address - 0x85b3fc
		p := memmap.PtrOff(0x85b3fc, uintptr(off))
		witness := twoSettingsPointerFixturePointerWitness40A740(p)
		if !witness.Above4GiB || !witness.Low32LosesBits || witness.Native != uintptr(p) {
			t.Fatal("actual production EDGE Cowner pointer")
		}
		v := (*byte)(p)
		saved = append(saved, restore{v, *v})
		*v = data[0]
	}
	return func() {
		for i := len(saved) - 1; i >= 0; i-- {
			*saved[i].pointer = saved[i].value
		}
	}
}
func tlDirectCall411350(in *tlInput411350, backend string, head, xy unsafe.Pointer) int32 {
	if backend == "C" {
		switch in.Entry {
		case 0x411350:
			return tileListFixtureList411350(head, xy, int32(in.Args[2]))
		case 0x4113a0:
			return tileListFixturePoint411350(xy, int32(in.Args[1]))
		case 0x411490:
			return tileListFixtureEdge411350(int32(in.Args[0]), int32(in.Args[1]))
		}
	}
	if backend != "Go" {
		panic("backend")
	}
	switch in.Entry {
	case 0x411350:
		return sub411350Native(head, xy, int32(in.Args[2]))
	case 0x4113a0:
		return sub4113A0Native(xy, int32(in.Args[1]))
	case 0x411490:
		return sub411490Native(int32(in.Args[0]), int32(in.Args[1]))
	}
	panic("selected root")
}
func tlControlPointers411350(t *testing.T, g *tlGraph411350) (head, xy unsafe.Pointer) {
	t.Helper()
	var e error
	switch g.Input.Entry {
	case 0x411350:
		head, e = g.pointer(g.Input.Args[0])
		if e != nil {
			t.Fatal(e)
		}
		xy, e = g.pointer(g.Input.Args[1])
	case 0x4113a0:
		xy, e = g.pointer(g.Input.Args[0])
	case 0x411490:
	default:
		t.Fatal("control entry")
	}
	if e != nil {
		t.Fatal(e)
	}
	return
}

func TestTileListNativeDefaultLinked411350(t *testing.T) {
	cases, e := tlLoadCases411350()
	if e != nil {
		t.Fatal(e)
	}
	controls := []struct {
		name string
		item *tlCase411350
	}{
		{"point", tlFindControl411350(t, cases, func(in *tlInput411350, p map[uint32][]byte) bool {
			return in.Family == "xy-selector-grid" && in.Args[1] == 0 && tlZeroXY411350(in, p)
		})},
		{"list", tlChainControl411350(t, cases)},
		{"edge-wrap40000000", tlFindControl411350(t, cases, func(in *tlInput411350, p map[uint32][]byte) bool {
			return in.Family == "raw-index-address-domain" && in.Args[0] == 0x40000000 && in.Args[1] == 0x80000000
		})},
		{"edge-wrapC0000000", tlFindControl411350(t, cases, func(in *tlInput411350, p map[uint32][]byte) bool {
			return in.Family == "raw-index-address-domain" && in.Args[0] == 0xc0000000 && in.Args[1] == 0x80000000
		})},
	}
	for _, backend := range []string{"Go", "C"} {
		for _, control := range controls {
			t.Run(backend+"/"+control.name, func(t *testing.T) {
				if tileListNativeObserverCurrent411350() != nil {
					t.Fatal("default control requires nil observer")
				}
				g, e := tlBuildGraph411350(&control.item.Input, fmt.Sprintf("%s/bank.bin", t.TempDir()))
				if e != nil {
					t.Fatal(e)
				}
				defer g.close()
				head, xy := tlControlPointers411350(t, g)
				restore := tlProductionEdge411350(t, &control.item.Input)
				defer restore()
				before := g.snapshot()
				value := tlDirectCall411350(&control.item.Input, backend, head, xy)
				tlControlReturn411350(t, control.item, value)
				if !reflectBytesEqual411350(before, g.snapshot()) {
					t.Fatal("default readonly whole native banks")
				}
				t.Logf("TILE_LIST_DEFAULT_OK backend=%s case=%s real-EDGE-owner=true native-calls=1 original-comparisons=2", backend, control.item.Input.Name)
			})
		}
	}
}
func reflectBytesEqual411350(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTileListNativePinnedGoHeap411350(t *testing.T) {
	cases, e := tlLoadCases411350()
	if e != nil {
		t.Fatal(e)
	}
	chain := tlChainControl411350(t, cases)
	cache := tlFindControl411350(t, cases, func(in *tlInput411350, p map[uint32][]byte) bool {
		return in.Family == "completed-xy-read-effects" && in.AfterReads[0].PC == 0x4113a7
	})
	for _, backend := range []string{"Go", "C"} {
		for _, shape := range []string{"chain", "cached-xy"} {
			t.Run(backend+"/"+shape, func(t *testing.T) {
				if tileListNativeObserverCurrent411350() != nil {
					t.Fatal("serial pin control")
				}
				item := chain
				if shape == "cached-xy" {
					item = cache
				}
				in := &item.Input
				pages, e := tlInitialPages411350(in)
				if e != nil {
					t.Fatal(e)
				}
				nodes := [3]*tileListNativeNode411350{new(tileListNativeNode411350), new(tileListNativeNode411350), new(tileListNativeNode411350)}
				byToken := map[uint32]*tileListNativeNode411350{0: nil}
				for i, n := range nodes {
					byToken[in.Roles[fmt.Sprintf("node%d", i)]] = n
				}
				for i, n := range nodes {
					token := in.Roles[fmt.Sprintf("node%d", i)]
					b, e := swOriginalBytes410550(pages, token, 20)
					if e != nil {
						t.Fatal(e)
					}
					n.Value = int32(swWord410550(b[:4]))
					n.Preserved = int32(swWord410550(b[4:8]))
					n.Edge = int32(swWord410550(b[8:12]))
					n.Variation = int32(swWord410550(b[12:16]))
					next, ok := byToken[swWord410550(b[16:20])]
					if !ok {
						t.Fatal("typed known pin Next")
					}
					n.Next = next
				}
				xy := new([2]int32)
				b, e := swOriginalBytes410550(pages, in.Roles["xy"], 8)
				if e != nil {
					t.Fatal(e)
				}
				xy[0], xy[1] = int32(swWord410550(b[:4])), int32(swWord410550(b[4:]))
				initialNodes := [3]tileListNativeNode411350{*nodes[0], *nodes[1], *nodes[2]}
				expectedXY := *xy
				var pins [4]runtime.Pinner
				defer func() {
					for i := range pins {
						pins[i].Unpin()
					}
					runtime.KeepAlive(nodes)
					runtime.KeepAlive(xy)
				}()
				for i, n := range nodes {
					pins[i].Pin(n)
				}
				pins[3].Pin(xy)
				for _, n := range nodes {
					w := twoSettingsPointerFixturePointerWitness40A740(unsafe.Pointer(n))
					if !w.Above4GiB || !w.Low32LosesBits {
						t.Fatal("actual pinned high C pointer")
					}
				}
				w := twoSettingsPointerFixturePointerWitness40A740(unsafe.Pointer(xy))
				if !w.Above4GiB || !w.Low32LosesBits {
					t.Fatal("actual pinned xy pointer")
				}
				restore := tlProductionEdge411350(t, in)
				defer restore()
				gcCount, used := 0, 0
				observer := &tileListNativeObserver411350{}
				gc := func() { runtime.GC(); gcCount++ }
				observer.Call = func(call tileListNativeCall411350) { gc() }
				observer.Operand = func(op tileListNativeOperand411350) {
					gc()
					if !op.Complete {
						return
					}
					for _, s := range in.AfterReads {
						if s.PC != op.Site {
							continue
						}
						if s.Occurrence != 1 || used != 0 {
							panic("bounded pin read stimulus")
						}
						for _, write := range s.Writes {
							data := swHex410550(write.Hex)
							if len(data) != 4 || (write.Address != in.Roles["xy"] && write.Address != in.Roles["xy"]+4) {
								panic("pin scalar-only input effects")
							}
							index := int((write.Address - in.Roles["xy"]) / 4)
							xy[index] = int32(swWord410550(data))
							expectedXY[index] = xy[index]
						}
						used++
					}
				}
				pop := pushTileListNativeObserver411350(observer)
				defer pop()
				gc()
				value := tlDirectCall411350(in, backend, unsafe.Pointer(nodes[0]), unsafe.Pointer(xy))
				tlControlReturn411350(t, item, value)
				if used != len(in.AfterReads) || gcCount < 3 || *xy != expectedXY {
					t.Fatal("actual forcedGC/cache input effects")
				}
				for i, n := range nodes {
					if *n != initialNodes[i] {
						t.Fatal("whole typed pinned node canary")
					}
				}
				t.Logf("TILE_LIST_PINNED_OK backend=%s shape=%s native-calls=1 original-comparisons=2 owners=4 GC=%d used-input-effects=%d", backend, shape, gcCount, used)
				runtime.KeepAlive(nodes)
				runtime.KeepAlive(xy)
			})
		}
	}
}

func TestTileListNativeObserverReentry411350(t *testing.T) {
	cases, e := tlLoadCases411350()
	if e != nil {
		t.Fatal(e)
	}
	outerCase := tlFindControl411350(t, cases, func(in *tlInput411350, p map[uint32][]byte) bool {
		return in.Family == "xy-selector-grid" && in.Args[1] == 0 && tlZeroXY411350(in, p)
	})
	innerCase := tlFindControl411350(t, cases, func(in *tlInput411350, p map[uint32][]byte) bool {
		return in.Family == "raw-index-address-domain" && in.Args[0] == 0x40000000 && in.Args[1] == 0x80000000
	})
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			if tileListNativeObserverCurrent411350() != nil {
				t.Fatal("serial reentry")
			}
			g, e := tlBuildGraph411350(&outerCase.Input, fmt.Sprintf("%s/bank.bin", t.TempDir()))
			if e != nil {
				t.Fatal(e)
			}
			defer g.close()
			_, xy := tlControlPointers411350(t, g)
			restore := tlProductionEdge411350(t, &innerCase.Input)
			defer restore()
			entered, outerEvents, innerEvents := false, 0, 0
			outer := &tileListNativeObserver411350{}
			outer.Operand = func(op tileListNativeOperand411350) {
				outerEvents++
				if entered {
					return
				}
				entered = true
				inner := &tileListNativeObserver411350{Operand: func(tileListNativeOperand411350) { innerEvents++ }}
				value := func() int32 {
					pop := pushTileListNativeObserver411350(inner)
					defer pop()
					return tlDirectCall411350(&innerCase.Input, backend, nil, nil)
				}()
				tlControlReturn411350(t, innerCase, value)
				if tileListNativeObserverCurrent411350() != outer {
					panic("LIFO outer not restored")
				}
			}
			value := func() int32 {
				pop := pushTileListNativeObserver411350(outer)
				defer pop()
				return tlDirectCall411350(&outerCase.Input, backend, nil, xy)
			}()
			tlControlReturn411350(t, outerCase, value)
			if !entered || outerEvents != 4 || innerEvents != 4 || tileListNativeObserverCurrent411350() != nil {
				t.Fatal("actual nested operand scope and restore")
			}
			t.Logf("TILE_LIST_REENTRY_OK backend=%s native-calls=2 original-comparisons=4", backend)
		})
	}
}
