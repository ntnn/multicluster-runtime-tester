package reconciler

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	mctrl "sigs.k8s.io/multicluster-runtime"

	"github.com/google/uuid"
)

type ReconcilerContext[T client.Object] struct {
	Manager mctrl.Manager
	// TODO maybe rename for disambiguation as this should be the client
	// for the cluster of the resource that triggered the reconciliation
	Client  client.Client
	Request mctrl.Request
	// TODO drop, replace with GetFilledObject (e.g.)
	Object T
}

type Reconciler[T client.Object] struct {
	name string
	uid  string

	// GetObject is a function that returns an empty object of the type
	// the reconciler is responsible for.
	GetObject func() T

	// Ensure is called when the object is created or updated.
	Ensure func(context.Context, ReconcilerContext[T]) (mctrl.Result, error)

	// Delete is triggered when an object is marked for deletion.
	Delete func(context.Context, ReconcilerContext[T]) (mctrl.Result, error)

	Manager mctrl.Manager
}

func NewReconciler[T client.Object](name string) *Reconciler[T] {
	return &Reconciler[T]{
		name: name,
		uid:  uuid.New().String(),
	}
}

func (r Reconciler[T]) Name() string {
	return r.name
}

const LockedByControllerPrefix = "locked-by-controller"

func (r Reconciler[T]) LockingAnnotation() string {
	// TODO meh.
	return LockedByControllerPrefix + "-" + r.Name()
}

func (r *Reconciler[T]) LockedByMe(obj client.Object) (bool, bool) {
	add := obj.GetAnnotations()
	currentLock, ok := add[r.LockingAnnotation()]
	if !ok {
		// Object is not locked
		return false, false
	}

	if currentLock == r.uid {
		// Object is locked by current controller
		return true, true
	}

	// Object is locked by another controller
	return true, false
}

func (r *Reconciler[T]) SetupWithManager(mgr mctrl.Manager) error {
	if r.Manager != nil {
		return fmt.Errorf("manager already set")
	}

	r.Manager = mgr
	return mctrl.NewControllerManagedBy(mgr).
		For(r.GetObject()).
		WithEventFilter(
			predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					return true
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					locked, byMe := r.LockedByMe(e.Object)
					if !locked {
						return true
					}
					return byMe
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					locked, byMe := r.LockedByMe(e.ObjectNew)
					if locked {
						return byMe
					}

					newObj := e.ObjectNew.DeepCopyObject().(T)
					newAnns := newObj.GetAnnotations()
					delete(newAnns, r.LockingAnnotation())
					newObj.SetAnnotations(newAnns)
					newObj.SetManagedFields(nil)
					newObj.SetResourceVersion("")

					oldObj := e.ObjectOld.DeepCopyObject().(T)
					oldAnns := oldObj.GetAnnotations()
					delete(oldAnns, r.LockingAnnotation())
					oldObj.SetAnnotations(oldAnns)
					oldObj.SetManagedFields(nil)
					oldObj.SetResourceVersion("")

					// Ignore updates to the locking annotation
					// otherwise
					// TODO i'm sure theres a better option
					return !reflect.DeepEqual(oldObj, newObj)
				},
				GenericFunc: func(e event.GenericEvent) bool {
					locked, byMe := r.LockedByMe(e.Object)
					if !locked {
						return true
					}
					return byMe
				},
			},
		).
		Named(r.Name()).
		Complete(r)
}

func (r *Reconciler[T]) GetFilledObject(ctx context.Context, rCtx ReconcilerContext[T]) (T, error) {
	obj := r.GetObject()
	if err := rCtx.Client.Get(ctx, rCtx.Request.NamespacedName, obj); err != nil {
		return *new(T), err
	}
	return obj, nil
}

func (r *Reconciler[T]) Upsert(ctx context.Context, cl client.Client, obj client.Object) error {
	if err := cl.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		if apierrors.IsNotFound(err) {
			return cl.Create(ctx, obj)
		}
		return err
	}
	return cl.Update(ctx, obj)
}

func (r *Reconciler[T]) SetLockingAnnotaiton(ctx context.Context, rCtx ReconcilerContext[T]) (bool, error) {
	// TODO logging
	obj, err := r.GetFilledObject(ctx, rCtx)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return true, err
	}

	patch := client.MergeFrom(obj.DeepCopyObject().(T))

	ann := obj.GetAnnotations()
	ann[r.LockingAnnotation()] = r.uid
	obj.SetAnnotations(ann)

	return true, rCtx.Client.Patch(ctx, obj, patch)
}

func (r *Reconciler[T]) RemoveLockingAnnotaiton(ctx context.Context, rCtx ReconcilerContext[T]) {
	// TODO logging
	obj, err := r.GetFilledObject(ctx, rCtx)
	if err != nil {
		return
	}

	patch := client.MergeFrom(obj.DeepCopyObject().(T))

	ann := obj.GetAnnotations()
	delete(ann, r.LockingAnnotation())
	obj.SetAnnotations(ann)

	// TODO log error
	rCtx.Client.Patch(ctx, obj, patch)
}

func (r *Reconciler[T]) Reconcile(ctx context.Context, req mctrl.Request) (mctrl.Result, error) {
	log := logf.FromContext(ctx).WithValues("cluster", req.ClusterName, "namespace", req.Namespace, "name", req.Name, "reconciler", r.uid)
	log.Info("Reconciling " + r.Name())

	cluster, err := r.Manager.GetCluster(ctx, req.ClusterName)
	if err != nil {
		return mctrl.Result{}, err
	}
	cl := cluster.GetClient()

	obj := r.GetObject()
	if err := cl.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Object not found, skipping")
			return mctrl.Result{}, nil
		}
		return mctrl.Result{}, err
	}

	rCtx := ReconcilerContext[T]{
		Manager: r.Manager,
		Client:  cl,
		Request: req,
		Object:  obj,
	}

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	lockingController, ok := annotations[r.LockingAnnotation()]
	if !ok {
		log.Info("Object is not locked, setting locking annotation")
		requeue, err := r.SetLockingAnnotaiton(ctx, rCtx)
		if err != nil {
			log.Error(err, "Failed to set locking annotation")
			return mctrl.Result{Requeue: true}, err
		}
		if requeue {
			log.Info("Requeueing after setting locking annotation")
			return mctrl.Result{Requeue: true}, nil
		}
	}

	log.Info("Checking locking annotation")
	if lockingController != r.uid {
		log.Info("Object is locked by another controller, skipping")
		return mctrl.Result{}, nil
	}

	defer r.RemoveLockingAnnotaiton(ctx, rCtx)

	log.Info("Proceeding with reconciliation")

	if obj.GetDeletionTimestamp() != nil {
		log.Info("Object is being deleted, finalizing")

		var result mctrl.Result

		if r.Delete != nil {
			log.Info("Running deletion")
			var err error
			result, err = r.Delete(ctx, rCtx)
			if err != nil {
				log.Error(err, "Failed to run deletion")
				return result, err
			}
		}

		log.Info("Removing finalizer")
		patch := client.MergeFrom(obj.DeepCopyObject().(T))
		obj.SetFinalizers(
			slices.DeleteFunc(
				obj.GetFinalizers(),
				func(f string) bool {
					return f == r.Name()
				},
			),
		)
		if err := cl.Patch(ctx, obj, patch); err != nil && !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to remove finalizer")
			return mctrl.Result{Requeue: true}, err
		}
		return result, nil
	}

	if !slices.Contains(obj.GetFinalizers(), r.Name()) {
		log.Info("Adding finalizer")
		patch := client.MergeFrom(obj.DeepCopyObject().(T))
		obj.SetFinalizers(append(obj.GetFinalizers(), r.Name()))
		if err := cl.Patch(ctx, obj, patch); err != nil {
			log.Error(err, "Failed to add finalizer")
			return mctrl.Result{Requeue: true}, err
		}
		return mctrl.Result{Requeue: true}, nil
	}

	if r.Ensure != nil {
		log.Info("Ensuring object")
		result, err := r.Ensure(ctx, rCtx)
		if err != nil {
			log.Error(err, "Failed to ensure object")
		}
		return result, err
	}

	return mctrl.Result{}, nil
}
