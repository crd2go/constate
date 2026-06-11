# constate

**constate** (CONtroller STATE) is a universal state machine library for Kubernetes operators. It replaces the open-ended `Reconcile` loop from controller-runtime with a structured finite state machine (FSM): instead of writing one monolithic reconciler, you implement focused handlers — one per state — and constate routes each reconciliation to the right handler automatically.

## Why

controller-runtime gives you a blank `Reconcile(ctx, req)` function. In practice, every operator ends up re-implementing the same patterns: detect whether a resource exists upstream, track whether it's being created or updated, manage finalizers during deletion. This leads to deeply nested conditionals and logic that is hard to follow and test in isolation.

constate models each managed resource as a CRUD state machine. Your code becomes a set of small, single-purpose functions, each responsible for one phase of the resource lifecycle.

## State machine

```
Initial ──► Creating ──► Created ◄──► Updating ──► Updated
   │                                                   │
   └──► ImportRequested ──► Imported ◄─────────────────┘
                                                        │
any ──► DeletionRequested ──► Deleting ──► Deleted ◄───┘
```

States are stored as `State` and `Ready` status conditions on the resource, so they are visible via `kubectl get` and usable in `wait` expressions.

| State               | Ready | Meaning                                    |
|---------------------|-------|--------------------------------------------|
| `Initial`           | False | Resource just appeared, never reconciled   |
| `Creating`          | False | Upstream resource creation in progress     |
| `Created`           | True  | Resource exists and is settled             |
| `Updating`          | False | Upstream resource update in progress       |
| `Updated`           | True  | Resource updated and settled               |
| `ImportRequested`   | False | External-id annotation detected, importing |
| `Imported`          | True  | Upstream resource adopted                  |
| `DeletionRequested` | False | `deletionTimestamp` set, clean-up starting |
| `Deleting`          | False | Upstream deletion in progress              |
| `Deleted`           | —     | Terminal — finalizers removed, object gone |

## Usage

Implement the `StateHandler[T]` interface for your CRD type. Each method receives the fully fetched object and returns a `Result` carrying the next state.

```go
type Result struct {
    reconcile.Result          // standard requeue/requeueAfter
    NextState state.ResourceState
    StateMsg  string
}
```

```go
type MyResourceHandler struct {
    client client.Client
}

func (h *MyResourceHandler) HandleInitial(ctx context.Context, obj *MyResource) (state.Result, error) {
    if err := h.client.CreateUpstream(ctx, obj); err != nil {
        return state.Result{}, err
    }
    return state.Result{
        NextState: state.StateCreating,
        Result:    reconcile.Result{RequeueAfter: 5 * time.Second},
    }, nil
}

func (h *MyResourceHandler) HandleCreating(ctx context.Context, obj *MyResource) (state.Result, error) {
    ready, err := h.client.IsReady(ctx, obj)
    if err != nil {
        return state.Result{}, err
    }
    if !ready {
        return state.Result{
            NextState: state.StateCreating,
            Result:    reconcile.Result{RequeueAfter: 5 * time.Second},
        }, nil
    }
    return state.Result{NextState: state.StateCreated}, nil
}

// ... implement remaining handlers
```

Wire it into controller-runtime:

```go
reconciler := state.NewStateReconciler(&MyResourceHandler{client: mgr.GetClient()})
if err := reconciler.SetupWithManager(mgr, controller.Options{}); err != nil {
    return err
}
```

## Importing existing resources

If a resource arrives with a `mongodb.com/external-<key>` annotation, constate automatically transitions it to `ImportRequested` instead of `Initial`. Implement `HandleImportRequested` to adopt the upstream resource and transition to `Imported`.

## Drift detection

`ComputeStateTracker` produces a stable hash of the resource's generation plus any referenced Secrets and ConfigMaps. Store it with `Patcher.UpdateStateTracker(deps...)` when the resource is settled and compare it on the next reconcile to detect drift without re-calling the upstream API on every loop.

## Patching

`Patcher` wraps Kubernetes server-side apply for status and annotations:

```go
NewPatcher(obj).
    UpdateStatus().
    UpdateStateTracker(secret, configMap).
    Patch(ctx, client)
```

## Skipping reconciliation

Annotate a resource with `mongodb.com/atlas-reconciliation-policy=skip` to pause reconciliation without deleting it.

## Install

```sh
go get github.com/crd2go/constate
```
