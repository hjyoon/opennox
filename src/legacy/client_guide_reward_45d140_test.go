package legacy

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	playerlib "github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/server"
)

func TestClientGuideRewardNative45D140UsesNativePlayerAndPreservesUIOrder(t *testing.T) {
	player := new(server.Player)
	player.Info().SetPlayerClass(playerlib.Wizard)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(player)) <= math.MaxUint32 {
		t.Fatalf("player pointer = %p, want native-width address", player)
	}
	player.BeastScrollLvl[7] = 9
	player.BeastScrollLvl[24] = 9

	var events []string
	got := clientGuideRewardNative45D140(24, true, clientGuideRewardHooks45D140{
		currentPlayer: func() *server.Player { return player },
		relatedGuides: func(guide int) []int {
			if guide != 24 {
				t.Fatalf("related guide = %d, want 24", guide)
			}
			return []int{7, 8, 25, 26, 0, -1, clientGuideCount45D140}
		},
		setBookModes: func() { events = append(events, "modes") },
		sortGuideList: func(class uint8) {
			events = append(events, "sort")
			if class != uint8(playerlib.Wizard) {
				t.Fatalf("sort class = %d, want %d", class, playerlib.Wizard)
			}
		},
		findGuidePage: func(guide int) (int, bool) {
			if guide != 24 {
				t.Fatalf("page guide = %d, want 24", guide)
			}
			return 6, true
		},
		hideBook: func(value int) {
			events = append(events, "hide")
			if value != 0 {
				t.Fatalf("hide value = %d, want 0", value)
			}
		},
		moveBookToPage: func(page int) {
			events = append(events, "move")
			if page != 6 {
				t.Fatalf("page = %d, want 6", page)
			}
		},
		openBook: func(value int) {
			events = append(events, "open")
			if value != 0 {
				t.Fatalf("open value = %d, want 0", value)
			}
		},
		showGuideReward: func(guide int) {
			events = append(events, "reward")
			if guide != 24 {
				t.Fatalf("reward guide = %d, want 24", guide)
			}
		},
	})
	if !got {
		t.Fatal("guide reward was rejected")
	}
	for _, guide := range []int{7, 8, 24, 25, 26} {
		if got := player.BeastScrollLvl[guide]; got != 1 {
			t.Errorf("guide %d level = %d, want 1", guide, got)
		}
	}
	if want := []string{"modes", "sort", "hide", "move", "open", "reward"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestClientGuideRewardNative45D140SkipsNotificationUI(t *testing.T) {
	player := new(server.Player)
	var events []string
	hooks := clientGuideRewardHooks45D140{
		currentPlayer: func() *server.Player { return player },
		relatedGuides: func(int) []int { return nil },
		setBookModes:  func() { events = append(events, "modes") },
		sortGuideList: func(uint8) { events = append(events, "sort") },
		findGuidePage: func(int) (int, bool) {
			events = append(events, "find")
			return 0, false
		},
		hideBook:        func(int) { events = append(events, "hide") },
		moveBookToPage:  func(int) { events = append(events, "move") },
		openBook:        func(int) { events = append(events, "open") },
		showGuideReward: func(int) { events = append(events, "reward") },
	}
	if !clientGuideRewardNative45D140(24, false, hooks) {
		t.Fatal("quiet guide reward was rejected")
	}
	if want := []string{"modes", "sort"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("quiet events = %v, want %v", events, want)
	}

	events = nil
	if !clientGuideRewardNative45D140(24, true, hooks) {
		t.Fatal("missing-page guide reward was rejected")
	}
	if want := []string{"modes", "sort", "find"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("missing-page events = %v, want %v", events, want)
	}
}

func TestClientGuideRewardNative45D140RejectsInvalidState(t *testing.T) {
	called := false
	hooks := clientGuideRewardHooks45D140{
		currentPlayer: func() *server.Player { return nil },
		relatedGuides: func(int) []int { called = true; return nil },
		setBookModes:  func() { called = true },
		sortGuideList: func(uint8) { called = true },
	}
	if clientGuideRewardNative45D140(24, true, hooks) || called {
		t.Fatal("nil-player reward performed work")
	}

	player := new(server.Player)
	hooks.currentPlayer = func() *server.Player { return player }
	for _, guide := range []int{-1, 0, clientGuideCount45D140, 255} {
		if clientGuideRewardNative45D140(guide, true, hooks) || called {
			t.Fatalf("invalid guide %d performed work", guide)
		}
	}
}

func TestClientGuideRelationsNative45D140UsesPackedIDsAndWidePointerSlots(t *testing.T) {
	InitBlobData()
	if got, want := clientGuideRelationsNative45D140(24), []int{7, 8, 25, 26}; !reflect.DeepEqual(got, want) {
		t.Fatalf("guide 24 relations = %v, want %v", got, want)
	}
	if got := clientGuideRelationsNative45D140(23); len(got) != 0 {
		t.Fatalf("guide 23 relations = %v, want none", got)
	}
}

func TestClientGuideRewardExport45D140PreservesCABIArguments(t *testing.T) {
	old := clientGuideRewardCall45D140
	t.Cleanup(func() { clientGuideRewardCall45D140 = old })
	var gotGuide, gotNotify int32
	clientGuideRewardCall45D140 = func(guide, notify int32) {
		gotGuide, gotNotify = guide, notify
	}
	clientGuideRewardExportCall45D140(math.MinInt32, math.MaxInt32)
	if gotGuide != math.MinInt32 || gotNotify != math.MaxInt32 {
		t.Fatalf("C ABI args = %d/%d, want %d/%d", gotGuide, gotNotify, int32(math.MinInt32), int32(math.MaxInt32))
	}
}
