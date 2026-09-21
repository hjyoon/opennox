package server

import "github.com/opennox/libs/object"

// MinimapMonsterIterator50AAE0 keeps the live server object selected by the
// Show-AI minimap traversal. Unlike the original PE32 global, current retains
// the complete native pointer on 64-bit hosts.
type MinimapMonsterIterator50AAE0 struct {
	current *Object
}

func (it *MinimapMonsterIterator50AAE0) selectFrom(obj *Object) *Object {
	for obj != nil {
		it.current = obj
		if obj.Class().Has(object.ClassMonster) {
			return obj
		}
		obj = obj.Next()
	}
	it.current = nil
	return nil
}

// First restarts traversal at the supplied server object-list head.
func (it *MinimapMonsterIterator50AAE0) First(first *Object) *Object {
	return it.selectFrom(first)
}

// Next resumes from the live successor of the previously selected object.
func (it *MinimapMonsterIterator50AAE0) Next() *Object {
	if it.current == nil {
		return nil
	}
	return it.selectFrom(it.current.Next())
}
