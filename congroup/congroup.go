package congroup

import (
	"context"

	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

type ConGroup struct {
	errGroup *errgroup.Group

	limiter   *rate.Limiter
	waitQueue []func() error
}

func WithContext(ctx context.Context) (*ConGroup, context.Context) {
	g, ctx := errgroup.WithContext(ctx)
	return &ConGroup{errGroup: g}, ctx
}

func (g *ConGroup) SetLimiter(lim *rate.Limiter) {
	g.limiter = lim
}

func (g *ConGroup) Wait(ctx context.Context) error {
	for _, f := range g.waitQueue {
		if g.limiter != nil && !g.limiter.Allow() {
			err := g.limiter.Wait(ctx)
			if err != nil {
				return err
			}
		}
		g.errGroup.Go(f)
	}

	return g.errGroup.Wait()
}

func (g *ConGroup) Go(f func() error) {
	if g.limiter != nil && !g.limiter.Allow() {
		g.waitQueue = append(g.waitQueue, f)
		return
	}

	if len(g.waitQueue) > 0 {
		t := g.waitQueue[0]
		g.waitQueue[0] = f
		f = t
	}

	g.errGroup.Go(f)
}
