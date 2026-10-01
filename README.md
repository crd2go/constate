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

Thanks to constate you write a handler but you still need to register a controller-runtime reconciler at setup time.

The way you do that is first you implement the handler (more on that later) and you wrap it within a reconciler that can be registered in the controller manager as usual:

```go
func setup(mgr ctrl.Manager) error {
  ...
  h := NewMyResourceHandler(...)
  reconciler := constate.NewStateReconciler(h)
  if err := reconciler.SetupWithManager(mgr, controller.Options{}); err != nil {
    return err
  }
  ...
}
```

### Simple handler

A handler implements the `constate.StateHandler[T]` interface of a `T` resource types.

Implement all `Handle...` lifecycle methods plus `SetupWithManager` and `For`. To implement only the relevant handle methods to your resource, you can embed `constate.FallbackHandler[T]` within your struct, which will fail loudly and explciciltly if any of the non implemented methods is used, except for `For`, which still must be implemented explicitly, flagged at compile time.

Example:
```go
...
  // MyResourceHandler implements constate.StateHandler
  type MyResourceHandler struct {
    constate.FallbackHandler[MyResource] // allows to only implement used handle methods
    apiClient APIClient // upstream API client injected by operator
    predicates builder.Predicates
    ...
  }

  func (h *MyResourceHandler) For() (client.Object, builder.Predicates) {
    return &MyResource{}, builder.WithPredicates(h.predicates...)
  }

  func (h *MyResourceHandler) SetupWithManager(mgr ctrl.Manager, r reconcile.Reconciler, opts controller.Options) error {
    ...
  }

  func (h *MyResourceHandler) HandleInitial(ctx context.Context, obj *MyResource) (constate.Result, error) {
    if err := h.apiClient.Create(ctx, obj); err != nil {
        return constate.Result{}, err
    }
    return constate.RequeueAfter(constate.StateCreating, 5*time.Second), nil
  }

  func (h *MyResourceHandler) HandleCreating(ctx context.Context, obj *MyResource) (constate.Result, error) {
    ready, err := h.apiClient.IsReady(ctx, obj)
    if err != nil {
        return constate.Result{}, err
    }
    if !ready {
        return constate.RequeueAfter(constate.StateCreating, 5*time.Second), nil
    }
    return constate.TransitionTo(constate.StateCreated), nil
  }
  // implementation of all remaining relevant interface methods follows
```

#### Raw API escape hatch

`constate.TransitionTo` and `constate.RequeueAfter` are convenience constructors for the plain `constate.Result` struct. You can always construct a `Result` literal directly, and `NewStateReconciler` works either way:

```go
func (h *MyResourceHandler) HandleInitial(ctx context.Context, obj *MyResource) (constate.Result, error) {
    if err := h.apiClient.Create(ctx, obj); err != nil {
        return constate.Result{}, err
    }
    return constate.Result{
        NextState: constate.StateCreating,
        Result:    reconcile.Result{RequeueAfter: 5 * time.Second},
    }, nil
}
```

### Multi-version handling

When resources are managed by APIs with support for multiple versions `constate` provides a `VersionDispatcher` embeddable helper:
- Embed a pointer to `VersionDispatcher[T]` within your top-level handler.
- Implement a `Selector func(ctx context.Context, obj *T) (StateHandler[T], error)`.
- Ensure your top level handler initializes its embedded dispatcher with `constate.NewVersionDispatcher(selector)`.
- Implement `SetupWithManager`, `For` and relevant `Handle...` methods, as usual. The dispatcher does not cover `For` or `SetupWithManager`.
- Provide a `constate.Selector[T]` function that decides which handler must handle the reconciliation given the resource version.
- Most likely need to keep track of the various version handlers within your top-level handler.

Note that `Selector[T]` returns a `StateHandler[T]`, so each selected handler must satisfy the full `constate.StateHandler[T]` interface, including `For` and `SetupWithManager`. Those registration methods are simply not used by the top-level controller setup, but need to be implemented either explicitly. Remenber `FallbackHandler[T]` will not cover `For`.

Example:

```go
...
  // MyMultiVersionResourceHandler implements constate.StateHandler
  type MyMultiVersionResourceHandler struct {
    *constate.VersionDispatcher[MyResource]
    apiClient APIClient // upstream API client injected by operator
    predicates builder.Predicates
    handlers map[string]constate.StateHandler[MyResource] // per version handlers
    ...
  }

  func NewMyMultiVersionResourceHandler(...) *MyMultiVersionResourceHandler {
    h := &MyMultiVersionResourceHandler{...}
    // The dispatcher must be initialized with a selector
    h.VersionDispatcher = constate.NewVersionDispatcher(h.Selector)
    return h
  }

  func (h *MyMultiVersionResourceHandler) Selector(ctx context.Context, obj *MyResource) (StateHandler[MyResource], error) {
    ...
  }

  func (h *MyMultiVersionResourceHandler) For() (client.Object, builder.Predicates) { ... }

  func (h *MyMultiVersionResourceHandler) SetupWithManager(mgr ctrl.Manager, r reconcile.Reconciler, opts controller.Options) error {
    ...
  }
  // No need to implement Handle... methods as those are promoted from the
  // embedded dispatcher, which calls the appropriate handler for each state.
  // The top-level `For` and `SetupWithManager` are supplied by the outer handler.
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
