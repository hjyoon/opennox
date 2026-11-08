//go:build darwin && arm64

package legacy

import (
	"bytes"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Finite integration controls, separate from all original-reference rows.
// These controls never claim original First/Next/Free helper-body coverage.
func swExtraCall410550(t *testing.T, backend string, entry, word uint32, p unsafe.Pointer) unsafe.Pointer {
	t.Helper()
	if backend == "C" {
		return secretWallFixtureCall410550(entry, word, p)
	}
	if backend != "Go" {
		t.Fatal("unknown SecretWall extra backend")
	}
	switch entry {
	case 0x410550:
		return secretWallLookupNative410550(word)
	case 0x410730:
		return secretWallClearNative410730()
	case 0x4107a0:
		return secretWallRemoveNative4107A0(p)
	}
	t.Fatal("unknown SecretWall extra entry")
	return nil
}
func swExtraPointer410550(t *testing.T, p unsafe.Pointer) {
	t.Helper()
	w := twoSettingsPointerFixturePointerWitness40A740(p)
	if p == nil || w.Native != uintptr(p) || w.Low32 != uint32(uintptr(p)) || !w.Above4GiB || !w.Low32LosesBits {
		t.Fatalf("full actual C pointer witness: %p %+v", p, w)
	}
}
func swExtraBytes410550(p unsafe.Pointer, n uintptr) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(p), int(n))...)
}
func swExtraHead410550(t *testing.T) unsafe.Pointer {
	t.Helper()
	head := secretWallNativeHeadCell410550()
	if head != secretWallFixtureHeadCell410550() || secretWallFixturePointerWidth410550() != unsafe.Sizeof(unsafe.Pointer(nil)) {
		t.Fatal("actual production C pointer cell/width")
	}
	swExtraPointer410550(t, head)
	return head
}
func TestSecretWallNativeDefaultLinked410550(t *testing.T) {
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			if secretWallNativeObserverCurrent410550() != nil {
				t.Fatal("default control requires nil observer")
			}
			head := swExtraHead410550(t)
			old := secretWallFixtureLoadPointer410550(head)
			layout := secretWallFixtureLayout410550()
			offsets := secretWallNativeOffsets410550()
			for i := range offsets {
				if offsets[i] != layout[i] {
					t.Fatalf("actual C/Go named layout %d", i)
				}
			}
			var nodes [2]unsafe.Pointer
			var walls [2]unsafe.Pointer
			defer func() {
				secretWallFixtureStorePointer410550(head, old)
				for _, p := range nodes {
					if p != nil {
						secretWallFixtureFree410550(p)
					}
				}
				for _, p := range walls {
					if p != nil {
						secretWallFixtureFree410550(p)
					}
				}
			}()
			for i := 0; i < 2; i++ {
				nodes[i] = secretWallFixtureAlloc410550(layout[3] + 16)
				walls[i] = secretWallFixtureAlloc410550(secretWallFixtureWallSize410550() + 16)
				if nodes[i] == nil || walls[i] == nil {
					t.Fatal("finite C default allocation")
				}
				swExtraPointer410550(t, nodes[i])
				swExtraPointer410550(t, walls[i])
				ddiFill4E0B30(nodes[i], int(layout[3]+16), 0x4d)
				ddiFill4E0B30(walls[i], int(secretWallFixtureWallSize410550()+16), 0x6b)
				secretWallFixtureStorePointer410550(unsafe.Add(nodes[i], layout[0]), nil)
				secretWallFixtureStorePointer410550(unsafe.Add(nodes[i], layout[1]), walls[i])
				*(*uint16)(unsafe.Add(walls[i], layout[2])) = uint16(0x1111 + 0x1111*i)
			}
			secretWallFixtureStorePointer410550(unsafe.Add(nodes[0], layout[0]), nodes[1])
			secretWallFixtureStorePointer410550(head, nodes[0])
			wallBefore := [2][]byte{swExtraBytes410550(walls[0], secretWallFixtureWallSize410550()+16), swExtraBytes410550(walls[1], secretWallFixtureWallSize410550()+16)}
			node1Before := swExtraBytes410550(nodes[1], layout[3]+16)
			for i := 0; i < 2; i++ {
				if got := swExtraCall410550(t, backend, 0x410550, uint32(0x1111+0x1111*i), nil); got != walls[i] {
					t.Fatalf("real default linked lookup %d got %p want %p", i, got, walls[i])
				}
			}
			matched := nodes[0]
			got := swExtraCall410550(t, backend, 0x4107a0, 0, matched)
			nodes[0] = nil
			if got != matched || secretWallFixtureLoadPointer410550(head) != nodes[1] {
				t.Fatal("existing C cached removed-pointer return and actual live head")
			}
			if !bytes.Equal(swExtraBytes410550(nodes[1], layout[3]+16), node1Before) {
				t.Fatal("remaining C node full canary")
			}
			got = swExtraCall410550(t, backend, 0x410730, 0, nil)
			nodes[1] = nil
			if got != nil || secretWallFixtureLoadPointer410550(head) != nil {
				t.Fatal("real default clear and actual C head nil")
			}
			for i := range walls {
				if !bytes.Equal(swExtraBytes410550(walls[i], secretWallFixtureWallSize410550()+16), wallBefore[i]) {
					t.Fatal("real default wall full canary")
				}
			}
			t.Logf("SECRET_WALL_DEFAULT_LINKED_OK backend=%s calls=4 actual-C-head=true actual-First-Next-Free=true original-helper-credit=false", backend)
		})
	}
}
func TestSecretWallNativePinnedGoHeap410550(t *testing.T) {
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			if secretWallNativeObserverCurrent410550() != nil {
				t.Fatal("serial pinned control requires empty observer")
			}
			nodes := [3]*server.SecretWall{new(server.SecretWall), new(server.SecretWall), new(server.SecretWall)}
			walls := [2]*server.Wall{new(server.Wall), new(server.Wall)}
			walls[0].Field10 = 0x1111
			walls[1].Field10 = 0x2222
			nodes[0].Next = nodes[1]
			nodes[1].Next = nodes[2]
			nodes[0].Wall = walls[0]
			nodes[1].Wall = walls[1]
			nodes[2].Wall = walls[0]
			for i, n := range nodes {
				n.X = int32(17 + i)
				n.Y = int32(-23 - i)
				n.OpenWait = uint32(0xa100 + i)
				n.PlayerBits = uint32(0xc200 + i)
			}
			var pins [5]runtime.Pinner
			defer func() {
				for i := range pins {
					pins[i].Unpin()
				}
				runtime.KeepAlive(nodes)
				runtime.KeepAlive(walls)
			}()
			for i, n := range nodes {
				pins[i].Pin(n)
			}
			for i, w := range walls {
				pins[3+i].Pin(w)
			}
			// Every transitive owner is pinned before the first actual C boundary,
			// including the numeric pointer witness. No uintptr hiding is used.
			for _, n := range nodes {
				swExtraPointer410550(t, unsafe.Pointer(n))
			}
			for _, w := range walls {
				swExtraPointer410550(t, unsafe.Pointer(w))
			}
			head := swExtraHead410550(t)
			old := secretWallFixtureLoadPointer410550(head)
			secretWallFixtureStorePointer410550(head, unsafe.Pointer(nodes[0]))
			counts := 0
			frees := []unsafe.Pointer{}
			gc := func(point string) {
				counts++
				runtime.GC()
				t.Logf("SECRET_WALL_PINNED_GC_OK backend=%s point=%s count=%d owners=5", backend, point, counts)
			}
			observer := &secretWallNativeObserver410550{Service: func(site uint32, args []secretWallNativeArg410550) (secretWallNativeReply410550, bool) {
				switch site {
				case 0x41073d, 0x4107e0, 0x4107f6:
					if len(args) != 1 || !args[0].IsPointer {
						t.Fatal("pinned Free actual pointer argument")
					}
					frees = append(frees, args[0].Pointer)
					gc("handled-Free")
					return secretWallNativeReply410550{Word: 0x80000001, ECX: 0xffffffff, EDX: 2}, true
				}
				return secretWallNativeReply410550{}, false
			}}
			pop := pushSecretWallNativeObserver410550(observer)
			// This later defer runs before the earlier Unpin defer, clearing
			// all C-retained references to these owners even when an assertion fails.
			defer func() { secretWallFixtureStorePointer410550(head, old); pop() }()
			var beforeNodes [3][]byte
			var beforeWalls [2][]byte
			for i, n := range nodes {
				beforeNodes[i] = swExtraBytes410550(unsafe.Pointer(n), unsafe.Sizeof(*n))
			}
			for i, w := range walls {
				beforeWalls[i] = swExtraBytes410550(unsafe.Pointer(w), unsafe.Sizeof(*w))
			}
			gc("before-lookup")
			got := swExtraCall410550(t, backend, 0x410550, 0x2222, nil)
			gc("after-lookup")
			if got != unsafe.Pointer(walls[1]) {
				t.Fatalf("pinned default First/Next lookup got %p want %p", got, walls[1])
			}
			// Prepare an independent expected object. The actual SUT node is never
			// overwritten to make its full bytes agree with this expectation.
			expectedFirst := *nodes[0]
			expectedFirst.Next = nodes[2]
			expectedFirstBytes := swExtraBytes410550(unsafe.Pointer(&expectedFirst), unsafe.Sizeof(expectedFirst))
			gc("before-middle-remove")
			got = swExtraCall410550(t, backend, 0x4107a0, 0, unsafe.Pointer(nodes[1]))
			gc("after-middle-remove")
			if got != unsafe.Pointer(nodes[1]) || secretWallFixtureLoadPointer410550(head) != unsafe.Pointer(nodes[0]) || nodes[0].Next != nodes[2] || len(frees) != 1 || frees[0] != unsafe.Pointer(nodes[1]) {
				t.Fatal("pinned middle removal: cached return, real head, previous.Next, handled Free identity")
			}
			if !bytes.Equal(swExtraBytes410550(unsafe.Pointer(nodes[0]), unsafe.Sizeof(*nodes[0])), expectedFirstBytes) {
				t.Fatal("pinned actual typed previous.Next store and full node canary")
			}
			for i := 1; i < len(nodes); i++ {
				if !bytes.Equal(swExtraBytes410550(unsafe.Pointer(nodes[i]), unsafe.Sizeof(*nodes[i])), beforeNodes[i]) {
					t.Fatalf("pinned untouched node after middle removal %d", i)
				}
			}
			gc("before-clear")
			got = swExtraCall410550(t, backend, 0x410730, 0, nil)
			gc("after-clear")
			if got != nil || secretWallFixtureLoadPointer410550(head) != nil || len(frees) != 3 || counts != 9 {
				t.Fatalf("pinned clear retention/GC got=%p frees=%d GC=%d", got, len(frees), counts)
			}
			for i, n := range []*server.SecretWall{nodes[1], nodes[0], nodes[2]} {
				if frees[i] != unsafe.Pointer(n) {
					t.Fatalf("pinned ordered Free identity %d", i)
				}
			}
			for i, n := range nodes {
				expected := beforeNodes[i]
				if i == 0 {
					expected = expectedFirstBytes
				}
				if !bytes.Equal(swExtraBytes410550(unsafe.Pointer(n), unsafe.Sizeof(*n)), expected) {
					t.Fatalf("pinned full node canary after actual middle-store and clear %d", i)
				}
			}
			for i, w := range walls {
				if !bytes.Equal(swExtraBytes410550(unsafe.Pointer(w), unsafe.Sizeof(*w)), beforeWalls[i]) {
					t.Fatal("pinned full wall canary")
				}
			}
			t.Logf("SECRET_WALL_PINNED_OK backend=%s calls=3 pins=5 GC=9 default-First-Next=true handled-Free=true production-head=true typed-previous-next-store=true", backend)
		})
	}
}
func TestSecretWallNativeObserverReentry410550(t *testing.T) {
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			if secretWallNativeObserverCurrent410550() != nil {
				t.Fatal("serial reentry control")
			}
			head := swExtraHead410550(t)
			old := secretWallFixtureLoadPointer410550(head)
			secretWallFixtureStorePointer410550(head, nil)
			defer secretWallFixtureStorePointer410550(head, old)
			outerCalls, innerCalls := 0, 0
			outer := &secretWallNativeObserver410550{}
			outer.Service = func(site uint32, args []secretWallNativeArg410550) (secretWallNativeReply410550, bool) {
				if site != 0x410551 || len(args) != 0 {
					t.Fatal("outer First actual argc")
				}
				outerCalls++
				inner := &secretWallNativeObserver410550{Service: func(site uint32, args []secretWallNativeArg410550) (secretWallNativeReply410550, bool) {
					if site != 0x410551 || len(args) != 0 {
						t.Fatal("inner First actual argc")
					}
					innerCalls++
					return secretWallNativeReply410550{}, true
				}}
				popInner := pushSecretWallNativeObserver410550(inner)
				func() {
					defer popInner()
					if swExtraCall410550(t, backend, 0x410550, 0xffff, nil) != nil {
						t.Fatal("nested nil First")
					}
				}()
				if secretWallNativeObserverCurrent410550() != outer {
					t.Fatal("outer observer restored after actual nested call")
				}
				return secretWallNativeReply410550{}, true
			}
			pop := pushSecretWallNativeObserver410550(outer)
			defer pop()
			if swExtraCall410550(t, backend, 0x410550, 0xffff, nil) != nil || outerCalls != 1 || innerCalls != 1 {
				t.Fatal("actual reentry and nil gates")
			}
			t.Logf("SECRET_WALL_REENTRY_OK backend=%s calls=2 outer=1 inner=1", backend)
		})
	}
}
func TestSecretWallNativeObserverLIFO410550(t *testing.T) {
	if secretWallNativeObserverCurrent410550() != nil {
		t.Fatal("serial LIFO control")
	}
	outer, inner := &secretWallNativeObserver410550{}, &secretWallNativeObserver410550{}
	popOuter := pushSecretWallNativeObserver410550(outer)
	defer popOuter()
	popInner := pushSecretWallNativeObserver410550(inner)
	panicked := false
	func() { defer func() { panicked = recover() != nil }(); popOuter() }()
	if !panicked || secretWallNativeObserverCurrent410550() != inner {
		t.Fatal("out-of-order pop must reject without changing current observer")
	}
	popInner()
	if secretWallNativeObserverCurrent410550() != outer {
		t.Fatal("inner pop restores exact outer observer")
	}
	t.Log("SECRET_WALL_LIFO_OK out-of-order-rejected=true state-preserved=true")
}
