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

	mctrl "sigs.k8s.io/multicluster-runtime"

	kubebindv1alpha1 "github.com/ntnn/multicluster-runtime-tester/api/kube-bind/v1alpha1"
	"github.com/ntnn/multicluster-runtime-tester/pkg/reconciler"
)

// +kubebuilder:rbac:groups=kube-bind.ntnn.github.io,resources=apiserviceexportrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kube-bind.ntnn.github.io,resources=apiserviceexportrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kube-bind.ntnn.github.io,resources=apiserviceexportrequests/finalizers,verbs=update

func NewAPIServiceExportRequestReconciler() *reconciler.Reconciler[*kubebindv1alpha1.APIServiceExportRequest] {
	r := reconciler.NewReconciler[*kubebindv1alpha1.APIServiceExportRequest]("api-service-export-request")
	r.GetObject = func() *kubebindv1alpha1.APIServiceExportRequest {
		return &kubebindv1alpha1.APIServiceExportRequest{}
	}
	r.Ensure = func(ctx context.Context, rCtx reconciler.ReconcilerContext[*kubebindv1alpha1.APIServiceExportRequest]) (mctrl.Result, error) {
		return mctrl.Result{}, nil

		// log := logf.FromContext(ctx)
		// log.Info("Ensuring APIServiceExport")
		//
		// obj := &kubebindv1alpha1.APIServiceExport{
		// 	ObjectMeta: metav1.ObjectMeta{
		// 		Name: rCtx.Request.Name,
		// 		// TODO should be the provider namespace, though
		// 		// I believe the APIServiceExportRequest should be on
		// 		// the consumer side
		// 		Namespace: rCtx.Request.Namespace,
		// 	},
		// 	Spec: kubebindv1alpha1.APIServiceExportSpec{
		// 		Target: kubebindv1alpha1.ExportTarget{
		// 			Cluster:   rCtx.Request.ClusterName,
		// 			Namespace: rCtx.Request.Namespace,
		// 		},
		// 		Resources: rCtx.Object.Spec.Resources,
		// 	},
		// }
		//
		// // TODO if consumer and provider are on different clusters need
		// // a way to get the cluster from the ExportRequest and store it in the
		// // Export?
		// var result mctrl.Result
		// err := r.Upsert(ctx, rCtx.Client, obj)
		// if err != nil {
		// 	log.Error(err, "Failed to upsert APIServiceExport")
		// 	result.Requeue = true
		// }
		// return result, err
	}
	r.Delete = func(ctx context.Context, rCtx reconciler.ReconcilerContext[*kubebindv1alpha1.APIServiceExportRequest]) (mctrl.Result, error) {

		return mctrl.Result{}, nil
		// log := logf.FromContext(ctx)
		// log.Info("Deleting APIServiceExport")
		//
		// // TODO if consumer and provider are on different clusters need
		// // a way to get the cluster from the ExportRequest and store it in the
		// // Export?
		// if err := rCtx.Client.DeleteAllOf(ctx, &kubebindv1alpha1.APIServiceExport{},
		// 	client.InNamespace(rCtx.Request.Namespace),
		// 	client.MatchingFields{"metadata.name": rCtx.Request.Name}, // TODO most definitely not correct
		// ); err != nil && !apierrors.IsNotFound(err) {
		// 	log.Error(err, "Failed to delete APIServiceExport")
		// 	return mctrl.Result{Requeue: true}, err
		// }
		//
		// return mctrl.Result{}, nil
	}
	return r
}
