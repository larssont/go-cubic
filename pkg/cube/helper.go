package cube

import (
	"log"
	"math"
	"strconv"
	"time"
)

func reverse[T any](original []T) (reversed []T) {
	reversed = make([]T, len(original))
	copy(reversed, original)

	for i := len(reversed)/2 - 1; i >= 0; i-- {
		tmp := len(reversed) - 1 - i
		reversed[i], reversed[tmp] = reversed[tmp], reversed[i]
	}

	return
}

func isEqualLength[T any](slices ...[]T) bool {
	if len(slices) == 0 {
		return true
	}
	length := len(slices[0])
	for _, slice := range slices[1:] {
		if len(slice) != length {
			return false
		}
	}
	return true
}
func isPerfectSquare(x int) bool {
	if x < 0 {
		return false
	}
	s := int(math.Sqrt(float64(x)))
	return s*s == x
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %d ns", name, elapsed.Nanoseconds())
}

func pow(v, exp int) int {
	x := 1
	for range exp {
		x *= v
	}
	return x
}

func strToInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return value, nil
}
