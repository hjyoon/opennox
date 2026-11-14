package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/player_marker_movable_original_422140_52e020.json.gz
var playerMarkerMovableOriginal []byte

type playerMarkerMovableOriginalCase struct {
	Kind    string          `json:"kind"`
	Present bool            `json:"present"`
	First   uint32          `json:"first"`
	Second  uint32          `json:"second"`
	Before  string          `json:"before"`
	After   string          `json:"after"`
	Flags   uint32          `json:"flags"`
	Class   uint32          `json:"cls"`
	Result  json.RawMessage `json:"result"`
}

func playerMarkerMovableCases(t *testing.T) []playerMarkerMovableOriginalCase {
	t.Helper()
	if got := fmt.Sprintf("%x", sha256.Sum256(playerMarkerMovableOriginal)); got != "902272844dc78d13dc412bc69d43242fab006da5e920551bd80612ae384f93ed" {
		t.Fatal("sealed original compressed marker/movability fixture changed", got)
	}
	z, err := gzip.NewReader(bytes.NewReader(playerMarkerMovableOriginal))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "63a4ec9a905d9c5e84ae22c13cf60af9e2bb29fb2a0c7e0715a9117006cba7fd" {
		t.Fatal("sealed original marker/movability fixture changed", got)
	}
	var cases []playerMarkerMovableOriginalCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	counts := make(map[string]int)
	for _, c := range cases {
		counts[c.Kind]++
	}
	if len(cases) != 1728 || len(counts) != 2 || counts["marker"] != 128 || counts["movable"] != 1600 {
		t.Fatal("original marker/movability case counts", len(cases), counts)
	}
	return cases
}

func TestPlayerMarkerMovableNativeLayoutCABI(t *testing.T) {
	layout := playerMarkerMovableLayoutC()
	playerType := reflect.TypeOf(server.Player{})
	local, ok := playerType.FieldByName("field3660")
	if !ok {
		t.Fatal("Player local polygon field missing")
	}
	current, ok := playerType.FieldByName("field3664")
	if !ok {
		t.Fatal("Player current polygon field missing")
	}
	if layout.playerSize != unsafe.Sizeof(server.Player{}) || layout.local != local.Offset || layout.current != current.Offset || local.Type.Size() != 4 || current.Type.Size() != 4 {
		t.Fatal("C and Go Player native marker layouts disagree", layout, local, current)
	}
	if layout.objectSize != unsafe.Offsetof(server.Object{}.ScriptPickup)+unsafe.Sizeof(server.Object{}.ScriptPickup) ||
		unsafe.Sizeof(server.Object{}) != layout.objectSize+2*unsafe.Sizeof(uintptr(0)) ||
		layout.class != unsafe.Offsetof(server.Object{}.ObjClass) || layout.flags != unsafe.Offsetof(server.Object{}.ObjFlags) {
		t.Fatal("C object prefix and Go native class/flags layouts disagree", layout)
	}
	want := playerMarkerMovableLayoutNative{4828, 3660, 3664, 772, 8, 16}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		want = playerMarkerMovableLayoutNative{6160, 4956, 4960, 912, 12, 20}
	}
	if layout != want {
		t.Fatal("unexpected native C layouts", layout, want)
	}
}

func TestPlayerMarkerMovableNativeOriginalCABI(t *testing.T) {
	cases := playerMarkerMovableCases(t)
	layout := playerMarkerMovableLayoutC()
	t.Logf("unchanged original player-marker/movability cases=%d, instruction bytes=54, declared services=0", len(cases))
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			playerOwner, freePlayer := alloc.MakeHigh([]byte{}, int(layout.playerSize)+128)
			defer freePlayer()
			objectOwner, freeObject := alloc.MakeHigh([]byte{}, int(unsafe.Sizeof(server.Object{}))+128)
			defer freeObject()
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(&playerOwner[0])) <= math.MaxUint32 || uintptr(unsafe.Pointer(&objectOwner[0])) <= math.MaxUint32) {
				t.Fatal("public C objects must have foreign addresses above4GiB")
			}
			start := 32 + placement
			for index, c := range cases {
				t.Run(fmt.Sprintf("%04d-%s", index, c.Kind), func(t *testing.T) {
					switch c.Kind {
					case "marker":
						for i := range playerOwner {
							playerOwner[i] = 0xa5
						}
						beforeWindow, err := hex.DecodeString(c.Before)
						if err != nil || len(beforeWindow) != 32 {
							t.Fatal("invalid original before window", err)
						}
						afterWindow, err := hex.DecodeString(c.After)
						if err != nil || len(afterWindow) != 32 {
							t.Fatal("invalid original after window", err)
						}
						window := start + int(layout.local) - 4
						copy(playerOwner[window:window+32], beforeWindow)
						// A widened argument must not keep writing PE32 offsets
						// inside the larger native Player.
						if layout.local != 3660 {
							binary.LittleEndian.PutUint32(playerOwner[start+3660:], c.First^0xffffffff)
							binary.LittleEndian.PutUint32(playerOwner[start+3664:], c.Second^0xffffffff)
						}
						player := (*server.Player)(unsafe.Pointer(&playerOwner[start]))
						if player.LocalPolygonID() != c.First || player.CurrentPolygonID() != c.Second {
							t.Fatal("fixture did not reach the native Player fields")
						}
						wantOwner := append([]byte(nil), playerOwner...)
						copy(wantOwner[window:window+32], afterWindow)
						var result *string
						if err := json.Unmarshal(c.Result, &result); err != nil {
							t.Fatal(err)
						}
						var input unsafe.Pointer
						if c.Present {
							input = unsafe.Pointer(player)
							if result == nil || *result != "player" {
								t.Fatal("invalid original result role", result)
							}
						} else if result != nil {
							t.Fatal("original NULL marker result must be NULL", result)
						}
						if got := playerMarkerResetC(input); got != input {
							t.Fatalf("public C reset returned%p, want full pointer%p", got, input)
						}
						if !bytes.Equal(playerOwner, wantOwner) {
							t.Fatal("public C reset changed a nonoutput field or used PE32 marker offsets")
						}
						if player.LocalPolygonID() != binary.LittleEndian.Uint32(afterWindow[4:8]) ||
							player.CurrentPolygonID() != binary.LittleEndian.Uint32(afterWindow[8:12]) {
							t.Fatal("public C reset and Go native marker fields disagree")
						}
					case "movable":
						for i := range objectOwner {
							objectOwner[i] = 0xa5
						}
						var want uint32
						if err := json.Unmarshal(c.Result, &want); err != nil || want > 1 {
							t.Fatal("invalid original movable result", err, want)
						}
						if layout.class != 8 {
							// PE32 class/flags decoys predict the opposite result.
							// They occupy unrelated native fields and stay unchanged.
							oldClass, oldFlags := uint32(0), uint32(0)
							if want == 1 {
								oldClass, oldFlags = 1<<22, 0x8068
							}
							binary.LittleEndian.PutUint32(objectOwner[start+8:], oldClass)
							binary.LittleEndian.PutUint32(objectOwner[start+16:], oldFlags)
						}
						binary.LittleEndian.PutUint32(objectOwner[start+int(layout.class):], c.Class)
						binary.LittleEndian.PutUint32(objectOwner[start+int(layout.flags):], c.Flags)
						before := append([]byte(nil), objectOwner...)
						if got := objectMovableC(unsafe.Pointer(&objectOwner[start])); got != want {
							t.Fatalf("public C movability=%d, want%d, flags=%#x class=%#x", got, want, c.Flags, c.Class)
						}
						if !bytes.Equal(objectOwner, before) {
							t.Fatal("read-only public C movability changed object or canaries")
						}
					default:
						t.Fatal("unknown original operation", c.Kind)
					}
				})
			}
		})
	}
}
