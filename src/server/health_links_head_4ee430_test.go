package server

import "testing"

func TestHealthLinksHead4EE430LoadsOnceAndReturnsLoadedPointer(t *testing.T) {
	head := "head-1"
	loads := 0
	got := healthLinksHead4EE430(func() string {
		loaded := head
		loads++
		head = "head-2"
		return loaded
	})
	if got != "head-1" || loads != 1 || head != "head-2" {
		t.Fatalf("result = %q, loads = %d, live head = %q", got, loads, head)
	}
}

func TestHealthLinksHead4EE430ReturnsNullHead(t *testing.T) {
	loads := 0
	got := healthLinksHead4EE430(func() *int {
		loads++
		return nil
	})
	if got != nil || loads != 1 {
		t.Fatalf("result = %p, loads = %d", got, loads)
	}
}

func TestHealthLinksHead4EE430PreservesLoadFault(t *testing.T) {
	defer func() {
		if got := recover(); got != "load-head" {
			t.Fatalf("panic = %v, want load-head", got)
		}
	}()
	healthLinksHead4EE430(func() string {
		panic("load-head")
	})
}
