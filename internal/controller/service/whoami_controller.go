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

package service

import (
	"context"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mctrl "sigs.k8s.io/multicluster-runtime"

	servicev1alpha1 "github.com/ntnn/multicluster-runtime-tester/api/service/v1alpha1"
)

// WhoamiReconciler reconciles a Whoami object
type WhoamiReconciler struct {
}

// +kubebuilder:rbac:groups=service.ntnn.github.io,resources=whoamis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=service.ntnn.github.io,resources=whoamis/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=service.ntnn.github.io,resources=whoamis/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Whoami object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *WhoamiReconciler) Reconcile(ctx context.Context, req mctrl.Request) (mctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// TODO(user): your logic here

	return mctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WhoamiReconciler) SetupWithManager(mgr mctrl.Manager) error {
	return mctrl.NewControllerManagedBy(mgr).
		For(&servicev1alpha1.Whoami{}).
		Named("whoami").
		Complete(r)
}
