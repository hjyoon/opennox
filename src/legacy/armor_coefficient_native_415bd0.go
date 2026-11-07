package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type armorCoefficientRead415BD0 struct {
	Site    uint32
	Address unsafe.Pointer
	Width   uint32
	Word    uint32
}

type armorCoefficientObserver415BD0 struct {
	Read   func(armorCoefficientRead415BD0)
	Lookup func(uint16) unsafe.Pointer
}

type armorCoefficientObserverFrame415BD0 struct {
	observer *armorCoefficientObserver415BD0
	previous *armorCoefficientObserverFrame415BD0
}

var armorCoefficientObserverStack415BD0 *armorCoefficientObserverFrame415BD0

// Diagnostic scopes are serial and restored in LIFO order.
func pushArmorCoefficientObserver415BD0(observer *armorCoefficientObserver415BD0) func() {
	frame := &armorCoefficientObserverFrame415BD0{observer: observer, previous: armorCoefficientObserverStack415BD0}
	armorCoefficientObserverStack415BD0 = frame
	return func() {
		if armorCoefficientObserverStack415BD0 != frame {
			panic("armor coefficient observer scopes must be restored in LIFO order")
		}
		armorCoefficientObserverStack415BD0 = frame.previous
	}
}

func armorCoefficientObserverCurrent415BD0() *armorCoefficientObserver415BD0 {
	if frame := armorCoefficientObserverStack415BD0; frame != nil {
		return frame.observer
	}
	return nil
}

func armorCoefficientObserveRead415BD0(site uint32, address unsafe.Pointer, width, word uint32) {
	if observer := armorCoefficientObserverCurrent415BD0(); observer != nil && observer.Read != nil {
		observer.Read(armorCoefficientRead415BD0{Site: site, Address: address, Width: width, Word: word})
	}
}

// The complete original 43-byte caller reads class, then the unsigned type
// WORD, then the returned definition's float. Native named offsets replace
// PE32 slots. There is no additional target guard or name-based admission.
func armorCoefficientNative415BD0(item unsafe.Pointer) float64 {
	classAddress := unsafe.Add(item, unsafe.Offsetof(server.Object{}.ObjClass))
	class := *(*uint32)(classAddress)
	armorCoefficientObserveRead415BD0(0x415BD4, classAddress, 4, class)
	if class&0x02000000 == 0 {
		return 0
	}
	typeAddress := unsafe.Add(item, unsafe.Offsetof(server.Object{}.TypeInd))
	typeInd := *(*uint16)(typeAddress)
	armorCoefficientObserveRead415BD0(0x415BDF, typeAddress, 2, uint32(typeInd))
	var definition unsafe.Pointer
	if observer := armorCoefficientObserverCurrent415BD0(); observer != nil && observer.Lookup != nil {
		definition = observer.Lookup(typeInd)
	} else {
		definition = unsafe.Pointer(GetServer().S().Modif.Nox_xxx_equipClothFindDefByTT413270(int(typeInd)))
	}
	if definition == nil {
		return 0
	}
	coefficientAddress := unsafe.Add(definition, unsafe.Offsetof(server.Modifier{}.DamageCoeffOrArmor64))
	word := *(*uint32)(coefficientAddress)
	armorCoefficientObserveRead415BD0(0x415BF0, coefficientAddress, 4, word)
	// Keep the loaded word cached if an observer mutates the definition.
	return float64(*(*float32)(unsafe.Pointer(&word)))
}
