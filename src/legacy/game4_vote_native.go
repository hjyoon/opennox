package legacy

/*
#include "game4_list_vote_native.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"unsafe"
)

// Records come from the native class allocator. Only the private list head
// moves from file-static C storage to a typed Go root; no PE32 diagnostic
// DWORD is used to store a native pointer.
var game4VoteHead *C.nox_vote_5066D0

func game4VotePlayer(unit *C.nox_object_t) *C.nox_playerInfo {
	if unit == nil || unit.obj_class&4 == 0 || unit.data_update == nil {
		return nil
	}
	return (*C.nox_player_update_data_t)(unit.data_update).player
}
func game4VoteTeam(unit *C.nox_object_t) *C.nox_object_team_t {
	return (*C.nox_object_team_t)(unsafe.Pointer(&unit.field_12))
}
func game4VoteNamesEqual(a, b *C.wchar2_t) bool {
	for offset := uintptr(0); ; offset += 2 {
		x := *(*uint16)(unsafe.Add(unsafe.Pointer(a), offset))
		y := *(*uint16)(unsafe.Add(unsafe.Pointer(b), offset))
		if x != y {
			return false
		}
		if x == 0 {
			return true
		}
	}
}
func game4VoteFindPlayer(name *C.wchar2_t) *C.nox_playerInfo {
	if name == nil {
		return nil
	}
	for p := C.game4_players_first_native(); p != nil; p = C.nox_common_playerInfoGetNext_416EE0(p) {
		if p.active == 1 && game4VoteNamesEqual(&p.name_final[0], name) {
			return p
		}
	}
	return nil
}
func game4VoteFind(kind uint32, target *C.nox_object_t) *C.nox_vote_5066D0 {
	for vote := game4VoteHead; vote != nil; vote = vote.next {
		if uint32(vote._type) == kind && (target == nil || vote.target == target) {
			return vote
		}
	}
	return nil
}
func game4VoteAddVoter(vote *C.nox_vote_5066D0, mask uint32) C.char {
	result := uint32(vote.voters)
	if result&mask == 0 {
		vote.voters |= C.uint32_t(mask)
		vote.count++
		result = uint32(vote.count)
	}
	return C.char(result)
}
func game4VoteRemoveVoter(vote *C.nox_vote_5066D0, mask uint32) {
	if vote == nil || uint32(vote.voters)&mask == 0 {
		return
	}
	vote.voters &^= C.uint32_t(mask)
	vote.count--
	if vote.count == 0 {
		sub_5067B0(vote)
	}
}

//export nox_xxx_allocVoteArray_5066D0
func nox_xxx_allocVoteArray_5066D0() C.int {
	allocator := C.nox_new_alloc_class(internCStr("VoteClass"), C.int(C.sizeof_nox_vote_5066D0), 64)
	C.nox_alloc_vote_1599652 = unsafe.Pointer(allocator)
	if allocator == nil {
		return 0
	}
	game4VoteHead = nil
	C.dword_5d4594_1599656 = 0
	return 1
}

//export sub_506700
func sub_506700() {
	if C.nox_alloc_vote_1599652 != nil {
		C.nox_alloc_class_free_all((*C.nox_alloc_class)(C.nox_alloc_vote_1599652))
	}
	game4VoteHead = nil
	C.dword_5d4594_1599656 = 0
}

//export sub_506720
func sub_506720() C.int {
	if C.nox_alloc_vote_1599652 != nil {
		C.nox_free_alloc_class((*C.nox_alloc_class)(C.nox_alloc_vote_1599652))
	}
	C.nox_alloc_vote_1599652 = nil
	game4VoteHead = nil
	C.dword_5d4594_1599656 = 0
	return 0
}

//export sub_506740
func sub_506740(unit *C.nox_object_t) C.int {
	p := game4VotePlayer(unit)
	if p == nil || p.playerInd >= 32 {
		return 0
	}
	mask := uint32(1) << uint32(p.playerInd)
	for vote := game4VoteHead; vote != nil; {
		next := vote.next
		if uint32(vote.voters)&mask != 0 {
			vote.voters &^= C.uint32_t(mask)
			vote.count--
		}
		if vote.count == 0 {
			sub_5067B0(vote)
		}
		vote = next
	}
	return 0
}

//export sub_5067B0
func sub_5067B0(vote *C.nox_vote_5066D0) {
	if vote == nil {
		return
	}
	if vote._type == 2 {
		for index := uint32(0); index < 32; index++ {
			if uint32(vote.voters)&(uint32(1)<<index) != 0 {
				nox_xxx_netSendVote_506840(C.int(index))
			}
		}
	}
	sub_506810(vote)
	if C.nox_alloc_vote_1599652 != nil {
		C.nox_alloc_class_free_obj_first((*C.nox_alloc_class)(C.nox_alloc_vote_1599652), unsafe.Pointer(vote))
	}
	if game4VoteHead == nil {
		sub_507190(255, 0)
	}
}

//export sub_506810
func sub_506810(vote *C.nox_vote_5066D0) *C.nox_vote_5066D0 {
	if vote == nil {
		return nil
	}
	result := vote
	if vote.next != nil {
		vote.next.previous = vote.previous
	}
	if vote.previous != nil {
		result = vote.next
		vote.previous.next = vote.next
	} else {
		game4VoteHead = vote.next
	}
	return result
}

//export nox_xxx_netSendVote_506840
func nox_xxx_netSendVote_506840(index C.int) C.int {
	return C.int(game4VoteSend(int32(index), []byte{0xee, 7}))
}

//export sub_506870
func sub_506870(kind C.int, unit *C.nox_object_t, name *C.wchar2_t) C.char {
	if unit == nil || unit.obj_class&4 == 0 {
		return 0
	}
	switch kind {
	case 0, 1:
		return sub_5068E0(kind, unit, name)
	case 2:
		if sub_506B00(kind, unit) != nil {
			return 1
		}
	case 3:
		if sub_506B80(kind, unit, name) != nil {
			return 1
		}
	}
	return 0
}

//export sub_5068E0
func sub_5068E0(kind C.int, unit *C.nox_object_t, name *C.wchar2_t) C.char {
	minimum := *memmap.PtrUint32(0x587000, 229980)
	if minimum == 0 || minimum > 32 || name == nil {
		return C.char(minimum)
	}
	voter := game4VotePlayer(unit)
	if voter == nil || voter.playerInd >= 32 {
		return 0
	}
	targetPlayer := game4VoteFindPlayer(name)
	if targetPlayer == nil || targetPlayer.playerInd == 31 || targetPlayer.playerUnit == nil || targetPlayer.playerUnit == unit {
		return 0
	}
	target := targetPlayer.playerUnit
	teamMode := bool(C.nox_xxx_CheckGameplayFlags_417DA0(4))
	if teamMode && C.nox_xxx_servCompareTeams_419150(game4VoteTeam(unit), game4VoteTeam(target)) == 0 {
		return 0
	}
	vote := game4VoteFind(uint32(kind), target)
	if vote == nil {
		vote = sub_506A20(kind, unit)
		if vote == nil {
			return 0
		}
		vote.target = target
		if teamMode {
			vote.team_mode = 1
		} else {
			vote.team_mode = 0
		}
	}
	return game4VoteAddVoter(vote, uint32(1)<<uint32(voter.playerInd))
}

//export nox_vote_record_alloc_5066D0
func nox_vote_record_alloc_5066D0() *C.nox_vote_5066D0 {
	if C.nox_alloc_vote_1599652 == nil {
		return nil
	}
	return (*C.nox_vote_5066D0)(C.nox_alloc_class_new_obj_zero((*C.nox_alloc_class)(C.nox_alloc_vote_1599652)))
}

//export sub_506A20
func sub_506A20(kind C.int, unit *C.nox_object_t) *C.nox_vote_5066D0 {
	if unit == nil || unit.obj_class&4 == 0 {
		return nil
	}
	wasEmpty := game4VoteHead == nil
	vote := nox_vote_record_alloc_5066D0()
	if vote == nil {
		return nil
	}
	vote._type = C.uint32_t(kind)
	vote.frame = C.uint32_t(gameFrame())
	vote.team = game4VoteTeam(unit)
	switch kind {
	case 0, 1:
		vote.threshold = C.uint8_t(*memmap.PtrUint8(0x587000, 229980))
	case 2, 3:
		vote.threshold = 6
	default:
		vote.threshold = C.uint8_t(*memmap.PtrUint8(0x587000, 229984))
	}
	nox_xxx_voteAddMB_506AD0(vote)
	if wasEmpty {
		sub_507190(255, 1)
	}
	return vote
}

//export nox_xxx_voteAddMB_506AD0
func nox_xxx_voteAddMB_506AD0(vote *C.nox_vote_5066D0) *C.nox_vote_5066D0 {
	if vote == nil {
		return nil
	}
	vote.previous, vote.next = nil, game4VoteHead
	if game4VoteHead != nil {
		game4VoteHead.previous = vote
	}
	game4VoteHead = vote
	return vote
}

//export sub_506B00
func sub_506B00(kind C.int, unit *C.nox_object_t) *C.nox_vote_5066D0 {
	if C.nox_server_resetQuestMinVotes_229988 == 0 {
		return nil
	}
	voter := game4VotePlayer(unit)
	if voter == nil || voter.playerInd >= 32 || voter.field_4792 == 0 {
		return nil
	}
	vote := game4VoteFind(uint32(kind), nil)
	if vote == nil {
		vote = sub_506A20(kind, unit)
		if vote == nil {
			return nil
		}
		vote.team_mode = 0
	}
	game4VoteAddVoter(vote, uint32(1)<<uint32(voter.playerInd))
	return vote
}

//export sub_506B80
func sub_506B80(kind C.int, unit *C.nox_object_t, name *C.wchar2_t) *C.nox_vote_5066D0 {
	if C.nox_server_kickQuestPlayerMinVotes_229992 == 0 || name == nil {
		return nil
	}
	voter := game4VotePlayer(unit)
	if voter == nil || voter.playerInd >= 32 || voter.field_4792 == 0 {
		return nil
	}
	targetPlayer := game4VoteFindPlayer(name)
	if targetPlayer == nil || targetPlayer.playerInd == 31 || targetPlayer.field_4792 == 0 || targetPlayer.playerUnit == nil || targetPlayer.playerUnit == unit {
		return nil
	}
	target := targetPlayer.playerUnit
	vote := game4VoteFind(uint32(kind), target)
	if vote == nil {
		vote = sub_506A20(kind, unit)
		if vote == nil {
			return nil
		}
		vote.target = target
	}
	game4VoteAddVoter(vote, uint32(1)<<uint32(voter.playerInd))
	return vote
}

//export sub_506C90
func sub_506C90(kind C.int, unit *C.nox_object_t, name *C.wchar2_t) {
	if unit == nil || unit.obj_class&4 == 0 {
		return
	}
	switch kind {
	case 0, 1:
		sub_506D00(unit, name)
	case 2:
		sub_506DE0(unit)
	case 3:
		sub_506E50(unit, name)
	}
}

//export sub_506D00
func sub_506D00(unit *C.nox_object_t, name *C.wchar2_t) {
	voter, target := game4VotePlayer(unit), game4VoteFindPlayer(name)
	if voter == nil || voter.playerInd >= 32 || target == nil || target.playerInd == 31 || target.playerUnit == nil {
		return
	}
	mask := uint32(1) << uint32(voter.playerInd)
	for vote := game4VoteHead; vote != nil; vote = vote.next {
		if vote._type == 0 && vote.target == target.playerUnit && uint32(vote.voters)&mask != 0 {
			game4VoteRemoveVoter(vote, mask)
			return
		}
	}
}

//export sub_506DE0
func sub_506DE0(unit *C.nox_object_t) {
	voter := game4VotePlayer(unit)
	if voter == nil || voter.playerInd >= 32 {
		return
	}
	mask := uint32(1) << uint32(voter.playerInd)
	for vote := game4VoteHead; vote != nil; vote = vote.next {
		if vote._type == 2 {
			game4VoteRemoveVoter(vote, mask)
			return
		}
	}
}

//export sub_506E50
func sub_506E50(unit *C.nox_object_t, name *C.wchar2_t) {
	voter, target := game4VotePlayer(unit), game4VoteFindPlayer(name)
	if voter == nil || voter.playerInd >= 32 || target == nil || target.playerInd == 31 || target.playerUnit == nil {
		return
	}
	mask := uint32(1) << uint32(voter.playerInd)
	for vote := game4VoteHead; vote != nil; vote = vote.next {
		if vote._type == 3 && vote.target == target.playerUnit && uint32(vote.voters)&mask != 0 {
			game4VoteRemoveVoter(vote, mask)
			return
		}
	}
}

//export nox_xxx_voteUptate_506F30
func nox_xxx_voteUptate_506F30() {
	for vote := game4VoteHead; vote != nil; {
		next := vote.next
		switch vote._type {
		case 0, 1:
			sub_506F80(vote)
		case 2:
			sub_507090(vote)
		case 3:
			sub_507100(vote)
		}
		vote = next
	}
}

//export sub_506F80
func sub_506F80(vote *C.nox_vote_5066D0) {
	if vote == nil || vote.target == nil || vote.target.obj_flags&0x20 != 0 {
		sub_5067B0(vote)
		return
	}
	vote.team = game4VoteTeam(vote.target)
	if sub_507000(vote) != 1 {
		return
	}
	if target := game4VotePlayer(vote.target); target != nil {
		C.nox_xxx_playerCallDisconnect_4DEAB0(C.int(target.playerInd), 4)
		C.sub_416770(15, &target.name_final[0], &target.serial[0])
	}
	sub_5067B0(vote)
}

//export sub_507000
func sub_507000(vote *C.nox_vote_5066D0) C.int {
	if vote == nil {
		return 0
	}
	if vote.count >= vote.threshold {
		return 1
	}
	var players uint32
	for p := C.game4_units_first_native(); p != nil; p = C.nox_xxx_getNextPlayerUnit_4DA7F0(p) {
		if vote.team_mode != 1 || C.nox_xxx_servCompareTeams_419150(vote.team, game4VoteTeam(p)) != 0 {
			players++
		}
	}
	if uint32(vote.count) >= players-1 && vote.count >= 2 {
		return 1
	}
	return 0
}
func game4VoteQuestPlayerCount() uint32 {
	var count uint32
	for unit := C.game4_units_first_native(); unit != nil; unit = C.nox_xxx_getNextPlayerUnit_4DA7F0(unit) {
		p := game4VotePlayer(unit)
		if p == nil {
			continue
		}
		if (!bool(C.nox_common_gameFlags_check_40A5C0(1)) || !bool(C.nox_common_getEngineFlag(C.NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) || p.playerInd != 31) && p.field_4792 == 1 {
			count++
		}
	}
	return count
}

//export sub_507090
func sub_507090(vote *C.nox_vote_5066D0) {
	if vote == nil || uint32(vote.count) < game4VoteQuestPlayerCount() {
		return
	}
	for unit := C.game4_units_first_native(); unit != nil; unit = C.nox_xxx_getNextPlayerUnit_4DA7F0(unit) {
		p := game4VotePlayer(unit)
		if p != nil && p.field_4792 == 1 {
			C.nox_xxx_playerRespawn_4F7EF0(p.playerUnit)
		}
	}
	C.nox_game_setQuestStage_4E3CD0(0)
	C.nox_xxx_mapLoad_4D2450(C.game4_quest_map_native())
	sub_5067B0(vote)
}

//export sub_507100
func sub_507100(vote *C.nox_vote_5066D0) {
	if vote == nil || vote.target == nil || vote.target.obj_flags&0x20 != 0 {
		sub_5067B0(vote)
		return
	}
	target := game4VotePlayer(vote.target)
	if target == nil || target.field_4792 == 0 {
		sub_5067B0(vote)
		return
	}
	if vote.count < vote.threshold {
		count := game4VoteQuestPlayerCount()
		if count <= 1 {
			sub_5067B0(vote)
			return
		}
		if !(uint32(vote.count) >= count-1 && vote.count >= 2) {
			return
		}
	}
	C.sub_4DCFB0(vote.target)
	C.sub_416770(15, &target.name_final[0], &target.serial[0])
	sub_5067B0(vote)
}

//export sub_507190
func sub_507190(recipient C.int, active C.char) C.int {
	return C.int(game4VoteSend(int32(recipient), []byte{0xee, 6, byte(active)}))
}

//export sub_5071C0
func sub_5071C0() C.int {
	if game4VoteHead != nil {
		return 1
	}
	return 0
}

//export nox_vote_head_5066D0
func nox_vote_head_5066D0() *C.nox_vote_5066D0 { return game4VoteHead }

//export nox_vote_allocator_5066D0
func nox_vote_allocator_5066D0() unsafe.Pointer { return C.nox_alloc_vote_1599652 }

//export nox_vote_record_size_5066D0
func nox_vote_record_size_5066D0() C.size_t { return C.sizeof_nox_vote_5066D0 }

//export nox_vote_next_5066D0
func nox_vote_next_5066D0(vote *C.nox_vote_5066D0) *C.nox_vote_5066D0 {
	if vote == nil {
		return nil
	}
	return vote.next
}

//export nox_vote_previous_5066D0
func nox_vote_previous_5066D0(vote *C.nox_vote_5066D0) *C.nox_vote_5066D0 {
	if vote == nil {
		return nil
	}
	return vote.previous
}

//export nox_vote_team_5066D0
func nox_vote_team_5066D0(vote *C.nox_vote_5066D0) *C.nox_object_team_t {
	if vote == nil {
		return nil
	}
	return vote.team
}

//export nox_vote_frame_5066D0
func nox_vote_frame_5066D0(vote *C.nox_vote_5066D0) C.uint32_t {
	if vote == nil {
		return 0
	}
	return vote.frame
}

//export nox_vote_voters_5066D0
func nox_vote_voters_5066D0(vote *C.nox_vote_5066D0) C.uint32_t {
	if vote == nil {
		return 0
	}
	return vote.voters
}

//export nox_vote_count_5066D0
func nox_vote_count_5066D0(vote *C.nox_vote_5066D0) C.uint8_t {
	if vote == nil {
		return 0
	}
	return vote.count
}

//export nox_vote_threshold_5066D0
func nox_vote_threshold_5066D0(vote *C.nox_vote_5066D0) C.uint8_t {
	if vote == nil {
		return 0
	}
	return vote.threshold
}

//export nox_vote_set_voters_5066D0
func nox_vote_set_voters_5066D0(vote *C.nox_vote_5066D0, voters C.uint32_t, count C.uint8_t) {
	if vote != nil {
		vote.voters, vote.count = voters, count
	}
}
