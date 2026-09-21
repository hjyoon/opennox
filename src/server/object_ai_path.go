package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const aiMapIndexSize = WallGridSize

const aiPathDangerousMargin50B2C0 = float32(11.5)

var aiPathReconstructionInset50C320 = math.Float32frombits(0x40133333)

type AIVisitNode struct {
	X0      uint16       // 0, 0
	Y2      uint16       // 0, 0
	Field4  *AIVisitNode // 1, 4
	Field8  *AIVisitNode // 2, 8
	Flags12 byte         // 3, 12
	Field13 byte         // 3, 13
	Field14 uint16       // 3, 14
}

type AIMapIndexFlags uint16

const (
	AIIndexOccupied      = AIMapIndexFlags(0x1)
	AIIndexOccupiedTall  = AIMapIndexFlags(0x2)
	AIIndexElevator      = AIMapIndexFlags(0x4)
	AIIndexElevatorShaft = AIMapIndexFlags(0x8)
	AIIndexHole          = AIMapIndexFlags(0x10)
	AIIndexTransporter   = AIMapIndexFlags(0x20)
	AIIndexWaypoint      = AIMapIndexFlags(0x40)
	AIIndexObject        = AIMapIndexFlags(0x100)
	AIIndexObjectTall    = AIMapIndexFlags(0x200)
	AIIndexFire          = AIMapIndexFlags(0x400)
)

// aiPathCellSamples50AFA0 is the literal 3x3 probe table at GAME.EXE
// 005C0278. The high sample is intentionally one ULP below the Go float32
// conversion of decimal 20.7, so preserve the executable's raw value.
var aiPathCellSamples50AFA0 = [...]types.Pointf{
	{X: math.Float32frombits(0x40133333), Y: math.Float32frombits(0x40133333)},
	{X: math.Float32frombits(0x41380000), Y: math.Float32frombits(0x40133333)},
	{X: math.Float32frombits(0x41a59999), Y: math.Float32frombits(0x40133333)},
	{X: math.Float32frombits(0x40133333), Y: math.Float32frombits(0x41380000)},
	{X: math.Float32frombits(0x41380000), Y: math.Float32frombits(0x41380000)},
	{X: math.Float32frombits(0x41a59999), Y: math.Float32frombits(0x41380000)},
	{X: math.Float32frombits(0x40133333), Y: math.Float32frombits(0x41a59999)},
	{X: math.Float32frombits(0x41380000), Y: math.Float32frombits(0x41a59999)},
	{X: math.Float32frombits(0x41a59999), Y: math.Float32frombits(0x41a59999)},
}

type AIMapIndexNode struct {
	Index0    uint32
	IndexGen4 uint32
	Flags8    AIMapIndexFlags
	Field10   uint16
}

type serverAIPaths struct {
	s            *Server
	mapIndex     [aiMapIndexSize][aiMapIndexSize]AIMapIndexNode
	MapIndexLast uint32
	mapIndexGen  uint32

	allocVisit alloc.ClassT[AIVisitNode]
	points     []types.Pointf
	pointsCnt  int
	PathStatus int
	lastFrame  uint32
	calculated bool
}

func (s *serverAIPaths) Valid() bool {
	return s.allocVisit.Class != nil
}

func (s *serverAIPaths) ResetVisitNodes() {
	s.allocVisit.FreeAllObjects()
}

func (s *serverAIPaths) NewVisitNode() *AIVisitNode {
	return s.allocVisit.NewObject()
}

func (s *serverAIPaths) ResetIndex() {
	s.MapIndexLast = 0
	s.mapIndexGen = 0
	s.mapIndex = [aiMapIndexSize][aiMapIndexSize]AIMapIndexNode{}
}

func (s *serverAIPaths) PathFindStatus() int {
	return s.PathStatus
}

func (s *serverAIPaths) MapIndex(x, y int) *AIMapIndexNode {
	if x < 0 || x >= aiMapIndexSize || y < 0 || y >= aiMapIndexSize {
		return nil
	}
	return &s.mapIndex[y][x]
}

// MapIndexFlags implements GAME.EXE sub_50AB50. The map record is a fixed
// 12-byte value on every host, while the accessor keeps the original
// out-of-range zero result without exposing the backing array to legacy C.
func (s *serverAIPaths) MapIndexFlags(x, y int) AIMapIndexFlags {
	p := s.MapIndex(x, y)
	if p == nil {
		return 0
	}
	return p.Flags8
}

// CheckIndexFlags implements GAME.EXE 0050B950. Ground objects use the low
// occupancy bit, while airborne objects use the adjacent tall-occupancy bit;
// the later hole bit is unrelated to either height.
func (s *serverAIPaths) CheckIndexFlags(obj *Object, x, y int) bool {
	p := s.MapIndex(x, y)
	if obj.Flags().Has(object.FlagAirborne) {
		return p.Flags8&AIIndexOccupiedTall != 0
	}
	return p.Flags8&AIIndexOccupied != 0
}

func (s *serverAIPaths) ResetPoints() {
	s.pointsCnt = 0
}

func (s *serverAIPaths) appendPoint(p types.Pointf) bool {
	if s.pointsCnt >= len(s.points) {
		return false
	}
	s.points[s.pointsCnt] = p
	s.pointsCnt++
	return true
}

func (s *serverAIPaths) swapPoints() {
	for i := 0; i < s.pointsCnt/2; i++ {
		p1 := &s.points[i]
		p2 := &s.points[s.pointsCnt-i-1]
		*p1, *p2 = *p2, *p1
	}
}

func (s *serverAIPaths) Points() []types.Pointf {
	return s.points[:s.pointsCnt]
}

func (s *serverAIPaths) MaybeAppendWorkPath(path []types.Pointf) int {
	if s.pointsCnt > len(path) && s.PathStatus == 0 {
		s.PathStatus = 1
	}
	return s.appendWorkPath(path, 0)
}

func (s *serverAIPaths) appendWorkPath(path []types.Pointf, ind int) int {
	v3 := 0
	if s.pointsCnt <= 0 {
		return ind
	}
	v4 := ind
	v5 := path[ind:]
	for {
		if v4 == len(path)-1 {
			if noxflags.HasEngine(noxflags.EngineShowAI) {
				ai.Log.Printf("appendWorkPath: Path truncated.\n")
			}
			return v4
		}
		v4++
		v6 := s.points[v3]
		v3++
		v5[0] = v6
		v5 = v5[1:]
		if v3 >= s.pointsCnt {
			return v4
		}
	}
}

// Init implements the storage setup from GAME.EXE 0050AB90. Visit nodes use
// native pointer fields, so the allocation class deliberately uses the host
// size of AIVisitNode instead of the original PE32 size of 16 bytes.
func (s *serverAIPaths) Init(srv *Server) {
	s.s = srv
	s.allocVisit = alloc.NewClassT("VisitNodes", AIVisitNode{}, 1024)
	s.points, _ = alloc.Make([]types.Pointf{}, 1024)
	s.lastFrame = 0
}
func (s *serverAIPaths) Sub_50B500() {
	s.calculated = false
}
func (s *serverAIPaths) Sub_50B510() {
	s.calculated = false
	s.lastFrame = 0
}

// Free implements GAME.EXE 0050ABF0. In particular, clear the typed allocator
// handle after releasing it: ClassT.Free has a value receiver and therefore
// cannot clear this owner field by itself.
func (s *serverAIPaths) Free() {
	if s.points != nil {
		alloc.FreeSlice(s.points)
	}
	s.points = nil
	s.allocVisit.Free()
	s.allocVisit = alloc.ClassT[AIVisitNode]{}
}

// sub50B8E0 implements the current-generation dynamic object and fire check
// from GAME.EXE 0050B8E0.
func (s *serverAIPaths) sub50B8E0(obj *Object, x, y int) uint32 {
	p := s.MapIndex(x, y)
	if p.IndexGen4 != s.mapIndexGen {
		return 0
	}
	if obj.Flags().Has(object.FlagAirborne) {
		return uint32((p.Flags8 & AIIndexObjectTall) >> 9)
	}
	if p.Flags8&AIIndexObject != 0 {
		return 1
	}
	if (obj.ObjSubClass>>10)&1 != 0 || p.Flags8&AIIndexFire == 0 {
		return 0
	}
	return 1
}

func (s *serverAIPaths) Nox_xxx_pathfind_preCheckWalls2_50B8A0(obj *Object, x, y int) bool {
	if s.CheckIndexFlags(obj, x, y) {
		return false
	}
	return s.sub50B8E0(obj, x, y) == 0
}

func (s *serverAIPaths) MaybeIndexObjects() {
	if s.calculated {
		return
	}
	s.IndexObjects()
}

func (s *serverAIPaths) IndexObjects() {
	frame := s.s.Frame()
	if (frame - s.lastFrame) < 15 {
		return
	}
	s.lastFrame = frame
	s.mapIndexGen++
	for it := s.s.Objs.List; it != nil; it = it.Next() {
		s.IndexObject(it)
	}
	s.calculated = true
}

func (s *serverAIPaths) IndexObject(obj *Object) {
	class := obj.Class()
	flags := obj.Flags()
	dangerous := class.Has(object.ClassDangerous)
	if class.HasAny(object.ClassDoor | object.ClassElevator | object.ClassElevatorShaft) {
		return
	}
	if !class.Has(object.ClassFire) && flags.HasAny(object.FlagBelow|object.FlagAllowOverlap|object.FlagNoCollide) {
		return
	}
	if !class.HasAny(object.ClassFire|object.ClassObstacle) && !flags.Has(object.FlagNoUpdate) {
		return
	}

	var (
		shape  Shape
		zSize1 float32
		zSize2 float32
	)
	if dangerous {
		// GAME.EXE copies exactly Shape plus the two Z-size words (60 bytes).
		// A raw +172 copy is valid only for PE32 and corrupts the native object
		// layout once pointers widen, so keep the same fields by type instead.
		shape, zSize1, zSize2 = obj.Shape, obj.ZSize1, obj.ZSize2
		aiPathExpandDangerousShape50B2C0(&obj.Shape)
		obj.Nox_xxx_objectUnkUpdateCoords_4E7290()
	}

	minX := aiPathGridCell50AFA0(obj.CollideP1.X)
	minY := aiPathGridCell50AFA0(obj.CollideP1.Y)
	maxX := aiPathGridCell50AFA0(obj.CollideP2.X)
	maxY := aiPathGridCell50AFA0(obj.CollideP2.Y)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			cell := s.MapIndex(int(x), int(y))
			for _, sample := range aiPathCellSamples50AFA0 {
				point := types.Pointf{
					X: float32(x)*23 + sample.X,
					Y: float32(y)*23 + sample.Y,
				}
				if !obj.Sub547DB0(&point) {
					continue
				}
				cell.IndexGen4 = s.mapIndexGen
				class = obj.Class()
				flags = obj.Flags()
				if class.Has(object.ClassObstacle) || flags.Has(object.FlagNoUpdate) {
					cell.Flags8 |= AIIndexObject
					if !flags.Has(object.FlagShort) {
						cell.Flags8 |= AIIndexObjectTall
					}
				} else if class.Has(object.ClassFire) {
					cell.Flags8 |= AIIndexFire
				}
				break
			}
		}
	}

	if dangerous {
		obj.Shape, obj.ZSize1, obj.ZSize2 = shape, zSize1, zSize2
		obj.Nox_xxx_objectUnkUpdateCoords_4E7290()
	}
}

func aiPathExpandDangerousShape50B2C0(shape *Shape) {
	switch shape.Kind {
	case ShapeKindCircle:
		// The original x87 sequence keeps the radius addition in extended
		// precision while calculating R2, even though it stores the rounded
		// binary32 radius first. Binary64 exactly represents this operation for
		// binary32 inputs and preserves that observable one-ULP distinction.
		radius := float64(shape.Circle.R) + float64(aiPathDangerousMargin50B2C0)
		shape.Circle.R = float32(radius)
		shape.Circle.R2 = float32(radius * radius)
	case ShapeKindBox:
		shape.Box.W += 2 * aiPathDangerousMargin50B2C0
		shape.Box.H += 2 * aiPathDangerousMargin50B2C0
		shape.Box.Calc()
	}
}

func (s *serverAIPaths) sub50AEA0(node *AIVisitNode, outPos *types.Pointf, outNetCode *uint32) bool {
	x := int32(node.X0)
	y := int32(node.Y2)
	p := s.MapIndex(int(x), int(y))
	if (p.Flags8 & 0x3C) == 0 {
		return false
	}
	var filter object.Class
	if p.Flags8&AIIndexHole != 0 {
		filter = object.ClassHole
	} else if p.Flags8&AIIndexTransporter != 0 {
		filter = object.ClassTransporter
	} else if p.Flags8&AIIndexElevator != 0 {
		filter = object.ClassElevator
	} else if p.Flags8&AIIndexElevatorShaft != 0 {
		filter = object.ClassElevatorShaft
	}
	var pos types.Pointf
	pos.X = float32(float64(x) * 23.0)
	pos.Y = float32(float64(y) * 23.0)
	obj := s.s.Map.Sub517B70byClass(pos, filter)
	if obj == nil || !obj.Flags().Has(object.FlagEnabled) {
		return false
	}
	*outPos = obj.PosVec
	*outNetCode = obj.NetCode
	return true
}

func (s *serverAIPaths) Sub_50C320(obj *Object, node *AIVisitNode, start *types.Pointf) {
	ud := obj.UpdateDataMonster()
	s.ResetPoints()
	if node == nil {
		return
	}
	if start != nil {
		s.appendPoint(*start)
	} else {
		var (
			netCode uint32
			out     types.Pointf
		)
		if node.Flags12&0x2 != 0 && s.sub50AEA0(node, &out, &netCode) {
			s.appendPoint(out)
			if int(ud.Field543_0) < len(ud.Field535) {
				ud.Field535[ud.Field543_0] = netCode
				ud.Field543_0++
			}
		} else {
			s.appendPoint(types.Pointf{
				X: float32(float64(node.X0)*23.0 + 11.5),
				Y: float32(float64(node.Y2)*23.0 + 11.5),
			})
		}
	}
	prev := node
	for it := node.Field4; it != nil; it, prev = it.Field4, it {
		dx := int32(it.X0) - int32(prev.X0)
		dy := int32(it.Y2) - int32(prev.Y2)
		if it.Flags12&0x2 != 0 {
			var (
				netCode uint32
				out     types.Pointf
			)
			if s.sub50AEA0(it, &out, &netCode) {
				s.appendPoint(out)
				if int32(ud.Field543_0) < 8 {
					if dx*dx+dy*dy > 4761 {
						ud.Field535[ud.Field543_0] = netCode
						ud.Field543_0++
					} else {
						var ppa, ppb types.Pointf
						ppa.X = float32(int32(it.X0)*23 + 11)
						ppa.Y = float32(int32(it.Y2)*23 + 11)
						ppb.X = float32(int32(prev.X0)*23 + 11)
						ppb.Y = float32(int32(prev.Y2)*23 + 11)
						if !s.s.MapTraceRayAt(ppa, ppb, nil, nil, 1) {
							ud.Field535[ud.Field543_0] = netCode
							ud.Field543_0++
						}
					}
				}
				continue
			}
		}
		if dx < 0 {
			if dy > 0 {
				s.appendPoint(types.Pointf{
					X: float32(float64(it.X0)*23.0 + 23.0 - float64(aiPathReconstructionInset50C320)),
					Y: float32(float64(it.Y2)*23.0 + float64(aiPathReconstructionInset50C320)),
				})
			} else if dy < 0 {
				s.appendPoint(types.Pointf{
					X: float32(float64(it.X0)*23.0 + 23.0 - float64(aiPathReconstructionInset50C320)),
					Y: float32(float64(it.Y2)*23.0 + 23.0 - float64(aiPathReconstructionInset50C320)),
				})
			} else {
				s.appendPoint(types.Pointf{
					X: float32(float64(it.X0)*23.0 + 23.0 - float64(aiPathReconstructionInset50C320)),
					Y: float32(float64(it.Y2)*23.0 + 11.5),
				})
			}
		} else if dx == 0 {
			if dy > 0 {
				s.appendPoint(types.Pointf{
					X: float32(float64(it.X0)*23.0 + 11.5),
					Y: float32(float64(it.Y2)*23.0 + float64(aiPathReconstructionInset50C320)),
				})
			} else if dy < 0 {
				s.appendPoint(types.Pointf{
					X: float32(float64(it.X0)*23.0 + 11.5),
					Y: float32(float64(it.Y2)*23.0 + 23.0 - float64(aiPathReconstructionInset50C320)),
				})
			} else if s.pointsCnt < len(s.points) {
				// GAME.EXE increments the point index even when a special link
				// resolves to the same grid cell. Preserve the existing slot so
				// the final count and reversal match that degenerate path.
				s.pointsCnt++
			}
		} else {
			if dy < 0 {
				s.appendPoint(types.Pointf{
					X: float32(float64(it.X0)*23.0 + float64(aiPathReconstructionInset50C320)),
					Y: float32(float64(it.Y2)*23.0 + 23.0 - float64(aiPathReconstructionInset50C320)),
				})
			} else if dy == 0 {
				s.appendPoint(types.Pointf{
					X: float32(float64(it.X0)*23.0 + float64(aiPathReconstructionInset50C320)),
					Y: float32(float64(it.Y2)*23.0 + 11.5),
				})
			} else {
				s.appendPoint(types.Pointf{
					X: float32(float64(it.X0)*23.0 + float64(aiPathReconstructionInset50C320)),
					Y: float32(float64(it.Y2)*23.0 + float64(aiPathReconstructionInset50C320)),
				})
			}
		}
	}
	s.swapPoints()
}

func (s *serverAIPaths) HasNoEnemiesAround(obj *Object, x, y int) bool {
	found := false
	var pos types.Pointf
	pos.X = float32(float64(x)*23.0 + 11.5)
	pos.Y = float32(float64(y)*23.0 + 11.5)
	s.s.Map.EachObjInCircle(pos, 100.0, func(it *Object) bool {
		if !found && s.s.IsEnemyTo(obj, it) {
			found = true
		}
		return true
	})
	return !found
}

func (s *serverAIPaths) Sub50AFA0() {
	s.ResetIndex()
	for it := s.s.Objs.List; it != nil; it = it.Next() {
		s.IndexObject(it)

		class := it.Class()
		if class.Has(object.ClassDoor) {
			continue
		}
		x := aiPathGridCell50AFA0(it.PosVec.X)
		y := aiPathGridCell50AFA0(it.PosVec.Y)
		switch {
		case class.Has(object.ClassHole):
			s.MapIndex(int(x), int(y)).Flags8 |= AIIndexHole
		case class.Has(object.ClassTransporter):
			s.MapIndex(int(x), int(y)).Flags8 |= AIIndexTransporter
		case class.Has(object.ClassElevator):
			s.MapIndex(int(x), int(y)).Flags8 |= AIIndexElevator
		case class.Has(object.ClassElevatorShaft):
			s.MapIndex(int(x), int(y)).Flags8 |= AIIndexElevatorShaft
		case class.Has(object.ClassImmobile):
			flags := it.Flags()
			if flags.HasAny(object.FlagBelow | object.FlagAllowOverlap | object.FlagNoCollide) {
				continue
			}
			minX := aiPathGridCell50AFA0(it.CollideP1.X)
			minY := aiPathGridCell50AFA0(it.CollideP1.Y)
			maxX := aiPathGridCell50AFA0(it.CollideP2.X)
			maxY := aiPathGridCell50AFA0(it.CollideP2.Y)
			for cy := minY; cy <= maxY; cy++ {
				for cx := minX; cx <= maxX; cx++ {
					cell := s.MapIndex(int(cx), int(cy))
					for _, sample := range aiPathCellSamples50AFA0 {
						point := types.Pointf{
							X: float32(cx)*23 + sample.X,
							Y: float32(cy)*23 + sample.Y,
						}
						if !it.Sub547DB0(&point) {
							continue
						}
						cell.Flags8 |= AIIndexOccupied
						if !flags.Has(object.FlagShort) {
							cell.Flags8 |= AIIndexOccupiedTall
						}
						break
					}
				}
			}
		}
	}
	for wp := s.s.WPs.First(); wp != nil; wp = wp.WpNext {
		if wp.HasFlag2Mask(0x80) {
			x := aiPathGridCell50AFA0(wp.PosVec.X)
			y := aiPathGridCell50AFA0(wp.PosVec.Y)
			s.MapIndex(int(x), int(y)).Flags8 |= AIIndexWaypoint
		}
	}
}

const aiPathGridInverse50AC20 = float32(1.0 / 23.0)

// aiPathHoleDestinationValid50AC20 preserves the wrapped PE32 check at
// GAME.EXE 0050ACCD. The original adds DestinationX's offset to CollideData
// and rejects the result only when that pointer-width addition becomes zero.
func aiPathHoleDestinationValid50AC20(data uintptr) bool {
	return data+unsafe.Offsetof(HoleCollideData{}.DestinationX) != 0
}

// aiPathTargetCell50AC20 models the binary32 multiply and nox_float2int call
// used by GAME.EXE 0050ADB4 and 0050AE32. The latter is x87 FISTP under the
// default round-to-nearest-even mode and returns integer-indefinite on invalid
// or out-of-range input; the caller stores its low word.
func aiPathFloatToInt419A70(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

func aiPathGridCell50AFA0(value float32) int32 {
	return aiPathFloatToInt419A70(value * aiPathGridInverse50AC20)
}

// GridCell50B810 converts a world position to the path-grid coordinates used
// by GAME.EXE 0050B810. Both products are spilled to binary32 before the
// shared 00419A70 x87 round-to-nearest-even conversion.
func (s *serverAIPaths) GridCell50B810(pos *types.Pointf) (int, int) {
	return int(aiPathGridCell50AFA0(pos.X)), int(aiPathGridCell50AFA0(pos.Y))
}

func aiPathTargetCell50AC20(value float32) uint16 {
	return uint16(aiPathGridCell50AFA0(value))
}

func (s *serverAIPaths) Sub_50AC20(node *AIVisitNode, out *[2]uint16) int32 {
	x := int32(node.X0)
	y := int32(node.Y2)
	p := s.MapIndex(int(x), int(y))
	if (p.Flags8 & 0x3C) == 0 {
		return 0
	}
	var pos types.Pointf
	pos.X = float32(float64(x) * 23.0)
	pos.Y = float32(float64(y) * 23.0)
	if p.Flags8&AIIndexHole != 0 {
		obj := s.s.Map.Sub517B70byClass(pos, object.ClassHole)
		if obj == nil {
			return 0
		}
		if !aiPathHoleDestinationValid50AC20(uintptr(obj.CollideData)) {
			return 0
		}
		if obj.Flags().Has(object.FlagEnabled) {
			data := (*HoleCollideData)(obj.CollideData)
			out[0] = uint16(data.DestinationX / 23)
			out[1] = uint16(data.DestinationY / 23)
			return 1
		}
		return 0
	} else if p.Flags8&AIIndexTransporter != 0 {
		obj := s.s.Map.Sub517B70byClass(pos, object.ClassTransporter)
		if obj == nil {
			return 0
		}
		targ := obj.TransporterTarget()
		if targ != nil && obj.Flags().Has(object.FlagEnabled) {
			out[0] = aiPathTargetCell50AC20(targ.PosVec.X)
			out[1] = aiPathTargetCell50AC20(targ.PosVec.Y)
			return 1
		}
		return 0
	} else if p.Flags8&AIIndexElevator != 0 {
		obj := s.s.Map.Sub517B70byClass(pos, object.ClassElevator)
		if obj == nil {
			return 0
		}
		targ := obj.ElevatorLink()
		if targ != nil && obj.Flags().Has(object.FlagEnabled) {
			out[0] = aiPathTargetCell50AC20(targ.PosVec.X)
			out[1] = aiPathTargetCell50AC20(targ.PosVec.Y)
			return 1
		}
		return 0
	} else if p.Flags8&AIIndexElevatorShaft != 0 {
		obj := s.s.Map.Sub517B70byClass(pos, object.ClassElevatorShaft)
		if obj == nil {
			return 0
		}
		targ := obj.ElevatorLink()
		if targ != nil && obj.Flags().Has(object.FlagEnabled) {
			out[0] = aiPathTargetCell50AC20(targ.PosVec.X)
			out[1] = aiPathTargetCell50AC20(targ.PosVec.Y)
			return 1
		}
		return 0
	} else {
		return 0
	}
}
