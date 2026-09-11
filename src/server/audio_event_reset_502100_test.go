package server

import (
	"reflect"
	"testing"
)

func TestResetAudioEvents502100OrderAndNativeHandle(t *testing.T) {
	const (
		classToken  = uint64(0x100000001)
		oldHead     = uint64(0x200000002)
		mutatedHead = uint64(0x300000003)
	)
	head := oldHead
	var trace []string

	resetAudioEvents502100(audioEventResetHooks502100[uint64]{
		loadClass: func() uint64 {
			trace = append(trace, "load-class")
			return classToken
		},
		freeAllObjects: func(class uint64) {
			trace = append(trace, "free-all")
			if class != classToken {
				t.Fatalf("class = %#x, want full native token %#x", class, classToken)
			}
			if head != oldHead {
				t.Fatalf("head before free = %#x, want %#x", head, oldHead)
			}
			head = mutatedHead
		},
		clearHead: func() {
			trace = append(trace, "clear-head")
			if head != mutatedHead {
				t.Fatalf("head before clear = %#x, want callback mutation %#x", head, mutatedHead)
			}
			head = 0
		},
	})

	if head != 0 {
		t.Fatalf("head = %#x, want zero", head)
	}
	if want := []string{"load-class", "free-all", "clear-head"}; !reflect.DeepEqual(trace, want) {
		t.Fatalf("trace = %v, want %v", trace, want)
	}
}

func TestResetAudioEvents502100NilClassStillClearsHead(t *testing.T) {
	head := true
	freeCalls := 0
	resetAudioEvents502100(audioEventResetHooks502100[*struct{}]{
		loadClass: func() *struct{} {
			return nil
		},
		freeAllObjects: func(class *struct{}) {
			freeCalls++
			if class != nil {
				t.Fatalf("class = %p, want nil", class)
			}
		},
		clearHead: func() {
			head = false
		},
	})
	if freeCalls != 1 {
		t.Fatalf("free calls = %d, want 1", freeCalls)
	}
	if head {
		t.Fatal("nil allocation class left the event head set")
	}
}
