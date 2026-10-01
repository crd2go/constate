package constate

import (
	"context"
	"errors"
)

// ErrUnsetSelector is returned when a VersionDispatcher has no selector configured.
var ErrUnsetSelector = errors.New("nil selector - please set it")

// Selector inspects the resource and selects the version-specific StateHandler.
type Selector[T any] func(ctx context.Context, obj *T) (StateHandler[T], error)

// VersionDispatcher routes lifecycle calls to a version-specific StateHandler
// chosen at runtime by a Selector. It does not implement For or
// SetupWithManager; embed it in your own dispatcher type and implement those
// registration methods yourself.
//
// Example:
//
//	 ...
//		type MyVersionDispatcher struct {
//		    *constate.VersionDispatcher[MyResource]
//		}
//
//		func NewMyVersionDispatcher() *MyVersionDispatcher {
//		    return &MyVersionDispatcher{
//		        VersionDispatcher: constate.NewVersionDispatcher(mySelector),
//		    }
//		}
//
//		func (d *MyVersionDispatcher) For() (client.Object, builder.Predicates) {
//		    return &MyResource{}, builder.WithPredicates(myPredicates...)
//		}
//
//		func (d *MyVersionDispatcher) SetupWithManager(mgr ctrl.Manager, r reconcile.Reconciler, opts controller.Options) error {
//		    ...
//		}
//
//		var _ constate.StateHandler[MyResource] = (*MyVersionDispatcher)(nil)
type VersionDispatcher[T any] struct {
	selectHandler Selector[T]
}

// NewVersionDispatcher wraps a version selection function into a unified
// StateHandler.
func NewVersionDispatcher[T any](selector Selector[T]) *VersionDispatcher[T] {
	return &VersionDispatcher[T]{
		selectHandler: selector,
	}
}

// dispatch selects the version-specific handler once and forwards the call to
// it via forward. A nil selector (the constructor does not validate it) is
// reported via ErrorState with the relevant lifecycle state. Selector errors
// are likewise wrapped via ErrorState. Per the Selector contract, when the
// error is nil the selector must return a valid non-nil handler.
func (d *VersionDispatcher[T]) dispatch(
	ctx context.Context, obj *T,
	fallback ResourceState,
	forward func(StateHandler[T]) (Result, error)) (Result, error) {
	if d.selectHandler == nil {
		return ErrorState(fallback, ErrUnsetSelector)
	}
	handler, err := d.selectHandler(ctx, obj)
	if err != nil {
		return ErrorState(fallback, err)
	}
	return forward(handler)
}

func (d *VersionDispatcher[T]) HandleInitial(ctx context.Context, obj *T) (Result, error) {
	return d.dispatch(ctx, obj, StateInitial, func(h StateHandler[T]) (Result, error) {
		return h.HandleInitial(ctx, obj)
	})
}

func (d *VersionDispatcher[T]) HandleCreating(ctx context.Context, obj *T) (Result, error) {
	return d.dispatch(ctx, obj, StateCreating, func(h StateHandler[T]) (Result, error) {
		return h.HandleCreating(ctx, obj)
	})
}

func (d *VersionDispatcher[T]) HandleCreated(ctx context.Context, obj *T) (Result, error) {
	return d.dispatch(ctx, obj, StateCreated, func(h StateHandler[T]) (Result, error) {
		return h.HandleCreated(ctx, obj)
	})
}

func (d *VersionDispatcher[T]) HandleImportRequested(ctx context.Context, obj *T) (Result, error) {
	return d.dispatch(ctx, obj, StateImportRequested, func(h StateHandler[T]) (Result, error) {
		return h.HandleImportRequested(ctx, obj)
	})
}

func (d *VersionDispatcher[T]) HandleImported(ctx context.Context, obj *T) (Result, error) {
	return d.dispatch(ctx, obj, StateImported, func(h StateHandler[T]) (Result, error) {
		return h.HandleImported(ctx, obj)
	})
}

func (d *VersionDispatcher[T]) HandleUpdating(ctx context.Context, obj *T) (Result, error) {
	return d.dispatch(ctx, obj, StateUpdating, func(h StateHandler[T]) (Result, error) {
		return h.HandleUpdating(ctx, obj)
	})
}

func (d *VersionDispatcher[T]) HandleUpdated(ctx context.Context, obj *T) (Result, error) {
	return d.dispatch(ctx, obj, StateUpdated, func(h StateHandler[T]) (Result, error) {
		return h.HandleUpdated(ctx, obj)
	})
}

func (d *VersionDispatcher[T]) HandleDeletionRequested(ctx context.Context, obj *T) (Result, error) {
	return d.dispatch(ctx, obj, StateDeletionRequested, func(h StateHandler[T]) (Result, error) {
		return h.HandleDeletionRequested(ctx, obj)
	})
}

func (d *VersionDispatcher[T]) HandleDeleting(ctx context.Context, obj *T) (Result, error) {
	return d.dispatch(ctx, obj, StateDeleting, func(h StateHandler[T]) (Result, error) {
		return h.HandleDeleting(ctx, obj)
	})
}
