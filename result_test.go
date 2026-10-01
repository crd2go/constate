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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var defaultRequeueResult = reconcile.Result{RequeueAfter: 15 * time.Second}

func TestNextState(t *testing.T) {
	tests := []struct {
		name        string
		state       ResourceState
		msg         string
		expected    Result
		expectedErr error
	}{
		{
			name:  "StateInitial",
			state: StateInitial,
			msg:   "",
			expected: Result{
				NextState: StateInitial,
				StateMsg:  "",
			},
		},
		{
			name:  "StateCreated",
			state: StateCreated,
			msg:   "Resource created",
			expected: Result{
				NextState: StateCreated,
				StateMsg:  "Resource created.",
			},
		},
		{
			name:  "StateCreating with requeue",
			state: StateCreating,
			msg:   "Creating resource",
			expected: Result{
				Result:    defaultRequeueResult,
				NextState: StateCreating,
				StateMsg:  "Creating resource.",
			},
		},
		{
			name:  "StateUpdated",
			state: StateUpdated,
			msg:   "Resource updated",
			expected: Result{
				NextState: StateUpdated,
				StateMsg:  "Resource updated.",
			},
		},
		{
			name:  "StateUpdating with requeue",
			state: StateUpdating,
			msg:   "Updating resource",
			expected: Result{
				Result:    defaultRequeueResult,
				NextState: StateUpdating,
				StateMsg:  "Updating resource.",
			},
		},
		{
			name:  "StateDeleted",
			state: StateDeleted,
			msg:   "Resource deleted",
			expected: Result{
				NextState: StateDeleted,
				StateMsg:  "Resource deleted.",
			},
		},
		{
			name:  "StateDeletionRequested",
			state: StateDeletionRequested,
			msg:   "Resource delete",
			expected: Result{
				Result:    defaultRequeueResult,
				NextState: StateDeletionRequested,
				StateMsg:  "Resource delete.",
			},
		},
		{
			name:  "StateDeleting",
			state: StateDeleting,
			msg:   "Deleting resource",
			expected: Result{
				Result:    defaultRequeueResult,
				NextState: StateDeleting,
				StateMsg:  "Deleting resource.",
			},
		},
		{
			name:  "StateImported",
			state: StateImported,
			msg:   "Resource imported",
			expected: Result{
				NextState: StateImported,
				StateMsg:  "Resource imported.",
			},
		},
		{
			name:  "StateImportRequested",
			state: StateImportRequested,
			msg:   "Resource import",
			expected: Result{
				NextState: StateImportRequested,
				StateMsg:  "Resource import.",
			},
		},
		{
			name:        "Unknown state",
			state:       ResourceState("Unknown"),
			msg:         "Unknown state",
			expectedErr: fmt.Errorf("unknown state %v", ResourceState("Unknown")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NextState(tt.state, tt.msg)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestErrorState(t *testing.T) {
	err := fmt.Errorf("an error occurred")
	st := StateCreating

	s, returnedErr := ErrorState(st, err)

	assert.Equal(t, Result{
		NextState: st,
	}, s)
	assert.EqualError(t, returnedErr, err.Error())
}

func TestTransitionTo(t *testing.T) {
	got := TransitionTo(StateUpdated)
	want := Result{NextState: StateUpdated}

	assert.Equal(t, want, got)
	assert.Equal(t, reconcile.Result{}, got.Result, "TransitionTo must not request a requeue")
}

func TestRequeueAfter(t *testing.T) {
	delay := 30 * time.Second
	got := RequeueAfter(StateCreating, delay)
	want := Result{
		Result:    reconcile.Result{RequeueAfter: delay},
		NextState: StateCreating,
	}

	assert.Equal(t, want, got)
	assert.Equal(t, delay, got.RequeueAfter, "only RequeueAfter should be set in the embedded reconcile.Result")
	assert.Equal(t, reconcile.Result{RequeueAfter: delay}, got.Result)
}
