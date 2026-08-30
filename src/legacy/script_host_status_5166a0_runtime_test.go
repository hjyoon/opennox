package legacy

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/noxscript/ns/asm"

	"github.com/opennox/opennox/v1/server"
)

func TestScriptHostStatusNativeLayout5166A0(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantUpdateSize := uintptr(556)
	wantPlayerSize := uintptr(4828)
	wantUpdateData := uintptr(748)
	wantTrade := uintptr(280)
	wantDialog := uintptr(284)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantUpdateSize = 656
		wantPlayerSize = 6160
		wantUpdateData = 872
		wantTrade = 344
		wantDialog = 352
	}
	for _, tc := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(server.Object{}), wantObjectSize},
		{"Object.UpdateData", unsafe.Offsetof(server.Object{}.UpdateData), wantUpdateData},
		{"Player size", unsafe.Sizeof(server.Player{}), wantPlayerSize},
		{"Player.PlayerUnit", unsafe.Offsetof(server.Player{}.PlayerUnit), 2056},
		{"PlayerUpdateData size", unsafe.Sizeof(server.PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Trade70", unsafe.Offsetof(server.PlayerUpdateData{}.Trade70), wantTrade},
		{"PlayerUpdateData.DialogWith", unsafe.Offsetof(server.PlayerUpdateData{}.DialogWith), wantDialog},
		{"Player pointer width", unsafe.Sizeof(server.Player{}.PlayerUnit), unsafe.Sizeof(uintptr(0))},
		{"Trade pointer width", unsafe.Sizeof(server.PlayerUpdateData{}.Trade70), unsafe.Sizeof(uintptr(0))},
		{"Dialog pointer width", unsafe.Sizeof(server.PlayerUpdateData{}.DialogWith), unsafe.Sizeof(uintptr(0))},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %d, want %d", tc.name, tc.got, tc.want)
		}
	}
}

func TestScriptHostStatusNative5166A0PreservesHighPointers(t *testing.T) {
	dialog := &server.Object{}
	trade := &server.TradeSession{}
	update := &server.PlayerUpdateData{DialogWith: dialog, Trade70: trade}
	unit := &server.Object{UpdateData: unsafe.Pointer(update)}
	player := &server.Player{PlayerUnit: unit}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"player": uintptr(unsafe.Pointer(player)),
			"unit":   uintptr(unsafe.Pointer(unit)),
			"update": uintptr(unsafe.Pointer(update)),
			"dialog": uintptr(unsafe.Pointer(dialog)),
			"trade":  uintptr(unsafe.Pointer(trade)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want above 4GiB", name, pointer)
			}
		}
	}

	for _, tc := range []struct {
		name      string
		loadState func(*server.PlayerUpdateData) bool
		want      int32
	}{
		{
			name: "talking",
			loadState: func(got *server.PlayerUpdateData) bool {
				if got != update || got.DialogWith != dialog {
					t.Fatalf("talking update/dialog = %p/%p, want %p/%p", got, got.DialogWith, update, dialog)
				}
				return got.DialogWith != nil
			},
			want: 1,
		},
		{
			name: "trading",
			loadState: func(got *server.PlayerUpdateData) bool {
				if got != update || got.Trade70 != trade {
					t.Fatalf("trading update/trade = %p/%p, want %p/%p", got, got.Trade70, update, trade)
				}
				return got.Trade70 != nil
			},
			want: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var pushed []int32
			if got := scriptHostStatusNative5166A0(
				func() *server.Player { return player },
				tc.loadState,
				func(value int32) { pushed = append(pushed, value) },
			); got != 0 {
				t.Fatalf("result = %d, want canonical zero", got)
			}
			if !reflect.DeepEqual(pushed, []int32{tc.want}) {
				t.Fatalf("pushed = %v, want [%d]", pushed, tc.want)
			}
		})
	}
}

func TestScriptHostStatusNative5166A0ServerQueries(t *testing.T) {
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if NoxScriptIsTalkingNative5166A0(s) || NoxScriptPlayerIsTradingNative5166E0(s) {
		t.Fatal("inactive host reported an active state")
	}

	host := s.Players.ResetInd(server.HostPlayerIndex)
	dialog := &server.Object{}
	trade := &server.TradeSession{}
	update := &server.PlayerUpdateData{DialogWith: dialog, Trade70: trade}
	host.PlayerUnit = &server.Object{UpdateData: unsafe.Pointer(update)}
	if !NoxScriptIsTalkingNative5166A0(s) {
		t.Fatal("native IsTalking did not observe DialogWith")
	}
	if !NoxScriptPlayerIsTradingNative5166E0(s) {
		t.Fatal("native PlayerIsTrading did not observe Trade70")
	}

	update.DialogWith = nil
	update.Trade70 = nil
	if NoxScriptIsTalkingNative5166A0(s) || NoxScriptPlayerIsTradingNative5166E0(s) {
		t.Fatal("native status remained active after both live pointers were cleared")
	}
}

func TestScriptHostStatusNative5166A0RetainsEagerFaults(t *testing.T) {
	for _, tc := range []struct {
		name   string
		player *server.Player
	}{
		{name: "nil unit", player: &server.Player{}},
		{
			name: "nil update",
			player: &server.Player{
				PlayerUnit: &server.Object{},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = scriptHostStatusNative5166A0(
					func() *server.Player { return tc.player },
					func(update *server.PlayerUpdateData) bool { return update.Trade70 != nil },
					func(int32) {},
				)
			}()
			if recovered == nil {
				t.Fatal("expected original eager pointer fault")
			}
		})
	}
}

func TestScriptHostStatusBuiltinsUseNativeRoutes5166A0(t *testing.T) {
	for _, tc := range []struct {
		index asm.Builtin
		got   any
		want  any
	}{
		{asm.BuiltinIsTalking, noxScriptBuiltins[asm.BuiltinIsTalking], noxScriptIsTalkingBuiltin5166A0},
		{asm.BuiltinIsTrading, noxScriptBuiltins[asm.BuiltinIsTrading], noxScriptPlayerIsTradingBuiltin5166E0},
	} {
		got := reflect.ValueOf(tc.got).Pointer()
		want := reflect.ValueOf(tc.want).Pointer()
		if got != want {
			t.Errorf("builtin %d route = %#x, want native %#x", tc.index, got, want)
		}
	}
}
