package randhelper

import (
	"math/rand"
)

// GenerateDistinctRandomNumbers generates N distinct random numbers between min and max (inclusive).
func GenerateDistinctRandomNumbers(N int, min, max int64) []int64 {
	if max <= min || N < 0 || int64(N) > (max-min+1) {
		return nil
	}
	distinctNumbers := make(map[int64]bool)
	var numbers []int64

	for len(distinctNumbers) < N {
		num := rand.Int63n(max-min+1) + min
		if !distinctNumbers[num] {
			distinctNumbers[num] = true
			numbers = append(numbers, num)
		}
	}

	return numbers
}
