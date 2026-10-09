package server

import "math/bits"

// PlayerKillStatsRuntime425CA0 isolates the scalar counters, wire records and
// external statistics flush from the native-width Player inputs.
type PlayerKillStatsRuntime425CA0 struct {
	GameFlag     func(uint32) bool
	PlayerCount  func() uint32
	SetCount     func(uint32)
	CopyName     func(uint32, *Player)
	ConnectionIP func(int) uint32
	StoreIP      func(uint32, uint32)
	StoreTeam    func(uint32, uint32)
	StoreClass   func(uint32, byte)
	PairCount    func() uint32
	StorePair    func(uint32, byte, byte)
	SetPairCount func(uint32)
	Flush        func()
}

// PlayerKillStats425CA0 follows GAME.EXE 00425CA0..00425E8F. The original
// deliberately observable destinations are retained: a newly registered
// second player's name overwrites the first record, and a host second
// player's IP is also stored in the first record. The first registration
// checks the SECOND player's host byte before choosing the FIRST connection.
// In particular, a nil second player faults there only after count/name
// writes; it is not an early no-op. These are not AI or statistics repairs.
func PlayerKillStats425CA0(first, second *Player, runtime PlayerKillStatsRuntime425CA0) {
	if !runtime.GameFlag(0x2000) || runtime.GameFlag(0x1000) || first == nil {
		return
	}
	firstIndex := first.Field4648
	if firstIndex == -1 {
		index := runtime.PlayerCount()
		runtime.SetCount(index + 1)
		firstIndex = int32(index)
		runtime.CopyName(index, first)
		connection := 0
		if second.PlayerInd != 31 {
			connection = int(first.PlayerInd) + 1
		}
		ip := runtime.ConnectionIP(connection)
		runtime.StoreIP(index, bits.ReverseBytes32(ip))
		runtime.StoreTeam(index, first.Field2068)
		runtime.StoreClass(index, byte(first.Info().PlayerClass()))
		first.Field4648 = firstIndex
	}
	secondIndex := int32(-1)
	if second != nil {
		secondIndex = second.Field4648
		if secondIndex == -1 {
			index := runtime.PlayerCount()
			runtime.SetCount(index + 1)
			secondIndex = int32(index)
			runtime.CopyName(uint32(firstIndex), second)
			ipIndex := uint32(firstIndex)
			connection := 0
			if second.PlayerInd != 31 {
				connection = int(second.PlayerInd) + 1
				ipIndex = index
			}
			ip := runtime.ConnectionIP(connection)
			runtime.StoreIP(ipIndex, bits.ReverseBytes32(ip))
			runtime.StoreTeam(index, second.Field2068)
			runtime.StoreClass(index, byte(second.Info().PlayerClass()))
			second.Field4648 = secondIndex
		}
	}
	pairIndex := runtime.PairCount()
	runtime.StorePair(pairIndex, byte(firstIndex), byte(secondIndex))
	runtime.SetPairCount(pairIndex + 1)
	if runtime.PlayerCount() >= 255 {
		runtime.Flush()
	}
}
