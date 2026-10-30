package legacy

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Real removal, cleanup, Spark/Burn and Raise/Pop/Push/HP services run. The
// isolated server has no peers, owner or scorch type; DelayedDelete is an
// observed boundary. This does not claim FX wire delivery or actual deletion.
type deadUpdateNativeServices544EC0 struct {
	Server
	deleted []*server.Object
}

func (s *deadUpdateNativeServices544EC0) DelayedDelete(unit *server.Object) {
	s.deleted = append(s.deleted, unit)
}

func TestMonsterDeadUpdateNativeComposition544EC0(t *testing.T) {
	const child = "NOX_MONSTER_DEAD_UPDATE_COMPOSITION_CHILD"
	if os.Getenv(child) == "" {
		modes := []string{"cleanup-0", "cleanup-1", "cleanup-2", "cleanup-16", "cleanup-negative", "ordinary-retain", "ordinary-delete"}
		for _, kind := range []string{"zombie", "vile"} {
			for _, branch := range []string{"boundary", "wait", "fire", "raise", "stay", "noenemy", "wrap"} {
				modes = append(modes, kind+"-"+branch)
			}
		}
		for _, mode := range modes {
			t.Run(mode, func(t *testing.T) {
				bin, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(bin, "-test.run=^TestMonsterDeadUpdateNativeComposition544EC0$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual public C DeadUpdate composition: %v", err)
				}
			})
		}
		return
	}
	mode := os.Getenv(child)
	parts := strings.Split(mode, "-")
	if len(parts) != 2 {
		t.Fatal("unknown frozen composition")
	}
	s, unit, update, missile := monsterMainBlockLegacyFixture547210(t)
	view := s.TypePredicateRuntime534A10()
	view.Write(1, 100)
	view.Write(2, 200)
	unit.TypeInd = 300
	query, status, frame, deadFrame, duration, wp, seen, delete := uint32(0), uint32(0), uint32(13), uint32(10), uint32(2), uint32(2), uint8(3), uint32(0)
	enemy, cleanup := true, parts[0] == "cleanup"
	if parts[0] == "zombie" {
		unit.TypeInd = 100
		query = 1
	} else if parts[0] == "vile" {
		unit.TypeInd = 200
		query = 1
	}
	if cleanup {
		if parts[1] == "negative" {
			wp, seen = 0xffffffff, 1
		} else {
			n, err := strconv.Atoi(parts[1])
			if err != nil {
				t.Fatal(err)
			}
			wp, seen = uint32(n), uint8(n)
		}
	} else if parts[0] == "ordinary" {
		seen = 14
		if parts[1] == "delete" {
			delete = 1
		} else if parts[1] != "retain" {
			t.Fatal("unknown ordinary branch")
		}
	} else {
		switch parts[1] {
		case "boundary":
			frame = 12
		case "wait":
			frame = 0
			deadFrame = 0xffffffff
			duration = 0xffffffff
		case "fire":
			status = 0x80000
		case "raise":
		case "stay":
			status = 0x100000
		case "noenemy":
			enemy = false
		case "wrap":
			frame = 1
			deadFrame = 0xfffffffe
		default:
			t.Fatal("unknown zombie branch")
		}
	}
	if got := uint32(zombieEntryCProbe534A10(1, unit)); got != query {
		t.Fatal("real native IsZombie", got, query)
	}
	definition, free := alloc.New(server.MonsterDef{})
	defer free()
	*definition = server.MonsterDef{StatusFlags92: object.MonsterStatus(delete)}
	waypoints, freeWaypoints := alloc.Make([]server.Waypoint{}, 16)
	defer freeWaypoints()
	update.MonsterDef = definition
	update.Field74 = wp
	update.Field282_1 = seen
	update.Field91 = 0x12345678
	update.Field523_2, update.Field523_3 = 0x34, 0xd5
	update.Stamina, update.Field282_2 = 0xa5, 0xa5a5
	update.StatusFlags = object.MonsterStatus(status)
	update.Field137, update.Field123 = deadFrame, duration
	update.CurrentEnemy = nil
	if enemy {
		update.CurrentEnemy = missile
	}
	for i := range update.Waypoints {
		update.Waypoints[i] = &waypoints[i]
		update.SeenEnemies[i] = missile
	}
	unit.Update = unit.CObj()
	update.AIStackInd = 0
	update.AIStack[0] = server.AIStackItem{Action: uint32(ai.ACTION_DEAD)}
	unit.ObjFlags = object.FlagActive | object.FlagEnabled | object.FlagDead | object.FlagAllowOverlap | object.FlagShort | object.FlagNoCollide
	unit.HealthData.Cur = 7
	unit.HealthData.Field2 = 9
	s.SetFrame(frame)
	s.AI.StackChanged = false
	if query == 0 && !cleanup {
		s.Objs.AddToUpdatable(unit)
	}
	if s.Types.ByID("ScorchMarkFloorMediumA") != nil || unit.ObjOwner != nil || unit.Obj130 != nil || s.Players.FirstUnit() != nil {
		t.Fatal("isolated native service boundaries changed")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff || uintptr(unsafe.Pointer(definition)) <= 0xffffffff || uintptr(unsafe.Pointer(&waypoints[0])) <= 0xffffffff) {
		t.Fatal("native C-owned links must exceed4GiB")
	}
	path := "../server/testdata/monster_dead_update_native_original_544ec0.json"
	if p := os.Getenv("NOX_MONSTER_DEAD_UPDATE_ORIGINAL_FIXTURE"); p != "" {
		path = p
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []struct {
			Input struct {
				Entry     string `json:"entry"`
				Update    int    `json:"update"`
				Zombie    uint32 `json:"zombie"`
				Status    uint32 `json:"status"`
				Frame     uint32 `json:"frame"`
				DeadFrame uint32 `json:"dead_frame"`
				Duration  uint32 `json:"duration"`
				Waypoints uint32 `json:"waypoints"`
				Seen      uint8  `json:"seen"`
				Delete    uint32 `json:"delete"`
				Enemy     bool   `json:"enemy"`
				Mutation  struct {
					Stage string `json:"stage"`
				} `json:"mutation"`
			} `json:"input"`
			Result struct {
				Events []struct {
					Kind string `json:"kind"`
				} `json:"events"`
				State struct {
					UpdateNil bool `json:"update_nil"`
					Updates   []struct {
						Waypoints   uint32     `json:"waypoints"`
						Seen        uint8      `json:"seen"`
						WaypointIDs [16]uint32 `json:"waypoint_ids"`
						EnemyIDs    [16]uint32 `json:"enemy_ids"`
						Enemy       bool       `json:"enemy"`
						Field91     uint32     `json:"field91"`
						Electric    uint8      `json:"electric"`
					} `json:"updates"`
				} `json:"state"`
			} `json:"result"`
			CleanupReturn int `json:"cleanup_return_update"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	found, wantDelete, wantRaise := false, false, false
	var expectedWP, expectedSeen [16]uint32
	var expectedEnemy, expectedUpdateNil bool
	var expectedField91 uint32
	var expectedElectric uint8
	for _, row := range fixture.Rows {
		c := row.Input
		if c.Mutation.Stage != "" || c.Update != 0 || c.Waypoints != wp || c.Seen != seen {
			continue
		}
		if cleanup {
			if c.Entry != "cleanup" || row.CleanupReturn != 0 {
				continue
			}
		} else if c.Entry != "update" || c.Zombie != query || c.Status != status || c.Frame != frame || c.DeadFrame != deadFrame || c.Duration != duration || c.Delete != delete || c.Enemy != enemy {
			continue
		}
		if found {
			t.Fatal("ambiguous frozen original composition")
		}
		found = true
		for _, e := range row.Result.Events {
			if e.Kind == "delete" || e.Kind == "burn" {
				wantDelete = true
			}
			if e.Kind == "raise" {
				wantRaise = true
			}
		}
		v := row.Result.State.Updates[0]
		expectedWP, expectedSeen, expectedEnemy, expectedField91, expectedElectric = v.WaypointIDs, v.EnemyIDs, v.Enemy, v.Field91, v.Electric
		expectedUpdateNil = row.Result.State.UpdateNil
	}
	if !found {
		t.Fatal("missing frozen original composition", mode)
	}
	native := &deadUpdateNativeServices544EC0{Server: GetServer()}
	old := GetServer
	GetServer = func() Server { return native }
	defer func() { GetServer = old }()
	a, ok := server.GetAIAction(ai.ACTION_DEAD).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary native dead dispatcher")
	}
	beforeUnit, beforeUpdate, beforeDefinition, beforeMissile, beforeHealth := *unit, *update, *definition, *missile, *unit.HealthData
	wpBytes := func() []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(&waypoints[0])), int(unsafe.Sizeof(waypoints[0]))*len(waypoints))
	}
	beforeWaypoints := append([]byte(nil), wpBytes()...)
	checkDecision := func() {
		if wantDelete {
			if len(native.deleted) != 1 || native.deleted[0] != unit {
				t.Fatal("original delete/burn boundary", native.deleted)
			}
		} else if len(native.deleted) != 0 {
			t.Fatal("unexpected native delete boundary")
		}
		if wantRaise {
			blocked := object.FlagDead | object.FlagAllowOverlap | object.FlagShort | object.FlagNoCollide
			if update.AIStackInd != 1 || update.AIStack[0].Type() != ai.DEPENDENCY_UNINTERRUPTABLE || update.AIStack[1].Type() != ai.ACTION_GET_UP || unit.ObjFlags.HasAny(blocked) || unit.HealthData.Cur != unit.HealthData.Max || unit.HealthData.Field2 != unit.HealthData.Cur || !s.AI.StackChanged {
				t.Fatal("real native Raise/Pop/Push/HP effects")
			}
		} else {
			if (unit.Update == nil) != expectedUpdateNil || update.Field91 != expectedField91 || update.Field523_2 != expectedElectric || (update.CurrentEnemy != nil) != expectedEnemy {
				t.Fatal("original root final fields")
			}
			for i := range update.Waypoints {
				if (update.Waypoints[i] != nil) != (expectedWP[i] != 0) || (update.SeenEnemies[i] != nil) != (expectedSeen[i] != 0) {
					t.Fatal("original counted prefixes", i)
				}
			}
		}
		if query == 0 && !cleanup {
			if unit.IsUpdatable != 0 || s.Objs.UpdatableList != nil {
				t.Fatal("real native RemoveUpdatable did not unlink")
			}
		}
		if !bytes.Equal(beforeWaypoints, wpBytes()) {
			t.Fatal("waypoint pointee canaries changed")
		}
	}
	if cleanup {
		if !s.MonsterActionDeadUpdate544EC0(unit, server.MonsterActionDeadRuntime544D80{IsZombie: func(*server.Object) bool { return false }, RemoveUpdatable: func(*server.Object) {}}) {
			t.Fatal("ordinary cleanup control")
		}
		unit.Update = beforeUnit.Update
	} else {
		a.Update(unit)
	}
	checkDecision()
	expectedUnit, expectedUpdate, expectedDefinition, expectedMissile, expectedHealth, expectedAI := *unit, *update, *definition, *missile, *unit.HealthData, s.AI.StackChanged
	t.Logf("DEAD UPDATE COMPOSITION ORDINARY COMPLETED: mode=%s original-delete=%t original-raise=%t unit=%p update=%p", mode, wantDelete, wantRaise, unit, update)
	*unit, *update, *definition, *missile, *unit.HealthData = beforeUnit, beforeUpdate, beforeDefinition, beforeMissile, beforeHealth
	s.AI.StackChanged = false
	native.deleted = nil
	s.Objs.UpdatableList = nil
	if unit.IsUpdatable != 0 {
		s.Objs.UpdatableList = unit
	}
	if cleanup {
		if got := ccall.CallUPtrPtr(monsterDeadCleanupPublicC544F70(), unit.CObj()); got != uintptr(unsafe.Pointer(update)) {
			t.Fatalf("original native cleanup return=%x want=%p", got, update)
		}
	} else {
		ccall.CallVoidPtr(monsterDeadUpdatePublicC544EC0(), unit.CObj())
	}
	checkDecision()
	raw := func(v unsafe.Pointer, n uintptr) []byte { return unsafe.Slice((*byte)(v), int(n)) }
	for _, v := range []struct {
		got, want unsafe.Pointer
		n         uintptr
	}{{unit.CObj(), unsafe.Pointer(&expectedUnit), unsafe.Sizeof(*unit)}, {unsafe.Pointer(update), unsafe.Pointer(&expectedUpdate), unsafe.Sizeof(*update)}, {unsafe.Pointer(definition), unsafe.Pointer(&expectedDefinition), unsafe.Sizeof(*definition)}, {missile.CObj(), unsafe.Pointer(&expectedMissile), unsafe.Sizeof(*missile)}, {unsafe.Pointer(unit.HealthData), unsafe.Pointer(&expectedHealth), unsafe.Sizeof(*unit.HealthData)}} {
		if !bytes.Equal(raw(v.got, v.n), raw(v.want, v.n)) {
			t.Fatal("full native unit/update/definition/missile/health differs from ordinary control")
		}
	}
	if s.AI.StackChanged != expectedAI {
		t.Fatal("native AI global differs")
	}
	t.Logf("DEAD UPDATE COMPOSITION NATIVE C PASS: mode=%s original-delete=%t original-raise=%t actual-cleanup-pointer=%t full-records=true", mode, wantDelete, wantRaise, cleanup)
}
