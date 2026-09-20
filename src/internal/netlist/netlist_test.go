package netlist

import (
	"bytes"
	"testing"

	"github.com/opennox/opennox/v1/common/ntype"
)

func TestCopyPacketsAReusesBackingBuffer(t *testing.T) {
	const ind = ntype.PlayerInd(hostIndex)

	list := New()
	list.Init()
	t.Cleanup(list.Free)

	for i := 0; i < bufSize*2; i++ {
		want := []byte{byte(i), byte(i >> 8)}
		if !list.AddToMsgListCli(ind, Kind1, want) {
			t.Fatalf("iteration %d: failed to queue packet after previous packets were drained", i)
		}
		if got := list.CopyPacketsA(ind, Kind1); !bytes.Equal(got, want) {
			t.Fatalf("iteration %d: CopyPacketsA() = %v, want %v", i, got, want)
		}
	}
}
