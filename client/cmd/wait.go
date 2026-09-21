package cmd

import (
	"context"
	"fmt"
	"math"
	"time"
)

func secondsDuration(flag string, seconds int, allowZero bool) (time.Duration, error) {
	if seconds < 0 || (!allowZero && seconds == 0) || int64(seconds) > math.MaxInt64/int64(time.Second) {
		qualifier := "positive"
		if allowZero {
			qualifier = "non-negative"
		}
		return 0, fmt.Errorf("--%s must be a valid %s number of seconds", flag, qualifier)
	}
	return time.Duration(seconds) * time.Second, nil
}

func wait(ctx context.Context, interval time.Duration) error {
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
