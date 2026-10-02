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
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	DefaultRequeueTime = 15 * time.Second
)

func NextState(s ResourceState, msg string) (Result, error) {
	if len(msg) > 0 && !strings.HasSuffix(msg, ".") {
		msg = msg + "."
	}

	switch s {
	case StateCreated:
		return Result{NextState: s, StateMsg: msg}, nil

	case StateImported:
		return Result{NextState: s, StateMsg: msg}, nil

	case StateUpdated:
		return Result{NextState: s, StateMsg: msg}, nil

	case StateDeleted:
		return Result{NextState: s, StateMsg: msg}, nil

	case StateInitial:
		return Result{NextState: s, StateMsg: msg}, nil

	case StateImportRequested:
		return Result{NextState: s, StateMsg: msg}, nil

	case StateCreating:
		return Result{
			Result:    reconcile.Result{RequeueAfter: DefaultRequeueTime},
			NextState: s,
			StateMsg:  msg,
		}, nil

	case StateUpdating:
		return Result{
			Result:    reconcile.Result{RequeueAfter: DefaultRequeueTime},
			NextState: s,
			StateMsg:  msg,
		}, nil

	case StateDeleting:
		return Result{
			Result:    reconcile.Result{RequeueAfter: DefaultRequeueTime},
			NextState: s,
			StateMsg:  msg,
		}, nil

	case StateDeletionRequested:
		return Result{
			Result:    reconcile.Result{RequeueAfter: DefaultRequeueTime},
			NextState: s,
			StateMsg:  msg,
		}, nil

	default:
		return Result{}, fmt.Errorf("unknown state %v", s)
	}
}

// TransitionTo returns a Result that transitions the resource to the given
// state without requesting a requeue. It is equivalent to the literal
// Result{NextState: state}.
func TransitionTo(state ResourceState) Result {
	return Result{NextState: state}
}

// RequeueAfter returns a Result that transitions the resource to the given
// state and requests a requeue after the given delay. It is equivalent to
// the literal Result{Result: reconcile.Result{RequeueAfter: delay}, NextState: state}.
func RequeueAfter(state ResourceState, delay time.Duration) Result {
	return Result{
		Result:    reconcile.Result{RequeueAfter: delay},
		NextState: state,
	}
}

func ErrorState(s ResourceState, err error) (Result, error) {
	return Result{
		NextState: s,
	}, err
}
