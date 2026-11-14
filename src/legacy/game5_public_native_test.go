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
	"runtime"
	"strconv"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/game5_public_original_57a160_57b810.json.gz
var game5PublicOriginalFixture []byte

//go:embed testdata/game5_public_native_x87_57b810.json.gz
var game5PublicNativeX87Fixture []byte

type game5PublicX87Sidecar struct {
	OriginalFixtureSHA256   string            `json:"original_fixture_sha256"`
	NativeReflectCases      int               `json:"native_reflect_cases"`
	OriginalAgreementCases  int               `json:"original_agreement_cases"`
	SignalingNaNCorrections map[string]string `json:"signaling_NaN_corrections"`
}

type game5PublicOriginalCase struct {
	Kind       string          `json:"kind"`
	Present    bool            `json:"present"`
	Flags      uint32          `json:"flags"`
	Field      uint32          `json:"field"`
	Normal     [2]uint32       `json:"normal"`
	AliasDelta *int            `json:"alias_delta"`
	Before     string          `json:"before"`
	After      string          `json:"after"`
	Result     json.RawMessage `json:"result"`
}

func TestGame5PublicNativePlayerLayoutCABI(t *testing.T) {
	size, offset := game5PublicPlayerLayoutC()
	if size != unsafe.Sizeof(server.Player{}) || offset != unsafe.Offsetof(server.Player{}.Field3680) || unsafe.Sizeof(server.Player{}.Field3680) != 4 {
		t.Fatalf("native Player layout C=%d/%d Go=%d/%d", size, offset, unsafe.Sizeof(server.Player{}), unsafe.Offsetof(server.Player{}.Field3680))
	}
}

func TestGame5PublicNativeOriginalCABI(t *testing.T) {
	if got := fmt.Sprintf("%x", sha256.Sum256(game5PublicOriginalFixture)); got != "c963174d95c46a4128671035082c2a2ee2bec88514ab5cd3699539015f188bba" {
		t.Fatal("sealed original fixture compressed bytes changed", got)
	}
	z, err := gzip.NewReader(bytes.NewReader(game5PublicOriginalFixture))
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != "a11e503153189392a24411186cc875b9bd6dec8e0b2f4923d618d7fdab27e339" {
		t.Fatal("sealed original fixture bytes changed", got)
	}
	var cases []game5PublicOriginalCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 3036 {
		t.Fatal("original case count", len(cases))
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(game5PublicNativeX87Fixture)); got != "892a522a900ecda2b94cfdf421c66e671e4635ec86bd13f784633e8762d907e5" {
		t.Fatal("sealed independent x87 fixture changed", got)
	}
	z, err = gzip.NewReader(bytes.NewReader(game5PublicNativeX87Fixture))
	if err != nil {
		t.Fatal(err)
	}
	nativeData, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(nativeData)); got != "18993f065813ffc73ca4449d2a6082a058e2519e75ae1b6a4556c36c3ab063e4" {
		t.Fatal("sealed independent x87 decoded bytes changed", got)
	}
	var x87 game5PublicX87Sidecar
	if err := json.Unmarshal(nativeData, &x87); err != nil {
		t.Fatal(err)
	}
	if x87.OriginalFixtureSHA256 != "a11e503153189392a24411186cc875b9bd6dec8e0b2f4923d618d7fdab27e339" || x87.NativeReflectCases != 2940 || x87.OriginalAgreementCases != 2415 || len(x87.SignalingNaNCorrections) != 525 {
		t.Fatal("independent x87 scope differs from the sealed original corpus")
	}
	for key, corrected := range x87.SignalingNaNCorrections {
		index, err := strconv.Atoi(key)
		if err != nil || index < 0 || index >= len(cases) || cases[index].Kind != "reflect" {
			t.Fatal("invalid independent x87 row", key)
		}
		a, err := hex.DecodeString(cases[index].After)
		if err != nil {
			t.Fatal(err)
		}
		b, err := hex.DecodeString(corrected)
		if err != nil || len(a) != 32 || len(b) != 32 || bytes.Equal(a, b) {
			t.Fatal("invalid independent x87 owner", key)
		}
		for word := 0; word < 8; word++ {
			original, native := binary.LittleEndian.Uint32(a[word*4:]), binary.LittleEndian.Uint32(b[word*4:])
			if original == native {
				continue
			}
			if (word != 2 && word != 3) || original&0x7f800000 != 0x7f800000 || original&0x007fffff == 0 || original&0x00400000 != 0 || native != original|0x00400000 {
				t.Fatal("independent x87 difference exceeds signaling-NaN quieting", key, word)
			}
		}
	}
	nativeX87 := game5PublicX87AvailableC()
	if nativeX87 != (runtime.GOARCH == "386" || runtime.GOARCH == "amd64") {
		t.Fatal("native x87 reference architecture mismatch")
	}
	t.Logf("sealed original cases=%d; independent x87 NaN cases=%d; local x87 cross-check=%t", len(cases), len(x87.SignalingNaNCorrections), nativeX87)
	previousFlags := noxflags.GetGame()
	defer func() { noxflags.ResetGame(); noxflags.SetGame(previousFlags) }()
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			playerBytes, freePlayer := alloc.MakeHigh([]byte{}, int(unsafe.Sizeof(server.Player{}))+128)
			defer freePlayer()
			vectorBytes, freeVector := alloc.MakeHigh([]byte{}, 128)
			defer freeVector()
			normalBytes, freeNormal := alloc.MakeHigh([]byte{}, 128)
			defer freeNormal()
			for _, b := range [][]byte{playerBytes, vectorBytes, normalBytes} {
				for i := range b {
					b[i] = 0xa5
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&b[0])) <= math.MaxUint32 {
					t.Fatal("native owner must be foreign and above the PE32 address range")
				}
			}
			start := 32 + placement
			player := (*server.Player)(unsafe.Pointer(&playerBytes[start]))
			owner := vectorBytes[start : start+32]
			velocity := unsafe.Pointer(&owner[8])
			for i, c := range cases {
				t.Run(fmt.Sprintf("%04d-%s", i, c.Kind), func(t *testing.T) {
					switch c.Kind {
					case "chat":
						noxflags.ResetGame()
						noxflags.SetGame(noxflags.GameFlag(c.Flags))
						if unsafe.Offsetof(player.Field3680) != 3680 {
							binary.LittleEndian.PutUint32(playerBytes[start+3680:], c.Field^8)
						}
						player.Field3680 = c.Field
						before := append([]byte(nil), playerBytes...)
						var arg unsafe.Pointer
						if c.Present {
							arg = unsafe.Pointer(player)
						}
						var want int32
						if err := json.Unmarshal(c.Result, &want); err != nil {
							t.Fatal(err)
						}
						if got := game5PublicCanTalkC(arg); got != want || uint32(noxflags.GetGame()) != c.Flags {
							t.Fatalf("actual C player entry: got%d want%d native player=%p flags=%08x", got, want, arg, c.Flags)
						}
						if !bytes.Equal(playerBytes, before) {
							t.Fatal("read-only player entry changed its owner or canaries")
						}
					case "reflect":
						before, err := hex.DecodeString(c.Before)
						if err != nil || len(before) != 32 {
							t.Fatal("original owner bytes", err)
						}
						want, err := hex.DecodeString(c.After)
						if corrected, ok := x87.SignalingNaNCorrections[strconv.Itoa(i)]; ok {
							want, err = hex.DecodeString(corrected)
						}
						if err != nil || len(want) != 32 {
							t.Fatal("original final owner bytes", err)
						}
						copy(owner, before)
						for i, v := range c.Normal {
							binary.LittleEndian.PutUint32(normalBytes[start+8+4*i:], v)
						}
						normalBefore := append([]byte(nil), normalBytes...)
						normal := unsafe.Pointer(&normalBytes[start+8])
						if c.AliasDelta != nil {
							normal = unsafe.Add(velocity, *c.AliasDelta)
						}
						if nativeX87 {
							var native [32]byte
							if !game5PublicX87ReflectC(normal, velocity, &native, c.AliasDelta) || !bytes.Equal(native[:], want) {
								t.Fatalf("local x87 sequence disagrees with sealed references: got=%x want=%x", native, want)
							}
						}
						if got := game5PublicReflectC(normal, velocity); got != uintptr(velocity) {
							t.Fatalf("C return truncated velocity: %x != %p", got, velocity)
						}
						if !bytes.Equal(owner, want) {
							t.Fatalf("original CPU reflection differs: normal=%08x alias=%v before=%x got=%x want=%x", c.Normal, c.AliasDelta, before, owner, want)
						}
						if !bytes.Equal(normalBytes, normalBefore) {
							t.Fatal("read-only collision normal owner changed")
						}
						for j, v := range vectorBytes {
							if (j < start || j >= start+32) && v != 0xa5 {
								t.Fatal("velocity owner outer canary changed")
							}
						}
					default:
						t.Fatal("unknown original case", c.Kind)
					}
				})
			}
		})
	}
}
