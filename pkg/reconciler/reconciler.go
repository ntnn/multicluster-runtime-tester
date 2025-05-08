package reconciler

import (
	"context"
	"fmt"
	"slices"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	mctrl "sigs.k8s.io/multicluster-runtime"

	"github.com/google/uuid"
)

type ReconcilerContext[T client.Object] struct {
	Manager mctrl.Manager
	Client  client.Client
	Request mctrl.Request
	Object  T
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

func (r *Reconciler[T]) SetupWithManager(mgr mctrl.Manager) error {
	if r.Manager != nil {
		return fmt.Errorf("manager already set")
	}

	r.Manager = mgr
	return mctrl.NewControllerManagedBy(mgr).
		For(r.GetObject()).
		Named(r.Name()).
		Complete(r)
}

func (r *Reconciler[T]) Upsert(ctx context.Context, cl client.Client, obj client.Object) error {
	// TODO set owner ?
	if err := cl.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		if apierrors.IsNotFound(err) {
			return cl.Create(ctx, obj)
		}
		return err
	}
	return cl.Update(ctx, obj)
}

func (r *Reconciler[T]) Reconcile(ctx context.Context, req mctrl.Request) (mctrl.Result, error) {
	log := logf.FromContext(ctx).WithValues("cluster", req.ClusterName, "namespace", req.Namespace, "name", req.Name)
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

	// TODO lock object

	if lockingController, ok := obj.GetAnnotations()[r.LockingAnnotation()]; ok && lockingController != r.Name() {
		// Object is locked by another controller,
		return mctrl.Result{}, nil
	}

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[r.LockingAnnotation()] = r.Name()
	obj.SetAnnotations(annotations)
	log.Info("Setting locking annotation")
	if err := r.Upsert(ctx, cl, obj); err != nil {
		log.Error(err, "Failed to set locking annotation")
		return mctrl.Result{Requeue: true}, err
	}

	defer func() {
		// TODO this could turn into a dead lock if removing the locking
		// annotation fails
		annotations := obj.GetAnnotations()
		delete(annotations, r.LockingAnnotation())
		obj.SetAnnotations(annotations)
		log.Info("Removing locking annotation")
		// Ignore IsNotFound error, as the object may have been deleted
		if err := r.Upsert(ctx, cl, obj); err != nil && !apierrors.IsNotFound(err) {
			// TODO this errors with "resourceVersion should not be set on
			// objects to be created"
			log.Error(err, "Failed to remove locking annotation")
		}
	}()

	if obj.GetDeletionTimestamp() != nil {
		log.Info("Object is being deleted, finalizing")

		if r.Delete != nil {
			log.Info("Deleting object")
			result, err := r.Delete(ctx, rCtx)
			if err != nil {
				log.Error(err, "Failed to run deletion")
			}
			return result, err
		}

		log.Info("Removing finalizer")
		obj.SetFinalizers(
			slices.DeleteFunc(
				obj.GetFinalizers(),
				func(f string) bool {
					return f == r.Name()
				},
			),
		)
		if err := cl.Update(ctx, obj); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return mctrl.Result{Requeue: true}, err
		}

		return mctrl.Result{}, nil
	}

	if !slices.Contains(obj.GetFinalizers(), r.Name()) {
		log.Info("Adding finalizer")
		obj.SetFinalizers(append(obj.GetFinalizers(), r.Name()))
		if err := cl.Update(ctx, obj); err != nil {
			log.Error(err, "Failed to add finalizer")
			return mctrl.Result{Requeue: true}, err
		}
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
