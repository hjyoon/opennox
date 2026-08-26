package legacy

import (
	"github.com/opennox/noxscript/ns/v4"

	"github.com/opennox/opennox/v1/server"
)

func Nox_xxx_comJournalEntryAdd_427500(a1 *server.Object, msg ns.StringID, a3 ns.EntryType) {
	GetServer().S().JournalEntryAdd427500(a1, string(msg), uint16(a3))
}
