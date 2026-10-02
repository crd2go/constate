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
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

func TestFallbackHandlerNotImplemented(t *testing.T) {
	tests := []struct {
		name     string
		call     func(h *FallbackHandler[dummyObject], ctx context.Context) (Result, error)
		wantErr  string
		wantNext ResourceState
	}{
		{
			name: "HandleInitial",
			call: func(h *FallbackHandler[dummyObject], ctx context.Context) (Result, error) {
				return h.HandleInitial(ctx, &dummyObject{})
			},
			wantErr:  "HandleInitial not implemented",
			wantNext: StateInitial,
		},
		{
			name: "HandleImportRequested",
			call: func(h *FallbackHandler[dummyObject], ctx context.Context) (Result, error) {
				return h.HandleImportRequested(ctx, &dummyObject{})
			},
			wantErr:  "HandleImportRequested not implemented",
			wantNext: StateImportRequested,
		},
		{
			name: "HandleImported",
			call: func(h *FallbackHandler[dummyObject], ctx context.Context) (Result, error) {
				return h.HandleImported(ctx, &dummyObject{})
			},
			wantErr:  "HandleImported not implemented",
			wantNext: StateImported,
		},
		{
			name: "HandleCreating",
			call: func(h *FallbackHandler[dummyObject], ctx context.Context) (Result, error) {
				return h.HandleCreating(ctx, &dummyObject{})
			},
			wantErr:  "HandleCreating not implemented",
			wantNext: StateCreating,
		},
		{
			name: "HandleCreated",
			call: func(h *FallbackHandler[dummyObject], ctx context.Context) (Result, error) {
				return h.HandleCreated(ctx, &dummyObject{})
			},
			wantErr:  "HandleCreated not implemented",
			wantNext: StateCreated,
		},
		{
			name: "HandleUpdating",
			call: func(h *FallbackHandler[dummyObject], ctx context.Context) (Result, error) {
				return h.HandleUpdating(ctx, &dummyObject{})
			},
			wantErr:  "HandleUpdating not implemented",
			wantNext: StateUpdating,
		},
		{
			name: "HandleUpdated",
			call: func(h *FallbackHandler[dummyObject], ctx context.Context) (Result, error) {
				return h.HandleUpdated(ctx, &dummyObject{})
			},
			wantErr:  "HandleUpdated not implemented",
			wantNext: StateUpdated,
		},
		{
			name: "HandleDeletionRequested",
			call: func(h *FallbackHandler[dummyObject], ctx context.Context) (Result, error) {
				return h.HandleDeletionRequested(ctx, &dummyObject{})
			},
			wantErr:  "HandleDeletionRequested not implemented",
			wantNext: StateDeletionRequested,
		},
		{
			name: "HandleDeleting",
			call: func(h *FallbackHandler[dummyObject], ctx context.Context) (Result, error) {
				return h.HandleDeleting(ctx, &dummyObject{})
			},
			wantErr:  "HandleDeleting not implemented",
			wantNext: StateDeleting,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &FallbackHandler[dummyObject]{}
			result, err := tc.call(h, context.Background())
			require.ErrorContains(t, err, tc.wantErr)
			assert.Equal(t, tc.wantNext, result.NextState)
		})
	}
}

func TestFallbackHandlerSetupWithManager(t *testing.T) {
	h := &FallbackHandler[dummyObject]{}
	err := h.SetupWithManager(nil, nil, controller.Options{})
	require.ErrorContains(t, err, "SetupWithManager not implemented")
}

func TestFallbackHandlerForcedFor(t *testing.T) {
	h := &testFallbackHandler{}
	obj, predicates := h.For()
	assert.Equal(t, &dummyObject{}, obj)
	assert.Equal(t, builder.Predicates{}, predicates)
}

// testFallbackHandler is an embedded consumer of FallbackHandler that must
// supply For itself; the fallback base intentionally omits it, which is
// enforced at compile time by the StateHandler assertion below.
type testFallbackHandler struct {
	FallbackHandler[dummyObject]
}

var _ StateHandler[dummyObject] = (*testFallbackHandler)(nil)

func (h *testFallbackHandler) For() (client.Object, builder.Predicates) {
	return &dummyObject{}, builder.Predicates{}
}
