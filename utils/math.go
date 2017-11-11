package utils

// IntPow returns the x ** y
func IntPow(x, y uint64) uint64 {
	var pow uint64 = 1
	for i := uint64(0); i < y; i++ {
		pow *= x
	}
	return pow
}
