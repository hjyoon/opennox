package opennox

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func TestSimplePolypDeathActualCProduction54CAE0(t *testing.T) {
	mode := os.Getenv("OPENNOX_SIMPLE_POLYP_PRODUCTION_CHILD")
	if mode == "" {
		for _, mode := range []string{"egg", "egg-live-flags", "potion", "polyp-cold", "polyp-warm", "polyp-allocation-failure", "polyp-nil-allocation-failure"} {
			t.Run(mode, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSimplePolypDeathActualCProduction54CAE0$", "-test.v")
				cmd.Env = append(os.Environ(), "OPENNOX_SIMPLE_POLYP_PRODUCTION_CHILD="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("actual default C %s: %v\n%s", mode, err, out)
				if ctx.Err() != nil || err != nil || !bytes.Contains(out, []byte("SIMPLE_POLYP_ACTUAL_DEFAULT_C_PRODUCTION")) {
					t.Fatal("actual default root-server services", err)
				}
			})
		}
		return
	}
	data, err := os.ReadFile("server/testdata/shared_cloud_callers_precision_original.json")
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != "84cdd8f4e1497119c8347a5517850ac16bdee23b08ec0da2def439e1669e075d" {
		t.Fatal("pre-existing complete original Polyp precision seal", err)
	}
	var reference struct {
		Polyp []struct {
			Input       struct{ Name string }
			OldDuration int32 `json:"old_duration"`
		}
	}
	if err = json.Unmarshal(data, &reference); err != nil {
		t.Fatal(err)
	}
	var expectedDuration int32
	found := false
	for _, row := range reference.Polyp {
		if row.Input.Name == "stock-style-37-half/none" {
			expectedDuration, found = row.OldDuration, true
		}
	}
	if !found {
		t.Fatal("frozen original 1.25 times signed FPS30 result")
	}
	legacy.InitBlobData()
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
		t.Fatal("actual object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.Init() == 0 {
		t.Fatal("actual wall grid")
	}
	t.Cleanup(s.Walls.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	wire := binary.LittleEndian.AppendUint32(nil, 3)
	for _, id := range []sound.ID{764, 284, 753} {
		name := id.String()
		wire = append(wire, byte(len(name)))
		wire = append(wire, name...)
		wire = append(wire, 8, 0, 0, 40, 0, 0, 0, 0, 0, 1, 's', 0)
	}
	file := luaCompatAudioFile{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(file) || file.Len() != 0 {
		t.Fatal("actual AUD descriptor reader")
	}
	s.Audio.Reset()
	dir, oldData := t.TempDir(), datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	if err = os.WriteFile(filepath.Join(dir, "gamedata.yml"), []byte("ToxicCloudLifetime: 1.25\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = s.Balance.Read(); err != nil || s.Balance.Float("ToxicCloudLifetime") != 1.25 {
		t.Fatal("actual balance reader", err)
	}
	cache := memmap.PtrUint32(0x5D4594, 2491672)
	oldCache := *cache
	*cache = 0
	t.Cleanup(func() { *cache = oldCache })
	polyp := strings.HasPrefix(mode, "polyp-")
	allocationFailure := strings.Contains(mode, "allocation-failure")
	if polyp && !allocationFailure {
		if err = s.Types.ReadObjectType(&things.Thing{Name: "ToxicCloud", OnUpdate: &things.ProcFunc{Name: "ToxicCloudUpdate"}}); err != nil {
			t.Fatal(err)
		}
		if mode == "polyp-warm" {
			*cache = uint32(s.Types.ByID("ToxicCloud").Ind())
		}
	}
	var unit *server.Object
	if mode != "polyp-nil-allocation-failure" {
		unit = s.Objs.NewObject(&server.ObjectType{})
		if unit == nil {
			t.Fatal("actual native source")
		}
		unit.ObjFlags, unit.PosVec = object.FlagActive|object.FlagEnabled, types.Ptf(300, 300)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
			t.Fatal("actual source must exceed4GiB")
		}
	}
	var sourceFrame []byte
	if unit != nil {
		sourceFrame = bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit))))
	}
	index, expectedSound := 0, sound.ID(764)
	if polyp {
		index, expectedSound = 1, 284
	} else if mode == "potion" {
		index, expectedSound = 2, 753
	}
	sounds := 0
	s.Audio.OnSound(func(id sound.ID, kind int, source *server.Object, position types.Pointf) {
		if source != unit || id != expectedSound || kind != 0 || unit == nil || position != unit.PosVec || unit.Flags().Has(object.FlagDestroyed) {
			t.Fatal("actual audio precedes source flags/deletion", id, source, position)
		}
		sounds++
		if !bytes.Equal(sourceFrame, unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))) {
			t.Fatal("source changed before real audio")
		}
		if polyp && !allocationFailure {
			cloud := s.Objs.Pending
			if cloud == nil || cloud.UpdateData == nil || cloud.ObjOwner != nil || cloud.PosVec != unit.PosVec || (*server.ToxicCloudUpdateData)(cloud.UpdateData).Duration != expectedDuration {
				t.Fatal("actual cloud duration/placement before audio", cloud)
			}
		}
		if mode == "egg-live-flags" {
			unit.ObjFlags = 0xabcdef81
		}
	})
	value, ok := legacy.SimplePolypDeathActualDefaultC54CAE0(index, unit)
	if !ok {
		t.Fatal("actual typed C entry with default services required")
	}
	if mode == "polyp-nil-allocation-failure" {
		if sounds != 0 || s.Objs.DeletedList != nil || s.Objs.Pending != nil || *cache != 0 {
			t.Fatal("actual failed allocation skips nil source, nil audio and nil deletion")
		}
	} else {
		if sounds != 1 {
			t.Fatal("actual audio count", sounds)
		}
		if index == 0 {
			expectedFlags := object.FlagActive | object.FlagEnabled | object.FlagNoCollide
			if mode == "egg-live-flags" {
				expectedFlags = 0xabcdefc1
			}
			if value != uint32(expectedFlags) || unit.ObjFlags != expectedFlags || s.Objs.DeletedList != nil {
				t.Fatal("actual full flags return/reload and no deletion", value, unit.ObjFlags)
			}
			binary.LittleEndian.PutUint32(sourceFrame[unsafe.Offsetof(unit.ObjFlags):], uint32(expectedFlags))
			if !bytes.Equal(sourceFrame, unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))) {
				t.Fatal("whole actual ImpEgg source outside flags")
			}
		} else if !unit.Flags().Has(object.FlagDestroyed) || s.Objs.DeletedList != unit {
			t.Fatal("actual source delayed deletion after audio", unit.ObjFlags, s.Objs.DeletedList)
		}
		if polyp && !allocationFailure {
			cloud := s.Objs.Pending
			if cloud == nil || cloud.ObjectTypeC().ID() != "ToxicCloud" || cloud.ObjNext != nil || *cache != uint32(cloud.TypeInd) {
				t.Fatal("actual one-cloud construction and live type cache", cloud, *cache)
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(cloud)) <= math.MaxUint32 || uintptr(cloud.UpdateData) <= math.MaxUint32) {
				t.Fatal("actual cloud/update must exceed4GiB")
			}
		} else if s.Objs.Pending != nil || *cache != 0 {
			t.Fatal("unexpected cloud or cache change")
		}
	}
	t.Logf("SIMPLE_POLYP_ACTUAL_DEFAULT_C_PRODUCTION mode=%s audio=%d original-duration=%d source=%p default-world=true", mode, sounds, expectedDuration, unit)
}
