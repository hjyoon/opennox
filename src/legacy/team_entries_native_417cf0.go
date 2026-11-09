package legacy

import (
	"encoding/binary"
	"math"
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// All links refer to actual native Object/Player/Team owners. DWORD fields and
// the 18-byte wire prefix retain their original widths.
type teamEntriesRuntime417CF0 struct {
	RemoveActive func() int32
	Limit        func() int32
	Count        func() byte
	SetGameplay  func(uint32)
	AutoAssign   func(int32)
	FirstTeam    func() *server.Team
	NextTeam     func(*server.Team) *server.Team
	FirstUnit    func() *server.Object
	NextUnit     func(*server.Object) *server.Object
	ByID         func(server.TeamID) *server.Team
	ClientCode   func() uint32
	NoRendering  func() bool
	Change       func(*server.ObjectTeam, *server.Team, int32)
	Attach       func(server.TeamID, *server.ObjectTeam, int32)
	Prepare      func(*server.Team)
	Flag512      func() bool
	Alloc        func(uintptr) unsafe.Pointer
	Free         func(unsafe.Pointer)
	Send         func(unsafe.Pointer, int)
}

type teamEntriesFrame417CF0 struct {
	Runtime  *teamEntriesRuntime417CF0
	Previous *teamEntriesFrame417CF0
}

var teamEntriesHead417CF0 *teamEntriesFrame417CF0

func pushTeamEntriesRuntime417CF0(value *teamEntriesRuntime417CF0) func() {
	f := &teamEntriesFrame417CF0{value, teamEntriesHead417CF0}
	teamEntriesHead417CF0 = f
	return func() {
		if teamEntriesHead417CF0 != f {
			panic("team entry runtime pop order")
		}
		teamEntriesHead417CF0 = f.Previous
	}
}

func teamEntriesCurrent417CF0() *teamEntriesRuntime417CF0 {
	if teamEntriesHead417CF0 != nil {
		return teamEntriesHead417CF0.Runtime
	}
	return &teamEntriesDefaults417CF0
}

func sub417CF0Native() int32 { return teamEntriesCurrent417CF0().RemoveActive() }
func sub417DC0Native() int32 { return teamEntriesCurrent417CF0().Limit() }

func sub417DE0Native() byte {
	rt := teamEntriesCurrent417CF0()
	var count byte
	for team := rt.FirstTeam(); team != nil; team = rt.NextTeam(team) {
		if team.Ind60() != 0 {
			count++ // original char result wraps at eight bits.
		}
	}
	return count
}

func sub418390Native() int32 {
	rt := teamEntriesCurrent417CF0()
	if rt.Count() == 0 {
		return 0
	}
	rt.SetGameplay(2)
	rt.AutoAssign(0)
	return 1
}

// Explicit conversions separate the original x87 operations and inhibit
// contraction. Match gameplay precision53/chop without changing thread state.
//
//go:noinline
func teamEntriesAddChop53_4183C0(a, b float64) float64 {
	sum := float64(a + b)
	vb := float64(sum - a)
	err := float64(float64(a-float64(sum-vb)) + float64(b-vb))
	if sum > 0 && err < 0 || sum < 0 && err > 0 {
		return math.Nextafter(sum, 0)
	}
	return sum
}

//go:noinline
func teamEntriesSquareChop53_4183C0(value float64) float64 {
	square := float64(value * value)
	if math.FMA(value, value, -square) < 0 {
		return math.Nextafter(square, 0)
	}
	return square
}

func teamEntriesSpillChop32_4183C0(value float64) float32 {
	narrow := float32(value)
	if math.Abs(float64(narrow)) > math.Abs(value) {
		narrow = math.Nextafter32(narrow, 0)
	}
	return narrow
}

func sub4183C0Native() int32 {
	rt := teamEntriesCurrent417CF0()
	for unit := rt.FirstUnit(); unit != nil; unit = rt.NextUnit(unit) {
		ud := (*server.PlayerUpdateData)(unit.UpdateData)
		if ud != nil {
			player := ud.Player
			if !((player.NetCodeVal != rt.ClientCode() || !rt.NoRendering()) &&
				(player.Field3680&1 == 0 || player.Field3680&0x20 != 0)) {
				continue
			}
		}
		previous := rt.ByID(unit.TeamVal.ID)
		var nearest *server.Team
		best := float32(1000000000)
		for team := rt.FirstTeam(); team != nil; team = rt.NextTeam(team) {
			flag := (*server.Object)(team.Field_72)
			if flag == nil {
				continue
			}
			dy := teamEntriesAddChop53_4183C0(float64(flag.PosVec.Y), -float64(unit.PosVec.Y))
			dx := teamEntriesAddChop53_4183C0(float64(flag.PosVec.X), -float64(unit.PosVec.X))
			distance := teamEntriesAddChop53_4183C0(teamEntriesSquareChop53_4183C0(dy), teamEntriesSquareChop53_4183C0(dx))
			// Original FCOM C0 admits less OR unordered, not just ordered less.
			if !(distance >= float64(best)) {
				best = teamEntriesSpillChop32_4183C0(distance)
				nearest = team
			}
		}
		if nearest != nil {
			if previous == nil {
				rt.Attach(nearest.ID(), &unit.TeamVal, int32(unit.NetCode))
			} else if nearest != previous {
				rt.Change(&unit.TeamVal, nearest, int32(unit.NetCode))
			}
		}
	}
	return 0
}

func sub4184D0Native(team *server.Team) {
	if team == nil {
		return
	}
	rt := teamEntriesCurrent417CF0()
	rt.Prepare(team)
	p := team.C()
	var header [18]byte
	// Preserve the source read order after Prepare's possible mutation.
	binary.LittleEndian.PutUint32(header[6:10], *(*uint32)(unsafe.Add(p, 60)))
	binary.LittleEndian.PutUint32(header[2:6], uint32(*(*byte)(unsafe.Add(p, 57))))
	binary.LittleEndian.PutUint16(header[:2], 196)
	binary.LittleEndian.PutUint32(header[10:14], *(*uint32)(unsafe.Add(p, 52)))
	header[16] = *(*byte)(unsafe.Add(p, 56))
	header[17] = *(*byte)(unsafe.Add(p, 68))
	if rt.Flag512() {
		header[14] = 1
	}
	var length uintptr
	for *(*uint16)(unsafe.Add(p, 2*length)) != 0 {
		length++
	}
	header[15] = byte(length)
	size := 18 + 2*int(header[15]) // length is the low BYTE, not native size_t.
	allocation := rt.Alloc(uintptr(size))
	defer func() { rt.Free(allocation); runtime.KeepAlive(team) }()
	packet := unsafe.Slice((*byte)(allocation), size)
	copy(packet, header[:])
	copy(packet[18:], unsafe.Slice((*byte)(p), size-18))
	rt.Send(allocation, size)
}
