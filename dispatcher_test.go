// Copyright 2025 MongoDB Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package constate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestVersionDispatcherLifecycleForwarding(t *testing.T) {
	ctx := context.Background()
	obj := newDummyObject(metav1.ObjectMeta{Name: "myobj"}, nil)

	tests := []struct {
		name string
		call func(*VersionDispatcher[dummyObject]) (Result, error)
		want string
	}{
		{
			name: "HandleInitial",
			call: func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleInitial(ctx, obj) },
			want: "HandleInitial",
		},
		{
			name: "HandleCreating",
			call: func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleCreating(ctx, obj) },
			want: "HandleCreating",
		},
		{
			name: "HandleCreated",
			call: func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleCreated(ctx, obj) },
			want: "HandleCreated",
		},
		{
			name: "HandleImportRequested",
			call: func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleImportRequested(ctx, obj) },
			want: "HandleImportRequested",
		},
		{
			name: "HandleImported",
			call: func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleImported(ctx, obj) },
			want: "HandleImported",
		},
		{
			name: "HandleUpdating",
			call: func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleUpdating(ctx, obj) },
			want: "HandleUpdating",
		},
		{
			name: "HandleUpdated",
			call: func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleUpdated(ctx, obj) },
			want: "HandleUpdated",
		},
		{
			name: "HandleDeletionRequested",
			call: func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleDeletionRequested(ctx, obj) },
			want: "HandleDeletionRequested",
		},
		{
			name: "HandleDeleting",
			call: func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleDeleting(ctx, obj) },
			want: "HandleDeleting",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := &recordingHandler[dummyObject]{}
			selectorCalled := false
			d := NewVersionDispatcher(func(_ context.Context, _ *dummyObject) (StateHandler[dummyObject], error) {
				selectorCalled = true
				return handler, nil
			})

			got, err := tc.call(d)
			require.NoError(t, err)
			assert.True(t, selectorCalled)
			assert.Equal(t, tc.want, handler.lastCalled)
			assert.Same(t, obj, handler.lastObj)
			assert.Equal(t, resultFor(tc.want), got)
		})
	}
}

func TestVersionDispatcherLifecycleSelectorError(t *testing.T) {
	ctx := context.Background()
	obj := newDummyObject(metav1.ObjectMeta{Name: "myobj"}, nil)

	tests := []struct {
		name          string
		call          func(*VersionDispatcher[dummyObject]) (Result, error)
		wantNextState ResourceState
	}{
		{
			name:          "HandleInitial",
			call:          func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleInitial(ctx, obj) },
			wantNextState: StateInitial,
		},
		{
			name:          "HandleCreating",
			call:          func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleCreating(ctx, obj) },
			wantNextState: StateCreating,
		},
		{
			name:          "HandleCreated",
			call:          func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleCreated(ctx, obj) },
			wantNextState: StateCreated,
		},
		{
			name:          "HandleImportRequested",
			call:          func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleImportRequested(ctx, obj) },
			wantNextState: StateImportRequested,
		},
		{
			name:          "HandleImported",
			call:          func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleImported(ctx, obj) },
			wantNextState: StateImported,
		},
		{
			name:          "HandleUpdating",
			call:          func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleUpdating(ctx, obj) },
			wantNextState: StateUpdating,
		},
		{
			name:          "HandleUpdated",
			call:          func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleUpdated(ctx, obj) },
			wantNextState: StateUpdated,
		},
		{
			name:          "HandleDeletionRequested",
			call:          func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleDeletionRequested(ctx, obj) },
			wantNextState: StateDeletionRequested,
		},
		{
			name:          "HandleDeleting",
			call:          func(h *VersionDispatcher[dummyObject]) (Result, error) { return h.HandleDeleting(ctx, obj) },
			wantNextState: StateDeleting,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewVersionDispatcher(func(_ context.Context, _ *dummyObject) (StateHandler[dummyObject], error) {
				return nil, assert.AnError
			})

			got, err := tc.call(d)
			require.ErrorIs(t, err, assert.AnError)
			assert.Equal(t, tc.wantNextState, got.NextState)
		})
	}
}

func TestVersionDispatcherNilSelectorReturnsError(t *testing.T) {
	ctx := context.Background()
	obj := newDummyObject(metav1.ObjectMeta{Name: "myobj"}, nil)

	d := NewVersionDispatcher[dummyObject](nil)

	got, err := d.HandleImportRequested(ctx, obj)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsetSelector)
	assert.Equal(t, StateImportRequested, got.NextState)
}

func TestVersionDispatcherHandlerResultForwarded(t *testing.T) {
	ctx := context.Background()
	obj := newDummyObject(metav1.ObjectMeta{Name: "myobj"}, nil)
	want := Result{Result: reconcile.Result{Requeue: true}, NextState: StateUpdating, StateMsg: "custom"}

	handler := &recordingHandler[dummyObject]{
		result: want,
	}
	d := NewVersionDispatcher(
		func(_ context.Context, _ *dummyObject) (StateHandler[dummyObject], error) { return handler, nil },
	)

	got, err := d.HandleUpdating(ctx, obj)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestVersionDispatcherEmbeddedConsumerForwardsLifecycle(t *testing.T) {
	ctx := context.Background()
	obj := newDummyObject(metav1.ObjectMeta{Name: "myobj"}, nil)

	handler := &recordingHandler[dummyObject]{}
	d := &testVersionDispatcher{
		VersionDispatcher: NewVersionDispatcher(func(_ context.Context, _ *dummyObject) (StateHandler[dummyObject], error) {
			return handler, nil
		}),
	}

	got, err := d.HandleUpdating(ctx, obj)
	require.NoError(t, err)
	assert.Equal(t, "HandleUpdating", handler.lastCalled)
	assert.Same(t, obj, handler.lastObj)
	assert.Equal(t, resultFor("HandleUpdating"), got)
}

// testVersionDispatcher is an embedded consumer of *VersionDispatcher that
// supplies the registration methods the base type intentionally omits.
type testVersionDispatcher struct {
	*VersionDispatcher[dummyObject]
}

var _ StateHandler[dummyObject] = (*testVersionDispatcher)(nil)

func (d *testVersionDispatcher) For() (client.Object, builder.Predicates) {
	return &dummyObject{}, builder.Predicates{}
}

func (d *testVersionDispatcher) SetupWithManager(ctrl.Manager, reconcile.Reconciler, controller.Options) error {
	return nil
}

// recordingHandler is a spy StateHandler that records which method was
// invoked. Only lifecycle methods called in tests are overridden; the
// embedded nil interface is never invoked for them.
type recordingHandler[T any] struct {
	StateHandler[T]

	lastCalled string
	lastObj    *T
	result     Result
}

func (r *recordingHandler[T]) HandleInitial(_ context.Context, obj *T) (Result, error) {
	return r.record("HandleInitial", obj)
}

func (r *recordingHandler[T]) HandleCreating(_ context.Context, obj *T) (Result, error) {
	return r.record("HandleCreating", obj)
}

func (r *recordingHandler[T]) HandleCreated(_ context.Context, obj *T) (Result, error) {
	return r.record("HandleCreated", obj)
}

func (r *recordingHandler[T]) HandleImportRequested(_ context.Context, obj *T) (Result, error) {
	return r.record("HandleImportRequested", obj)
}

func (r *recordingHandler[T]) HandleImported(_ context.Context, obj *T) (Result, error) {
	return r.record("HandleImported", obj)
}

func (r *recordingHandler[T]) HandleUpdating(_ context.Context, obj *T) (Result, error) {
	return r.record("HandleUpdating", obj)
}

func (r *recordingHandler[T]) HandleUpdated(_ context.Context, obj *T) (Result, error) {
	return r.record("HandleUpdated", obj)
}

func (r *recordingHandler[T]) HandleDeletionRequested(_ context.Context, obj *T) (Result, error) {
	return r.record("HandleDeletionRequested", obj)
}

func (r *recordingHandler[T]) HandleDeleting(_ context.Context, obj *T) (Result, error) {
	return r.record("HandleDeleting", obj)
}

func (r *recordingHandler[T]) record(name string, obj *T) (Result, error) {
	r.lastCalled = name
	r.lastObj = obj
	if r.result.NextState == "" && r.result.Result == (reconcile.Result{}) && r.result.StateMsg == "" {
		return resultFor(name), nil
	}
	return r.result, nil
}

func resultFor(method string) Result {
	var state ResourceState
	switch method {
	case "HandleInitial":
		state = StateInitial
	case "HandleCreating":
		state = StateCreating
	case "HandleCreated":
		state = StateCreated
	case "HandleImportRequested":
		state = StateImportRequested
	case "HandleImported":
		state = StateImported
	case "HandleUpdating":
		state = StateUpdating
	case "HandleUpdated":
		state = StateUpdated
	case "HandleDeletionRequested":
		state = StateDeletionRequested
	case "HandleDeleting":
		state = StateDeleting
	}
	return Result{NextState: state}
}
