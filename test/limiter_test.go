package test

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"sync"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {

	txCount := 20
	txRate := rate.NewLimiter(1, 1)
	var wg sync.WaitGroup

	for i := 0; i < txCount; i++ {
		if err := txRate.Wait(context.Background()); err == nil {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				fmt.Printf("%s TX=%d\n", time.Now(), idx)
			}(i)
		} else {
			fmt.Printf("ERR; ")
		}
	}
	wg.Wait()
	fmt.Printf("--DONE--")
}
