package server

// These operations preserve the original PC53 arithmetic and binary32 stores.
// Production uses startup ToZero; Nearest is retained for the independently
// executed alternate-control-word audit.
func monsterRandomWalkMul53_545020(a, b float64, nearest bool) float64 {
	if nearest {
		return objectForceResult53_52DF80(a, b, float64(a*b))
	}
	return objectForceMul53_52DF80(a, b)
}

func monsterRandomWalkAdd53_545020(a, b float64, nearest bool) float64 {
	if nearest {
		return objectForceResult53_52DF80(a, b, float64(a+b))
	}
	return objectForceAdd53_52DF80(a, b)
}

func monsterRandomWalkSpill545020(value float64, nearest bool) float32 {
	if nearest {
		return float32(value)
	}
	return experienceSpill32TowardZero(value)
}
