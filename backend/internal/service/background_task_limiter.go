package service

import "context"

const defaultBackgroundTaskMaxConcurrent = 4

var backgroundTaskSlots = make(chan struct{}, defaultBackgroundTaskMaxConcurrent)

// AcquireBackgroundTaskSlot limits expensive background work so foreground requests
// still have room to use shared DB/network resources.
func AcquireBackgroundTaskSlot(ctx context.Context) (func(), error) {
	select {
	case backgroundTaskSlots <- struct{}{}:
		return func() {
			select {
			case <-backgroundTaskSlots:
			default:
			}
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
