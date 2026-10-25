package balance

import (
	"math"
	"testing"

	"github.com/shoenig/test/must"

	"github.com/opennox/libs/noxtest"
)

func TestReadBalance(t *testing.T) {
	f, err := ReadBalance(noxtest.DataPath(t, GamedataFile))
	must.NoError(t, err)
	must.EqOp(t, math.Float64frombits(0x3fc9999980000000), f.FloatDef("", "BerserkerPainRatio", -1))
	must.EqOp(t, math.Float64frombits(0x3fc9999980000000), f.FloatDef(TagSolo, "BerserkerPainRatio", -1))
	must.EqOp(t, float64(-1), f.FloatDef("", "BerserkerDamage", -1))
	must.EqOp(t, float64(100), f.FloatDef(TagSolo, "BerserkerDamage", -1))
	must.EqOp(t, float64(150), f.FloatDef(TagArena, "BerserkerDamage", -1))
	must.EqOp(t, math.Float64frombits(0x3fe3333320000000), f.FloatDef("", "RepairCoefficient", -1))
	must.EqOp(t, float64(150), f.FloatDef(TagArena, "WarriorMaxHealth", -1))
	must.Eq(t, []float64{math.Float64frombits(0x3fe0000000000000), math.Float64frombits(0x3fe6666660000000), math.Float64frombits(0x3ff0000000000000), math.Float64frombits(0x3ff4ccccc0000000), math.Float64frombits(0x3ff9999980000000)}, f.Array(TagSolo, "EnergyBoltDamage"))
	must.Eq(t, []float64{math.Float64frombits(0x3ff4ccccc0000000), math.Float64frombits(0x3ff9999980000000), math.Float64frombits(0x4000000000000000), math.Float64frombits(0x4002666660000000), math.Float64frombits(0x4004ccccc0000000)}, f.Array(TagArena, "EnergyBoltDamage"))
	must.Eq(t, []float64{
		-1,
		0, 5000, 11000, 18200, 26840,
		37208, 49650, 64580, 82495, 103995,
		126891, 151276, 177246, 204903, 234359,
		265729, 299139, 334720, 372613, 412970,
		455950, 501723, 550472, 602390, 657682,
		716568, 779282, 846072, 917204, 992959,
		2000000000,
	}, f.Array("", "XPTable"))
}
