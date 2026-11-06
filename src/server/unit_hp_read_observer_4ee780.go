package server

// UnitHPRead4EE780 observes only reads performed by the unchanged current and
// maximum HP getter algorithms. Field 0 is Unit.HealthData, 1 Health.Cur,
// and 2 Health.Max. Entry precedes the actual read, including a fault.
type UnitHPRead4EE780 struct {
	Unit        *Object
	Health      *HealthData
	Field       int
	Complete    bool
	ValueHealth *HealthData
	Value       uint16
}

// ObserveUnitHPRead4EE780 is nil during ordinary production execution.
var ObserveUnitHPRead4EE780 func(UnitHPRead4EE780)
