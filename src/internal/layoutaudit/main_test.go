package main

import (
	"go/token"
	"go/types"
	"reflect"
	"testing"
)

func TestNames(t *testing.T) {
	got := names(" Player, Object,Player, ,NPC ")
	want := []string{"NPC", "Object", "Player"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("names = %q, want %q", got, want)
	}
}

func TestStructType(t *testing.T) {
	st := types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, nil, "Field", types.Typ[types.Uint32], false),
	}, nil)
	if got, err := structType(st); err != nil || got != st {
		t.Fatalf("structType = (%v, %v), want (%v, nil)", got, err, st)
	}
	if _, err := structType(types.Typ[types.Uint32]); err == nil {
		t.Fatal("non-struct type accepted")
	}
}
