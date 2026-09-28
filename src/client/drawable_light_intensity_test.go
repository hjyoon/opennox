package client

import "testing"

func TestDrawableSetLightIntensityWithoutLegacyBlobInit(t *testing.T) {
	var dr Drawable
	dr.SetLightIntensity(75.5)
	if dr.LightIntensity != 63 || dr.LightIntensityU16 != 63*0x10000 || dr.LightIntensityRad != 181 {
		t.Fatalf("light intensity = %v, fixed %d, radius %d; want 63, %d, 181",
			dr.LightIntensity, dr.LightIntensityU16, dr.LightIntensityRad, 63*0x10000)
	}
}
