package opennox

import (
	"math"

	"github.com/opennox/opennox/v1/client"
)

// updateMeteorDrawable4CCD00 replaces the PE32 callback, which received the
// drawable pointer in an int before updating its vertical motion.
func updateMeteorDrawable4CCD00(dr *client.Drawable, frame func() uint32) int {
	effect := dr.UnionEffect()
	height := math.Float32frombits(effect.Field_109)
	velocity := math.Float32frombits(effect.Field_110)
	for i := effect.Field_108; i < frame(); i++ {
		if height > 0 {
			height += velocity
			velocity -= 1
		}
		if height <= 0 {
			height = 0
			velocity = 0
		}
	}
	effect.Field_109 = math.Float32bits(height)
	effect.Field_110 = math.Float32bits(velocity)
	dr.ZVal = uint16(int64(height))
	dr.VelZ = int8(int64(velocity))
	effect.Field_108 = frame()
	return 1
}

func (c *Client) updateMeteorDrawable4CCD00(dr *client.Drawable) int {
	return updateMeteorDrawable4CCD00(dr, c.srv.Frame)
}

// updateFistDrawable4CCDB0 replaces the PE32 callback, which received the
// drawable pointer in an int before updating its vertical motion.
func updateFistDrawable4CCDB0(dr *client.Drawable, frame func() uint32) int {
	effect := dr.UnionEffect()
	height := math.Float32frombits(effect.Field_109)
	velocity := math.Float32frombits(effect.Field_110)
	bounce := math.Float32frombits(effect.Field_111)
	for i := effect.Field_108; i < frame(); i++ {
		nextHeight := float64(height) + float64(velocity)
		height = float32(nextHeight)
		if nextHeight >= 0 {
			velocity = float32(float64(velocity) - 0.5)
		} else {
			nextVelocity := -float64(velocity) * float64(bounce) * 0.1
			height = 0
			velocity = float32(nextVelocity)
			if nextVelocity < 2 {
				height = 0
				velocity = 0
			}
		}
	}
	effect.Field_109 = math.Float32bits(height)
	effect.Field_110 = math.Float32bits(velocity)
	dr.ZVal = uint16(int64(height))
	dr.VelZ = int8(int64(velocity))
	effect.Field_108 = frame()
	return 1
}

func (c *Client) updateFistDrawable4CCDB0(dr *client.Drawable) int {
	return updateFistDrawable4CCDB0(dr, c.srv.Frame)
}
