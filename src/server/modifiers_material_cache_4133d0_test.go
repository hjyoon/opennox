package server

import (
	"path/filepath"
	"testing"
)

func TestMaterial7CacheServerIsolation4133D0(t *testing.T) {
	first, second := new(Server), new(Server)
	a, b := first.Modif.Material7CacheCell4133D0(), second.Modif.Material7CacheCell4133D0()
	if a == b || *a != nil || *b != nil {
		t.Fatal("each server must own an initially empty native pointer cell")
	}
	value := new(ModifierEff)
	*a = value
	if *b != nil {
		t.Fatal("one server populated another server's descriptor cache")
	}
	first.Modif.Nox_xxx_freeWeaponArmorDefAndModifs_413060()
	if *a != nil || *b != nil {
		t.Fatal("descriptor free must invalidate its cache before releasing definitions")
	}
}

func TestMaterial7CacheFailedReload4133D0(t *testing.T) {
	srv := new(Server)
	cell := srv.Modif.Material7CacheCell4133D0()
	*cell = new(ModifierEff)
	missing := filepath.Join(t.TempDir(), "missing-material-cache-4133d0.bin")
	if err := srv.Modif.Nox_xxx_parseModifierBin_412930(missing); err == nil {
		t.Fatal("missing modifier input unexpectedly loaded")
	}
	if *cell != nil {
		t.Fatal("failed modifier reload retained the old descriptor pointer")
	}
}
