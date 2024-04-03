package concurrency

import (
	"context"
	"fmt"
	"time"
)

func SchduleRecurringTaskInMonth(ctx context.Context, name string, noOfHours, monthDay int, f func()) {
	gap := time.Duration(noOfHours) * time.Hour
	timer := time.NewTimer(gap)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("stopping recurring task, ", name)
			return
		case <-timer.C:
			if time.Now().Day() == monthDay {
				f()
			}
			timer.Reset(gap)
		}
	}
}
