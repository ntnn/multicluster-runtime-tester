/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubebind

import (
	"context"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	mctrl "sigs.k8s.io/multicluster-runtime"

	kubebindv1alpha1 "github.com/ntnn/multicluster-runtime-tester/api/kube-bind/v1alpha1"
	"github.com/ntnn/multicluster-runtime-tester/pkg/reconciler"
)

// APIServiceExportReconciler reconciles a APIServiceExport object
//
// The APIServiceExport is in the provider cluster and manages:
// - APIServiceBinding in the consumer cluster
// - CRD in the consumer cluster
// - BoundAPIResourceSchema in the provider cluster
//
// +kubebuilder:rbac:groups=kube-bind.ntnn.github.io,resources=apiserviceexports,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kube-bind.ntnn.github.io,resources=apiserviceexports/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kube-bind.ntnn.github.io,resources=apiserviceexports/finalizers,verbs=update

func NewAPIServiceExportReconciler() *reconciler.Reconciler[*kubebindv1alpha1.APIServiceExport] {
	r := reconciler.NewReconciler[*kubebindv1alpha1.APIServiceExport]("api-service-export")
	r.GetObject = func() *kubebindv1alpha1.APIServiceExport {
		return &kubebindv1alpha1.APIServiceExport{}
	}
	r.Ensure = func(ctx context.Context, rCtx reconciler.ReconcilerContext[*kubebindv1alpha1.APIServiceExport]) (mctrl.Result, error) {
		log := logf.FromContext(ctx)
		log.Info("Ensuring ")

		consumerCluster, err := rCtx.Manager.GetCluster(ctx, rCtx.Object.Spec.Target.Cluster)
		if err != nil {
			log.Error(err, "Failed to get consumer cluster")
			return mctrl.Result{Requeue: true}, err
		}
		consumerClient := consumerCluster.GetClient()

		var result mctrl.Result
		var errs error

		// TODO APIServiceBinding
		apiServiceBinding := &kubebindv1alpha1.APIServiceBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rCtx.Object.Name,
				Namespace: rCtx.Object.Spec.Target.Namespace,
			},
			Spec:   kubebindv1alpha1.APIServiceBindingSpec{},
			Status: kubebindv1alpha1.APIServiceBindingStatus{
				// TODO
			},
		}
		if err := r.Upsert(ctx, consumerClient, apiServiceBinding); err != nil {
			log.Error(err, "Failed to upsert APIServiceBinding")
			result.Requeue = true
			errs = errors.Join(errs, err)
		}

		// TODO CRD

		// providerCluster, err := rCtx.Manager.GetCluster(ctx, rCtx.Request.ClusterName)
		// if err != nil {
		// 	log.Error(err, "Failed to get provider cluster")
		// 	return err
		// }
		// providerClient := providerCluster.GetClient()

		// If object creation in the consumer cluster failed skip
		// creating the BoundAPIResourceSchema in the provider.
		if errs != nil {
			return result, errs
		}

		result = mctrl.Result{}

		// TODO BoundAPIResourceSchema

		return result, errs
	}
	r.Delete = func(ctx context.Context, rCtx reconciler.ReconcilerContext[*kubebindv1alpha1.APIServiceExport]) (mctrl.Result, error) {
		log := logf.FromContext(ctx)
		log.Info("Deleting ")

		providerCluster, err := rCtx.Manager.GetCluster(ctx, rCtx.Request.ClusterName)
		if err != nil {
			log.Error(err, "Failed to get provider cluster")
			return mctrl.Result{Requeue: true}, err
		}
		providerClient := providerCluster.GetClient()

		consumerCluster, err := rCtx.Manager.GetCluster(ctx, rCtx.Object.Spec.Target.Cluster)
		if err != nil {
			log.Error(err, "Failed to get consumer cluster")
			return mctrl.Result{Requeue: true}, err
		}
		consumerClient := consumerCluster.GetClient()

		var result mctrl.Result
		var errs error

		if err := consumerClient.DeleteAllOf(ctx, &kubebindv1alpha1.APIServiceBinding{},
			// TODO fix selector
			client.InNamespace(rCtx.Object.Spec.Target.Namespace),
			client.MatchingFields{"metadata.name": rCtx.Object.Name},
		); err != nil {
			log.Error(err, "Failed to delete APIServiceBinding")
			result.Requeue = true
			errs = errors.Join(errs, err)
		}

		// TODO delte CRD

		if err := providerClient.DeleteAllOf(ctx, &kubebindv1alpha1.BoundAPIResourceSchema{},
			// TODO fix selector
			client.InNamespace(rCtx.Object.Spec.Target.Namespace),
			client.MatchingFields{"metadata.name": rCtx.Object.Name},
		); err != nil {
			log.Error(err, "Failed to delete BoundAPIResourceSchema")
			result.Requeue = true
			errs = errors.Join(errs, err)
		}

		return result, errs
	}
	return r
}
