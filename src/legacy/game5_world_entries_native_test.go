package legacy

import (
	"bytes"
	"encoding/binary"
	"image"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestGame5SecretWallPublicCStateAndAudio(t *testing.T) {
	s := npcReflectLegacyServer4E17B0(t)
	if s.Walls.InitHigh() == 0 {
		t.Fatal("wall owner")
	}
	defer s.Walls.Free()
	// A valid empty-sprite wall definition with the stock secret-wall sound.
	var definition bytes.Buffer
	put := func(v uint32) {
		if err := binary.Write(&definition, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	string8 := func(v string) { definition.WriteByte(byte(len(v))); definition.WriteString(v) }
	align64 := func() {
		for definition.Len()%8 != 0 {
			definition.WriteByte(0)
		}
		definition.Write(make([]byte, 8))
	}
	put(0)
	string8("Game5Wall")
	put(0)
	put(0)
	put(0)
	definition.Write([]byte{0, 100})
	align64()
	string8("SecretWallOpen")
	string8("SecretWallClose")
	string8("")
	definition.WriteByte(0)
	for i := 0; i < 15; i++ {
		align64()
	}
	put(0x454e4420)
	data, _ := alloc.MakeHigh([]byte{}, definition.Len())
	copy(data, definition.Bytes())
	mf := binfile.NewMemFile(unsafe.Pointer(&data[0]), len(data))
	if err := s.Walls.ReadWall(mf); err != nil {
		mf.Free()
		t.Fatal(err)
	}
	mf.Free()
	wall := s.Walls.CreateAtGrid(image.Pt(37, 91))
	if wall == nil {
		t.Fatal("wall allocation")
	}
	wall.Flags4 = 4
	wall.Tile1 = 0
	secret, free := alloc.NewHigh(server.SecretWall{})
	defer free()
	wall.SetData(unsafe.Pointer(secret))
	defer wall.SetData(nil)
	owner := game5HighObject(t, 17)
	owner.ObjClass = object.ClassPlayer
	s.Audio.Init(s)
	defer s.Audio.Free()
	id := sound.SoundSecretWallOpen
	if !s.Audio.ReadAVNT(&doorSoundFile53AC50{name: id.String(), playerAttackHammerAudioFile538960: playerAttackHammerAudioFile538960{data: []byte{7, 1, 'x', 0, 0}}}) {
		t.Fatal("audio definition")
	}
	for _, grid := range [][2]int32{{48, 62}, {math.MaxInt32, math.MinInt32}, {186737709, -186737709}} {
		secret.X, secret.Y = grid[0], grid[1]
		secret.Flags = 2
		secret.State = 1
		secret.OpenDelay = 255
		expected := *secret
		expected.State = 4
		expected.OpenDelay = 0
		center := func(v int32) float32 { return float32(float64(int32(uint32(v)*23)) + 11.5) }
		want := types.Ptf(center(grid[0]), center(grid[1]))
		s.Audio.Reset()
		game5SecretWallTouchC([2]int32{37, 91}, owner)
		game5SecretWallTouchC([2]int32{37, 91}, owner)
		if *secret != expected {
			t.Fatal("native secret-wall stores", *secret, expected)
		}
		n := 0
		s.Audio.EachEvent(func(ev *server.AudioEvent) {
			n++
			if ev.Sound != id || ev.Pos != want || ev.Obj != nil || ev.Kind != 0 || ev.Code != 0 {
				t.Fatal("actual audio queue", *ev, want)
			}
		})
		if n != 1 {
			t.Fatal("repeated contact sound count", n)
		}
	}
}

func TestGame5CursorNativeProjectionAndPublicC(t *testing.T) {
	s := npcReflectLegacyServer4E17B0(t)
	s.Map.Init()
	defer game5CursorClearFixture()
	defer s.Map.Free()
	source := game5HighObject(t, 1)
	source.ObjClass = object.ClassPlayer
	update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
	defer freeUpdate()
	player, freePlayer := alloc.NewHigh(server.Player{})
	defer freePlayer()
	update.Player = player
	source.UpdateData = unsafe.Pointer(update)
	defer func() { source.UpdateData = nil; update.Player = nil }()
	player.SetCursorPos(image.Pt(100, 80))
	if got := game5CursorFindC(source); got != nil {
		t.Fatal("empty native cursor query", got)
	}
	game5CursorScanC(source, &types.Pointf{X: 100, Y: 80}) // Source is excluded before vision lookup.
	obj := game5HighObject(t, 2)
	obj.PosVec = types.Ptf(100, 100)
	obj.ZVal = 20
	obj.Shape.Kind = 2
	obj.Shape.Circle.R = 10
	for _, tc := range []struct {
		point types.Pointf
		want  bool
	}{
		{types.Ptf(100, 80), true}, {types.Ptf(109, 80), true}, {types.Ptf(110, 80), false},
		{types.Ptf(100, 89), true}, {types.Ptf(100, 90), false}, {types.Ptf(100, 71), true},
		{types.Ptf(100, 70), false}, {types.Ptf(109, 89), false},
	} {
		if got := game5CursorContainsForTest(obj, &tc.point); got != tc.want {
			t.Fatal("projected circle boundary", tc.point, got, tc.want)
		}
	}
	obj.Shape.Kind = 3
	obj.Shape.Box.W = 20
	obj.Shape.Box.H = 20
	obj.Shape.Box.Calc()
	center := types.Ptf(100, 80)
	if !game5CursorContainsForTest(obj, &center) {
		t.Fatal("native box projection")
	}
	obj.ObjFlags = 0x20
	game5CursorScanC(obj, &center)
	if game5Cursor.result != nil {
		t.Fatal("disabled cursor candidate selected")
	}
}
