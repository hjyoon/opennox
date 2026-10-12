package opennox

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	ns4 "github.com/opennox/noxscript/ns/v4"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

type nsTeamClearCase419570 struct {
	teamClearCase419570
	Kind    string
	Ignored bool
}
type nsForeignTeam419570 struct{ ns4.Team }

func nsTeamClearCases419570() []nsTeamClearCase419570 {
	var out []nsTeamClearCase419570
	for _, tc := range teamClearCases419570() {
		out = append(out, nsTeamClearCase419570{teamClearCase419570: tc})
	}
	for _, kind := range []string{"foreign-value", "foreign-nil", "native-pointer", "native-pointer-nil", "native-value-nil"} {
		out = append(out, nsTeamClearCase419570{
			teamClearCase419570: teamClearCase419570{Name: kind, Flags: noxflags.GameHost | noxflags.GameOnline, Player: true, Code: 0xf123abcd},
			Kind:                kind, Ignored: kind != "native-value-nil",
		})
	}
	out = append(out, nsTeamClearCase419570{
		teamClearCase419570: teamClearCase419570{Name: "foreign team with nil object", Flags: noxflags.GameHost | noxflags.GameOnline, Nil: true, Code: 0xf123abcd},
		Kind:                "foreign-nil", Ignored: true,
	})
	return out
}

// This adds SDK dispatch to a separate copy of the frozen native fixture.
// All original 13 inputs, full-record comparison and wire assertions remain.
// The six added cases observe interface type guards and recognized nil payload.
func nsTeamClearRun419570(t *testing.T, tc nsTeamClearCase419570, route string) teamClearResult419570 {
	t.Helper()
	noxflags.ResetGame()
	s := NewServer(nil, console.NewMultiPrinter(), strman.New())
	noxServer = s
	defer s.Close()
	s.Types.Free()
	handles.Init()
	defer handles.Release()
	s.SetFrame(37)
	s.SetTickRate(30)
	s.NetList.Init()
	defer s.NetList.Free()
	legacy.Sub_4E4F80()
	legacy.Sub_4E4EF0()
	legacy.Sub_4E4ED0()
	defer legacy.Sub_4E4F80()
	legacy.Set_dword_5d4594_2649712(1)
	*memmap.PtrUint32(0x587000, 4728) = 3
	noxflags.SetGame(tc.Flags)
	if err := s.Types.ReadObjectType(&things.Thing{Name: "GameBall", Class: []things.Class{"SIMPLE"}}); err != nil {
		t.Fatal(err)
	}
	team, second := s.Teams.Create(7), s.Teams.Create(9)
	team.SetNameAnd68("original team", 0x10203040)
	team.Lessons = 0x7654321
	team.Field_72 = nil
	type record struct {
		Pointer unsafe.Pointer
		Size    uintptr
	}
	records := map[string]record{"team": {unsafe.Pointer(team), unsafe.Sizeof(*team)}, "second-team": {unsafe.Pointer(second), unsafe.Sizeof(*second)}}
	newObject := func(name string, value server.Object) *server.Object {
		p, free := alloc.New(server.Object{})
		*p = value
		// Match serverObjects.NewObject's real server association. Keep the
		// original C-owned fixture bytes/inputs; copy only the native handle
		// from this genuinely registered server, never a Go pointer or result.
		handle := reflect.ValueOf(s.Server).Elem().FieldByName("handle")
		field, ok := reflect.TypeOf(*p).FieldByName("serverHandle")
		if !ok || !handle.IsValid() || handle.Kind() != reflect.Uintptr || handle.Uint() == 0 || field.Type.Kind() != reflect.Uintptr {
			t.Fatal("native server association fields unavailable")
		}
		*(*uintptr)(unsafe.Add(unsafe.Pointer(p), field.Offset)) = uintptr(handle.Uint())
		if p.Server() != s.Server {
			t.Fatal("native object did not bind to actual fixture server")
		}
		t.Cleanup(free)
		records[name] = record{unsafe.Pointer(p), unsafe.Sizeof(*p)}
		return p
	}
	newMonsterUpdate := func(name string) unsafe.Pointer {
		p, free := alloc.New(server.MonsterUpdateData{})
		*p = server.MonsterUpdateData{}
		t.Cleanup(free)
		records[name] = record{unsafe.Pointer(p), unsafe.Sizeof(*p)}
		return unsafe.Pointer(p)
	}
	other := newObject("other-object", server.Object{TypeInd: 8, ObjClass: object.ClassSimple, ObjFlags: object.FlagEnabled, NetCode: 0x2468, Extent: 0xcafef00d})
	var unit *server.Object
	if !tc.Nil {
		unit = newObject("object", server.Object{TypeInd: 7, ObjClass: object.ClassMonster, ObjFlags: object.FlagEnabled, NetCode: tc.Code, Extent: 0x89abcdef, Weight: 71})
	}
	if tc.Player {
		player := s.Players.NewRaw(int(tc.Code))
		if player == nil {
			t.Fatal("real player registry allocation failed")
		}
		player.Field3680 = 0x10
		records["player"] = record{unsafe.Pointer(player), unsafe.Sizeof(*player)}
		newPlayerUpdate := func(name string) unsafe.Pointer {
			p, free := alloc.New(server.PlayerUpdateData{})
			*p = server.PlayerUpdateData{Player: player, State: 9}
			t.Cleanup(free)
			records[name] = record{unsafe.Pointer(p), unsafe.Sizeof(*p)}
			return unsafe.Pointer(p)
		}
		unit.ObjClass = object.ClassPlayer
		unit.UpdateData = newPlayerUpdate("update-owner")
		player.PlayerUnit = unit
		defer func() { player.PlayerUnit = nil }()
		var previous *server.Object
		for i, cls := range []object.Class{object.ClassMonster, object.ClassPlayer, object.ClassSimple} {
			owned := newObject(fmt.Sprintf("owned-%d", i), server.Object{TypeInd: uint16(10 + i), ObjClass: cls, ObjFlags: object.FlagEnabled, ObjOwner: unit, NetCode: uint32(40 + i), Field35: 0x100})
			if i == 0 {
				owned.UpdateData = newMonsterUpdate("update-owned-monster")
			}
			if i == 1 {
				owned.UpdateData = newPlayerUpdate("update-owned-player")
			}
			if previous == nil {
				unit.Field129 = owned
			} else {
				previous.Field128 = owned
			}
			previous = owned
		}
	} else if unit != nil {
		unit.UpdateData = newMonsterUpdate("update-monster")
	}
	if !tc.Alone {
		s.Teams.AttachObject(other.TeamPtr(), team.ID())
	}
	if unit != nil {
		switch {
		case tc.Stale:
			unit.TeamVal = server.ObjectTeam{Field0: 0xabcdef12, ID: 199}
		case tc.Unteamed:
			unit.TeamVal = server.ObjectTeam{Field0: 0xabcdef12}
		default:
			if s.Teams.AttachObject(unit.TeamPtr(), team.ID()) != team {
				t.Fatal("real team attachment failed")
			}
		}
	}
	snapshot := func() map[string][]byte {
		out := make(map[string][]byte)
		for name, r := range records {
			out[name] = bytes.Clone(unsafe.Slice((*byte)(r.Pointer), int(r.Size)))
		}
		return out
	}

	var wrapped ns4.Obj = nsObj{s: s}
	if unit != nil {
		wrapped = s.noxScriptP().toObj(unit)
		if wrapped == nil {
			t.Fatal("actual namespace factory returned nil")
		}
	}
	var supplied ns4.Team
	if tc.Assign {
		for _, candidate := range s.NoxScript().Teams() {
			if native, ok := candidate.(nsTeam); ok && native.t == second {
				supplied = candidate
			}
		}
		if supplied == nil {
			t.Fatal("actual namespace Teams did not expose second team")
		}
	} else {
		switch tc.Kind {
		case "foreign-value":
			supplied = nsForeignTeam419570{}
		case "foreign-nil":
			supplied = (*nsForeignTeam419570)(nil)
		case "native-pointer":
			supplied = &nsTeam{s: s, t: second}
		case "native-pointer-nil":
			supplied = (*nsTeam)(nil)
		case "native-value-nil":
			supplied = nsTeam{s: s}
		}
	}
	before := snapshot()
	high := true
	for _, r := range records {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(r.Pointer) <= 0xffffffff {
			high = false
		}
	}
	invoke := func() {
		if route == "C" {
			if tc.Ignored {
				return
			}
			if tc.Assign {
				legacy.Nox_xxx_createAtImpl_4191D0(second.ID(), unit.TeamPtr(), 1, int(unit.NetCode), 0)
			} else {
				legacy.Nox_xxx_netChangeTeamMb_419570(unit.TeamPtr(), tc.Code)
			}
		} else {
			wrapped.SetTeam(supplied)
		}
	}
	invoke()
	if unit != nil {
		if wrapped.HasTeam(nil) != (unit.TeamVal.ID == 0) {
			t.Fatal("public SDK HasTeam(nil) differs from actual native ID")
		}
		want, got := s.Teams.ByID(unit.TeamVal.ID), wrapped.Team()
		if want == nil {
			if got != nil {
				t.Fatal("public SDK Team exposes nonexistent team")
			}
		} else if native, ok := got.(nsTeam); !ok || native.t != want || !wrapped.HasTeam(got) {
			t.Fatal("public SDK Team/HasTeam differs from actual native team")
		}
	}
	after := snapshot()
	result := teamClearResult419570{Changes: make(map[string][]teamClearByte419570), Members: s.Teams.MemberCount(team), SecondMembers: s.Teams.MemberCount(second), Active: team.Active(), OtherID: other.TeamVal.ID, TeamCount: s.Teams.Count(), HighPointers: high}
	if unit != nil {
		result.TeamID = unit.TeamVal.ID
	}
	for name, current := range after {
		var delta []teamClearByte419570
		for i, value := range current {
			if value != before[name][i] {
				delta = append(delta, teamClearByte419570{i, before[name][i], value})
			}
		}
		result.Changes[name] = delta
	}
	result.Packets = teamClearPackets419570(t, s)
	invoke()
	result.RepeatChanged = !reflect.DeepEqual(after, snapshot())
	result.RepeatPackets = teamClearPackets419570(t, s)
	return result
}

func TestNoxScriptTeamClearActualNative419570(t *testing.T) {
	if route := os.Getenv("OPENNOX_NS_CLEAR419570_ROUTE"); route != "" {
		for _, tc := range nsTeamClearCases419570() {
			if tc.Name == os.Getenv("OPENNOX_NS_CLEAR419570_CASE") {
				result, err := json.Marshal(nsTeamClearRun419570(t, tc, route))
				if err != nil {
					t.Fatal(err)
				}
				fmt.Printf("ACTUAL NS TEAM CLEAR SNAPSHOT: %s\n", result)
				return
			}
		}
		t.Fatal("unknown child fixture")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	child := func(t *testing.T, tc nsTeamClearCase419570, route string) teamClearResult419570 {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, executable, "-test.run=^TestNoxScriptTeamClearActualNative419570$", "-test.v", "-test.count=1")
		cmd.Env = append(os.Environ(), "OPENNOX_NS_CLEAR419570_ROUTE="+route, "OPENNOX_NS_CLEAR419570_CASE="+tc.Name)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s %s child failed: %v\n%s", tc.Name, route, err, output)
		}
		for _, line := range strings.Split(string(output), "\n") {
			if at := strings.Index(line, "ACTUAL NS TEAM CLEAR SNAPSHOT: "); at >= 0 {
				var result teamClearResult419570
				if err := json.Unmarshal([]byte(line[at+len("ACTUAL NS TEAM CLEAR SNAPSHOT: "):]), &result); err != nil {
					t.Fatal(err)
				}
				return result
			}
		}
		t.Fatalf("native snapshot missing: %s", output)
		return teamClearResult419570{}
	}
	for _, tc := range nsTeamClearCases419570() {
		t.Run(tc.Name, func(t *testing.T) {
			want := child(t, tc, "C")
			if !want.HighPointers || want.RepeatChanged || len(want.RepeatPackets) != 0 {
				t.Fatalf("C control pointer/idempotence failed: %+v", want)
			}
			valid := !tc.Nil && !tc.Stale && !tc.Unteamed && !tc.Ignored
			if valid && !tc.Assign && (want.TeamID != 0 || want.Members != 1-bool2int(tc.Alone)) {
				t.Fatalf("actual C detach/count failed: %+v", want)
			}
			if valid && !tc.Assign && tc.Flags.HasAll(noxflags.GameHost|noxflags.GameOnline) {
				packet := []byte{0xc4, 2, 0, 0, 0, 0}
				binary.LittleEndian.PutUint32(packet[2:], tc.Code)
				count := 0
				for _, p := range want.Packets {
					if bytes.Equal(p.Payload, packet) {
						count++
					}
				}
				if count != 1 {
					t.Fatalf("actual full-width removal packet missing/duplicated: %+v", want.Packets)
				}
			}
			if tc.Player && !tc.Ignored {
				count := 0
				for _, p := range want.Packets {
					if bytes.Equal(p.Payload, []byte{54}) {
						count++
					}
				}
				if count != 1 {
					t.Fatalf("actual player refresh packet missing/duplicated: %+v", want.Packets)
				}
			}
			if tc.Ignored {
				if len(want.Packets) != 0 {
					t.Fatal("ignored team argument produced packet")
				}
				for name, delta := range want.Changes {
					if len(delta) != 0 {
						t.Fatalf("ignored team argument changed %s: %+v", name, delta)
					}
				}
			}
			got := child(t, tc, "SDK")
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("SDK setter differs from full actual C service\ngot=%+v\nwant=%+v", got, want)
			}
		})
	}
}
