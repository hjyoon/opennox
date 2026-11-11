package opennox

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func TestTrollDeathActualCProduction54A270(t *testing.T) {
	modes := []string{"cold", "warm", "audio-replaces-update", "allocation-failure"}
	mode := os.Getenv("OPENNOX_TROLL_PRODUCTION_CHILD")
	if mode == "" {
		for _, mode := range modes {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.run=^TestTrollDeathActualCProduction54A270$", "-test.v")
				cmd.Env = append(os.Environ(), "OPENNOX_TROLL_PRODUCTION_CHILD="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("actual default C %s: %v\n%s", mode, err, out)
				if err != nil || !bytes.Contains(out, []byte("TROLL_ACTUAL_DEFAULT_C_PRODUCTION")) {
					t.Fatal("actual default C root-server composition", err)
				}
			})
		}
		return
	}
	if !slices.Contains(modes, mode) {
		t.Fatal("unknown production child", mode)
	}
	data, err := os.ReadFile("server/testdata/troll_death_native_original_54a270_pc53_tozero.json.gz")
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != "410eb49fb7016edd692668b5127a2f855c2c3bbf9de6c89193eae3ff985dd5b6" {
		t.Fatal("frozen original", err)
	}
	z, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name      string
		Durations [2]uint32
	}
	if err = json.Unmarshal(decoded, &rows); err != nil {
		t.Fatal(err)
	}
	var want uint32
	found := false
	for _, row := range rows {
		if row.Name == "historical-input-stock-style-tie-37.5" {
			want = row.Durations[0]
			found = true
			break
		}
	}
	if !found {
		t.Fatal("independent original lifetime result")
	}
	handles.Init()
	t.Cleanup(handles.Release)
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	oldFlags, oldGameplay := noxflags.GetGame(), noxflags.GetGamePlay()
	noxflags.ResetGame()
	noxflags.SetGamePlay(noxflags.GameplayFlag1)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
		noxflags.UnsetGamePlay(noxflags.GetGamePlay())
		noxflags.SetGamePlay(oldGameplay)
	})
	s.SetFrame(1400)
	s.SetTickRate(30)
	s.spells.duration.Init(s)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if !s.Objs.InitHigh(8) {
		t.Fatal("real object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.Init() == 0 {
		t.Fatal("native wall grid")
	}
	t.Cleanup(s.Walls.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	name := sound.SoundTrollFlatus.String()
	wire := binary.LittleEndian.AppendUint32(nil, 1)
	wire = append(wire, byte(len(name)))
	wire = append(wire, name...)
	wire = binary.LittleEndian.AppendUint16(wire, 8)
	wire = append(wire, 0)
	wire = binary.LittleEndian.AppendUint16(wire, 40)
	wire = append(wire, 0, 0, 0, 0, 1, 's', 0)
	file := luaCompatAudioFile{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(file) || file.Len() != 0 || s.Audio.Field12(sound.SoundTrollFlatus) != 1 {
		t.Fatal("real Troll AUD descriptor")
	}
	s.Audio.Reset()
	dir, oldData := t.TempDir(), datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	if err = os.WriteFile(filepath.Join(dir, "gamedata.yml"), []byte("SmallToxicCloudLifetime: 1.25\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = s.Balance.Read(); err != nil || s.Balance.Float("SmallToxicCloudLifetime") != 1.25 {
		t.Fatal("real binary64 balance service", err)
	}
	cache := memmap.PtrUint32(0x5D4594, 2491584)
	oldCache := *cache
	*cache = 0
	t.Cleanup(func() { *cache = oldCache })
	if mode == "allocation-failure" {
		value, ok := legacy.TrollDeathActualDefaultC54A270(nil)
		if !ok || value != 1 || *cache != 0 || s.Objs.Pending != nil {
			t.Fatal("real allocation failure must not read nil source", value, ok, *cache)
		}
		n := 0
		s.Audio.EachEvent(func(*server.AudioEvent) { n++ })
		if n != 0 {
			t.Fatal("allocation failure emitted sound")
		}
		t.Log("TROLL_ACTUAL_DEFAULT_C_PRODUCTION mode=allocation-failure original-return1-no-source-read=true")
		return
	}
	if err = s.Types.ReadObjectType(&things.Thing{Name: "SmallToxicCloud", OnUpdate: &things.ProcFunc{Name: "SmallToxicCloudUpdate"}}); err != nil {
		t.Fatal(err)
	}
	id := uint32(s.Types.ByID("SmallToxicCloud").Ind())
	if id == 0 {
		t.Fatal("nonzero type ID")
	}
	if mode == "warm" {
		*cache = id
	}
	unit := s.Objs.NewObject(&server.ObjectType{})
	if unit == nil {
		t.Fatal("native source")
	}
	unit.ObjFlags = object.FlagActive | object.FlagEnabled
	unit.PosVec = types.Ptf(300, 300)
	unitFrame := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit))))
	replacement, freeReplacement := alloc.NewHigh(server.ToxicCloudUpdateData{Duration: -456})
	t.Cleanup(freeReplacement)
	replacement.Duration = -456
	var cloud *server.Object
	var originalData *server.ToxicCloudUpdateData
	var cloudFrame []byte
	audioCalls := 0
	s.Audio.OnSound(func(id sound.ID, kind int, source *server.Object, pos types.Pointf) {
		audioCalls++
		cloud = s.Objs.Pending
		if id != sound.SoundTrollFlatus || kind != 0 || source != unit || pos != unit.PosVec || cloud == nil || cloud.ObjOwner != unit || cloud.PosVec != unit.PosVec || !cloud.Flags().Has(object.FlagActive|object.FlagPending) || cloud.UpdateData == nil {
			t.Fatal("actual create-before-audio/owner/native position", id, kind, source, cloud)
		}
		originalData = (*server.ToxicCloudUpdateData)(cloud.UpdateData)
		if originalData.Duration != 0 {
			t.Fatal("duration stored before audio")
		}
		cloudFrame = bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(cloud)), int(unsafe.Sizeof(*cloud))))
		if mode == "audio-replaces-update" {
			cloud.UpdateData = unsafe.Pointer(replacement)
			at := int(unsafe.Offsetof(server.Object{}.UpdateData))
			n := int(unsafe.Sizeof(uintptr(0)))
			copy(cloudFrame[at:at+n], unsafe.Slice((*byte)(unsafe.Pointer(&cloud.UpdateData)), n))
		}

	})
	value, ok := legacy.TrollDeathActualDefaultC54A270(unit)
	if !ok || value != 1 || audioCalls != 1 || *cache != id || cloud == nil || originalData == nil || uint32(originalData.Duration) != want || replacement.Duration != -456 {
		t.Fatal("actual default C lifetime/cache/cached-update", value, ok, audioCalls, *cache, originalData)
	}
	if cloud.UpdateData != unsafe.Pointer(originalData) && (mode != "audio-replaces-update" || cloud.UpdateData != unsafe.Pointer(replacement)) {
		t.Fatal("live pointer replacement identity")
	}
	// SetOwner legitimately links the new cloud into its owner's child list.
	// Capture the independently observable owner-link fields changed by the
	// real world CreateObjectAt, leaving every other source byte guarded.
	if unit.Field129 != cloud {
		t.Fatal("real owner child-list head")
	}
	off := unsafe.Offsetof(unit.Field129)
	n := int(unsafe.Sizeof(uintptr(0)))
	copy(unitFrame[int(off):int(off)+n], unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(unit), off)), n))

	if !bytes.Equal(unitFrame, unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))) || !bytes.Equal(cloudFrame, unsafe.Slice((*byte)(unsafe.Pointer(cloud)), int(unsafe.Sizeof(*cloud)))) {
		t.Fatal("whole source/cloud records outside declared world effects")
	}
	var audio []*server.AudioEvent
	s.Audio.EachEvent(func(ev *server.AudioEvent) { audio = append(audio, ev) })
	if len(audio) != 1 || audio[0].Sound != sound.SoundTrollFlatus || audio[0].Obj != unit || audio[0].Pos != unit.PosVec || audio[0].Kind != 0 || audio[0].Code != 0 {
		t.Fatal("real allocated Troll AudioEvent")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(cloud), unsafe.Pointer(originalData), unsafe.Pointer(replacement), unsafe.Pointer(audio[0])} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatal("real production pointer below4GiB", p)
			}
		}
	}
	// FreeObject must see its allocated update record, not the observer's
	// temporary replacement; restore only after all lifetime/guard assertions.
	cloud.UpdateData = unsafe.Pointer(originalData)
	t.Cleanup(func() { s.Objs.Pending = nil; s.Objs.FreeObject(cloud); s.Objs.FreeObject(unit) })
	t.Logf("TROLL_ACTUAL_DEFAULT_C_PRODUCTION mode=%s source=%p cloud=%p cached-update=%p duration=%d real-root-create-owner-AUD-balance-FPS=true", mode, unit, cloud, originalData, originalData.Duration)
}
