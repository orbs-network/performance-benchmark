package benchmark

import "time"

const eventuallyIterations = 50

func Eventually(timeout time.Duration, f func() bool) bool {
	for i := 0; i < eventuallyIterations; i++ {
		if testButDontPanic(f) {
			return true
		}
		time.Sleep(timeout / eventuallyIterations)
	}
	return false
}

func testButDontPanic(f func() bool) bool {
	defer func() { recover() }()
	return f()
}
