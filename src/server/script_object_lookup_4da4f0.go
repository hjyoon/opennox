package server

import "strings"

// FindByScriptName searches active objects before pending objects, including
// each object's immediate inventory.
func (s *serverObjects) FindByScriptName(name string) *Object {
	qualified := strings.Contains(name, ":")
	for _, head := range [2]*Object{s.First(), s.Pending} {
		for obj := head; obj != nil; obj = obj.Next() {
			if found := obj.FindByScriptName(name, qualified); found != nil {
				return found
			}
		}
	}
	return nil
}

func (obj *Object) FindByScriptName(name string, qualified bool) *Object {
	if obj.matchesScriptName(name, qualified) {
		return obj
	}
	for item := obj.InvFirstItem; item != nil; item = item.InvNextItem {
		if item.matchesScriptName(name, qualified) {
			return item
		}
	}
	return nil
}

func (obj *Object) matchesScriptName(name string, qualified bool) bool {
	if obj.IDPtr == nil {
		return false
	}
	id := obj.ID()
	if !qualified {
		if _, suffix, ok := strings.Cut(id, ":"); ok {
			id = suffix
		}
	}
	return id == name
}
