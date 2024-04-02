package concurrency

import (
	"context"
	"fmt"
	"time"
)

func SchduleRecurringTaskInDays(ctx context.Context, name string, noOfDays int, f func()) {
	gap := time.Duration(noOfDays) * 24 * time.Hour
	timer := time.NewTimer(gap)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("stopping recurring task, ", name)
			return
		case <-timer.C:
			f()
			timer.Reset(gap)
		}
	}
}
