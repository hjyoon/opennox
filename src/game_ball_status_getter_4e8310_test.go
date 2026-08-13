package opennox

import "testing"

func TestGameBallStatusGetter4E8310ReturnsExactPointerWithoutMutation(t *testing.T) {
	record := gameBallStatusRecord4E8290{
		State:    0x81,
		Reserved: 0x7b,
		NetCode:  0xbcde,
	}
	want := record
	got := gameBallStatusGetter4E8310(&record)
	if got != &record {
		t.Fatalf("record pointer = %p, want %p", got, &record)
	}
	if record != want {
		t.Fatalf("record = %#v, want unchanged %#v", record, want)
	}
}

func TestGameBallStatusGetter4E8310ReturnsNilWithoutFault(t *testing.T) {
	if got := gameBallStatusGetter4E8310(nil); got != nil {
		t.Fatalf("nil record pointer = %p, want nil", got)
	}
}

func TestGameBallStatusGetter4E8310AliasesLiveRecord(t *testing.T) {
	record := gameBallStatusRecord4E8290{Reserved: 0xa5}
	got := gameBallStatusGetter4E8310(&record)
	record.State = 0xfe
	record.NetCode = 0xabcd
	if got.State != 0xfe || got.NetCode != 0xabcd || got.Reserved != 0xa5 {
		t.Fatalf("returned record = %#v, want live source values", *got)
	}
	got.State = 0x42
	if record.State != 0x42 {
		t.Fatalf("source state = %#x, want returned-pointer write 0x42", record.State)
	}
}
