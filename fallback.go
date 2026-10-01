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
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// FallbackHandler is an embeddable base that provides fail-loud
// implementations for all StateHandler methods except For,
// since it has no error return. Omitting it forces embedding types to
// implement For() themselves, which is checked at compile time.
//
// Embed it in your handler type when you only want to implement a subset of
// the lifecycle: any call to an unimplemented method returns an ErrorState
// error instead of panicking or silently doing nothing.
//
// Example:
//
//	type MyHandler struct {
//		constate.FallbackHandler[MyResource]
//	}
//
//	func (h *MyHandler) For() (client.Object, builder.Predicates) { ... }
//	func (h *MyHandler) HandleInitial(ctx context.Context, obj *MyResource) (constate.Result, error) { ... }
type FallbackHandler[T any] struct{}

func (h *FallbackHandler[T]) HandleInitial(ctx context.Context, obj *T) (Result, error) {
	return ErrorState(StateInitial, fmt.Errorf("HandleInitial not implemented"))
}

func (h *FallbackHandler[T]) HandleImportRequested(ctx context.Context, obj *T) (Result, error) {
	return ErrorState(StateImportRequested, fmt.Errorf("HandleImportRequested not implemented"))
}

func (h *FallbackHandler[T]) HandleImported(ctx context.Context, obj *T) (Result, error) {
	return ErrorState(StateImported, fmt.Errorf("HandleImported not implemented"))
}

func (h *FallbackHandler[T]) HandleCreating(ctx context.Context, obj *T) (Result, error) {
	return ErrorState(StateCreating, fmt.Errorf("HandleCreating not implemented"))
}

func (h *FallbackHandler[T]) HandleCreated(ctx context.Context, obj *T) (Result, error) {
	return ErrorState(StateCreated, fmt.Errorf("HandleCreated not implemented"))
}

func (h *FallbackHandler[T]) HandleUpdating(ctx context.Context, obj *T) (Result, error) {
	return ErrorState(StateUpdating, fmt.Errorf("HandleUpdating not implemented"))
}

func (h *FallbackHandler[T]) HandleUpdated(ctx context.Context, obj *T) (Result, error) {
	return ErrorState(StateUpdated, fmt.Errorf("HandleUpdated not implemented"))
}

func (h *FallbackHandler[T]) HandleDeletionRequested(ctx context.Context, obj *T) (Result, error) {
	return ErrorState(StateDeletionRequested, fmt.Errorf("HandleDeletionRequested not implemented"))
}

func (h *FallbackHandler[T]) HandleDeleting(ctx context.Context, obj *T) (Result, error) {
	return ErrorState(StateDeleting, fmt.Errorf("HandleDeleting not implemented"))
}

func (h *FallbackHandler[T]) SetupWithManager(mgr ctrl.Manager, r reconcile.Reconciler, opts controller.Options) error {
	return fmt.Errorf("SetupWithManager not implemented")
}
