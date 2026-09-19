package legacy

import (
	"sync"

	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/server"
)

type questPlayerIdentity509C30 struct {
	name  string
	class player.Class
	id    uint32
}

// questPlayerIdentityStore509C30 replaces GAME.EXE's 32-byte PE32 records.
// The original record starts with three 32-bit intrusive-list words, followed
// by name[12], a uint32 identity and a class byte. Using nox_list_item_t for
// those words makes the payload overlap 64-bit links, so the native store owns
// fixed-width values and never retains a Player or C pointer.
type questPlayerIdentityStore509C30 struct {
	sync.Mutex
	initialized bool
	entries     []questPlayerIdentity509C30
}

func (s *questPlayerIdentityStore509C30) remember(name string, class player.Class, id uint32) {
	s.Lock()
	defer s.Unlock()
	if !s.initialized {
		s.entries = nil
		s.initialized = true
	}

	// Every original record has sort key zero, so sub_4257F0 inserts a new
	// record before the current first record. Preserve that walk order even
	// though the two queries below only expose membership predicates.
	s.entries = append(s.entries, questPlayerIdentity509C30{})
	copy(s.entries[1:], s.entries[:len(s.entries)-1])
	s.entries[0] = questPlayerIdentity509C30{name: name, class: class, id: id}
}

func (s *questPlayerIdentityStore509C30) clear() {
	s.Lock()
	defer s.Unlock()
	if s.initialized {
		s.entries = nil
	}
}

func (s *questPlayerIdentityStore509C30) allows(name string, class player.Class, id uint32) bool {
	s.Lock()
	defer s.Unlock()
	for _, entry := range s.entries {
		if entry.name == name && (entry.class != class || entry.id != id) {
			return false
		}
	}
	return true
}

func (s *questPlayerIdentityStore509C30) contains(name string) bool {
	s.Lock()
	defer s.Unlock()
	for _, entry := range s.entries {
		if entry.name == name {
			return true
		}
	}
	return false
}

var questPlayerIdentities509C30 questPlayerIdentityStore509C30

func Sub_509C30(p *server.Player) {
	questPlayerIdentities509C30.remember(p.Field2096(), p.PlayerClass(), p.Field2068)
}

func Sub_509CB0() {
	questPlayerIdentities509C30.clear()
}

func Sub_509CF0(name string, class player.Class, id uint32) int {
	if questPlayerIdentities509C30.allows(name, class, id) {
		return 1
	}
	return 0
}

func Sub_509D80(p *server.Player) int {
	if questPlayerIdentities509C30.contains(p.Field2096()) {
		return 1
	}
	return 0
}
