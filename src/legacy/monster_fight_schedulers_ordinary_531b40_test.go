package legacy

import (
	"testing"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

type schedOrdinaryServer531B40 struct {
	Server
	server *server.Server
}

func (s *schedOrdinaryServer531B40) S() *server.Server { return s.server }

// Exercise the actual public C wrappers without a runtime service provider.
// These original rows neither call related spells nor cancel a started action.
// The complete original output shadows, push implementation and C return ABI
// remain the same as in the separately frozen full comparator.
func TestFightSchedulersOrdinaryPublicC531B40(t *testing.T) {
	rows := schedLoad531B40(t)
	f := schedNewFixture531B40(t)
	old := GetServer
	GetServer = func() Server { return &schedOrdinaryServer531B40{server: f.server} }
	t.Cleanup(func() { GetServer = old })
	debug := noxflags.GetEngine() & noxflags.EngineShowAI
	noxflags.UnsetEngine(noxflags.EngineShowAI)
	t.Cleanup(func() { noxflags.SetEngine(debug) })
	count := 0
	for i := range rows {
		row := &rows[i]
		if row.Input.number("started") != 0 || row.Input.number("alt_started") != 0 ||
			row.Entry == 0x531b40 && row.Input.number("buffs")&(1<<29) == 0 {
			continue
		}
		count++
		t.Run(row.Name, func(t *testing.T) {
			f.reset(row)
			result := schedExecute531B40(f, row.Entry)
			f.validate(t, result, false)
		})
	}
	if count == 0 {
		t.Fatal("no ordinary-runtime original contexts")
	}
	t.Logf("original scheduler ordinary public C contexts=%d", count)
}
