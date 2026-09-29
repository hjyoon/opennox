package legacy

import (
	"strconv"
	"testing"
)

func TestMapgenConnectPrefabs5266F0PreservesNativePointers(t *testing.T) {
	got := mapgenConnectPrefabsFixture5266F0()
	if !got.result || got.calls != 2 {
		t.Fatalf("connection result=%v calls=%d, want success on second candidate", got.result, got.calls)
	}
	if got.callbackAddresses[0] != got.candidateAddresses[0] ||
		got.callbackAddresses[1] != got.candidateAddresses[1] {
		t.Fatalf("candidate order = [%#x %#x], want [%#x %#x]",
			got.callbackAddresses[0], got.callbackAddresses[1],
			got.candidateAddresses[0], got.candidateAddresses[1])
	}
	if got.firstNextAddress != got.candidateAddresses[1] {
		t.Fatalf("first candidate next = %#x, want %#x", got.firstNextAddress, got.candidateAddresses[1])
	}
	if got.newRoomAddress == 0 {
		t.Fatal("replacement room token did not resolve")
	}
	if !got.linked || !got.hasTypeOneNeighbor {
		t.Fatalf("native adjacency link=%v type-one-neighbor=%v", got.linked, got.hasTypeOneNeighbor)
	}
	if got.forwardNeighborAddress != got.candidateAddresses[1] {
		t.Fatalf("forward neighbor = %#x, want %#x", got.forwardNeighborAddress, got.candidateAddresses[1])
	}
	if got.reverseNeighborAddress != got.candidateAddresses[0] {
		t.Fatalf("reverse neighbor = %#x, want %#x", got.reverseNeighborAddress, got.candidateAddresses[0])
	}
	if strconv.IntSize == 64 {
		addresses := map[string]uintptr{
			"theme":    got.themeAddress,
			"prefab":   got.prefabAddress,
			"old room": got.oldRoomAddress,
			"new room": got.newRoomAddress,
		}
		for i, address := range got.candidateAddresses {
			addresses["candidate "+strconv.Itoa(i)] = address
		}
		for name, address := range addresses {
			if address <= 1<<32 {
				t.Fatalf("%s fixture address = %#x, want address above PE32 range", name, address)
			}
		}
	}
}

func TestMapgenConnectPrefabs5266F0BuildsNativeConnector(t *testing.T) {
	got := mapgenConnectPrefabsActualFixture5266F0()
	if !got.result {
		t.Fatal("actual prefab connector rejected an unobstructed aligned room")
	}
	if got.roomCount != 3 {
		t.Fatalf("room count = %d, want replacement, candidate, and connector", got.roomCount)
	}
	if got.replacementAddress == 0 || got.candidateAddress == 0 || got.connectorAddress == 0 {
		t.Fatalf("room addresses = replacement %#x candidate %#x connector %#x",
			got.replacementAddress, got.candidateAddress, got.connectorAddress)
	}
	if got.replacementNeighborAddress != got.connectorAddress ||
		got.candidateNeighborAddress != got.connectorAddress {
		t.Fatalf("outer neighbors = replacement %#x candidate %#x, want connector %#x",
			got.replacementNeighborAddress, got.candidateNeighborAddress, got.connectorAddress)
	}
	if got.connectorNorthNeighborAddress != got.candidateAddress ||
		got.connectorSouthNeighborAddress != got.replacementAddress {
		t.Fatalf("connector neighbors = north %#x south %#x, want candidate %#x replacement %#x",
			got.connectorNorthNeighborAddress, got.connectorSouthNeighborAddress,
			got.candidateAddress, got.replacementAddress)
	}
	if strconv.IntSize == 64 {
		addresses := map[string]uintptr{
			"replacement": got.replacementAddress,
			"candidate":   got.candidateAddress,
			"connector":   got.connectorAddress,
		}
		for name, address := range addresses {
			if address <= 1<<32 {
				t.Fatalf("%s fixture address = %#x, want address above PE32 range", name, address)
			}
		}
	}
}
