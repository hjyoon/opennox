package legacy

import "github.com/opennox/opennox/v1/internal/oracletest"

func readOriginalNonItemStartupFixture4E17B0(path string) ([]byte, error) {
	return oracletest.ReadNonItemStartup(path, "../server/testdata/nonitem_damage_historical_startup_original.json")
}
