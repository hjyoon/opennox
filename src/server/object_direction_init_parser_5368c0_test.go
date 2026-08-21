package server

import (
	"testing"
	"unsafe"
)

func TestObjectDirectionAtoi5368C0MatchesWrappingPrefixConversion(t *testing.T) {
	tests := []struct {
		input string
		want  int32
	}{
		{"", 0},
		{"words", 0},
		{" \t\r\n+17tail", 17},
		{"-1", -1},
		{"2147483648", -2147483648},
		{"4294967295", -1},
		{"4294967296", 0},
		{"-4294967295", 1},
	}
	for _, tc := range tests {
		if got := objectDirectionAtoi5368C0(tc.input); got != tc.want {
			t.Errorf("atoi(%q) = %d (%#x), want %d (%#x)", tc.input, got, uint32(got), tc.want, uint32(tc.want))
		}
	}
}

func TestObjectDirectionInitParseStoresInOrderAndIgnoresExtraTokens(t *testing.T) {
	data := DirectionInitData{X: 111, Y: 222}
	typ := &ObjectType{InitData: unsafe.Pointer(&data)}
	if err := objectDirectionInitParse(typ, []string{"-1tail", "+4294967295junk", "ignored"}); err != nil {
		t.Fatal(err)
	}
	if data.X != -1 || data.Y != -1 {
		t.Fatalf("data = %+v, want {-1 -1}", data)
	}
}

func TestObjectDirectionInitParseMissingSecondPreservesFirstStore(t *testing.T) {
	data := DirectionInitData{X: 111, Y: 222}
	typ := &ObjectType{InitData: unsafe.Pointer(&data)}
	if err := objectDirectionInitParse(typ, []string{"-7"}); err == nil {
		t.Fatal("missing second value accepted")
	}
	if data.X != -7 || data.Y != 222 {
		t.Fatalf("partial data = %+v, want {-7 222}", data)
	}
}

func TestObjectDirectionInitParseMissingBothPreservesRecord(t *testing.T) {
	data := DirectionInitData{X: 111, Y: 222}
	typ := &ObjectType{InitData: unsafe.Pointer(&data)}
	if err := objectDirectionInitParse(typ, nil); err == nil {
		t.Fatal("missing values accepted")
	}
	if data.X != 111 || data.Y != 222 {
		t.Fatalf("data changed to %+v", data)
	}
}
