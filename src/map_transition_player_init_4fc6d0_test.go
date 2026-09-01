package opennox

import (
	"testing"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

func TestMapTransitionPlayerInitRoot4FC6D0BindsExactStateAndUnitGates(t *testing.T) {
	for _, tc := range []struct {
		name       string
		initState  int32
		entryState int32
	}{
		{name: "both states zero"},
		{name: "map-init exact one without a unit", initState: 1},
		{name: "entry exact one without a unit", initState: 2, entryState: 1},
		{name: "noncanonical states", initState: -1, entryState: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := server.New(nil, nil, strman.New())
			t.Cleanup(base.Close)
			base.SetMapInitState4FC570(tc.initState)
			base.SetMapEntryState4FC580(tc.entryState)

			s := &Server{Server: base}
			s.mapTransitionPlayerInit4FC6D0()
		})
	}
}
