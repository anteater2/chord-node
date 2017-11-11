package utils

import (
	"fmt"
)

func ExampleIntPow() {
	fmt.Printf(
		"%d\n%d\n%d",
		IntPow(3, 4),
		IntPow(2, 20),
		IntPow(200, 0))
	// Output: 81
	// 1048576
	// 1
}
