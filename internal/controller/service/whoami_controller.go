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
	"errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mctrl "sigs.k8s.io/multicluster-runtime"

	servicev1alpha1 "github.com/ntnn/multicluster-runtime-tester/api/service/v1alpha1"
)

// WhoamiReconciler reconciles a Whoami object
type WhoamiReconciler struct {
	Manager mctrl.Manager
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
	log := logf.FromContext(ctx).WithValues("cluster", req.ClusterName, "namespace", req.Namespace, "name", req.Name)
	log.Info("Reconciling Whoami")

	cluster, err := r.Manager.GetCluster(ctx, req.ClusterName)
	if err != nil {
		return mctrl.Result{}, err
	}

	cl := cluster.GetClient()

	whoami := &servicev1alpha1.Whoami{}
	if err := cl.Get(ctx, req.NamespacedName, whoami); err != nil {
		if apierrors.IsNotFound(err) {
			return mctrl.Result{}, DeleteWhoami(ctx, cl, req)
		}
		return mctrl.Result{}, err
	}

	if err := EnsureWhoami(ctx, cl, req); err != nil {
		log.Error(err, "Failed to ensure Whoami resources")
		return mctrl.Result{}, err
	}

	return mctrl.Result{}, nil
}

var usedResourceTypes = []client.Object{
	&appsv1.Deployment{},
	&corev1.Service{},
}

func GenWhoamiResources(req mctrl.Request) []client.Object {
	meta := metav1.ObjectMeta{
		Name:      req.Name,
		Namespace: req.Namespace,
		Labels: map[string]string{
			"whoami": req.Name,
		},
	}

	return []client.Object{
		&appsv1.Deployment{
			ObjectMeta: meta,
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"whoami": req.Name,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: req.Namespace,
						Labels: map[string]string{
							"whoami": req.Name,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "whoami",
								Image: "traefik/whoami",
							},
						},
					},
				},
			},
		},
		&corev1.Service{
			ObjectMeta: meta,
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"whoami": req.Name,
				},
				Ports: []corev1.ServicePort{
					{
						Port:       80,
						TargetPort: intstr.FromInt(80),
					},
				},
			},
		},
	}
}

func EnsureWhoami(ctx context.Context, cl client.Client, req mctrl.Request) error {
	log := logf.FromContext(ctx)
	log.Info("Creating Whoami")

	var errs error
	for _, resource := range GenWhoamiResources(req) {
		log.Info("Creating Whoami resource", "resource", resource)

		if err := cl.Get(ctx, client.ObjectKeyFromObject(resource), resource); err != nil {
			// Resource doesn't exist, create it
			if err := cl.Create(ctx, resource); err != nil {
				log.Error(err, "Failed to create Whoami resource", "resource", resource)
				err = errors.Join(errs, err)
			}
		} else {
			// Resource exists, update it
			if err := cl.Update(ctx, resource); err != nil {
				log.Error(err, "Failed to update Whoami resource", "resource", resource)
				err = errors.Join(errs, err)
			}
		}
	}
	return errs
}

func DeleteWhoami(ctx context.Context, cl client.Client, req mctrl.Request) error {
	log := logf.FromContext(ctx)
	log.Info("Deleting Whoami")

	var errs error
	for _, resource := range usedResourceTypes {
		if err := cl.DeleteAllOf(ctx, resource,
			client.InNamespace(req.Namespace),
			client.MatchingLabels{"whoami": req.Name},
		); err != nil {
			log.Error(err, "Failed to delete Whoami resources", "resource", resource)
			err = errors.Join(errs, err)
		}
	}
	return errs
}

// SetupWithManager sets up the controller with the Manager.
func (r *WhoamiReconciler) SetupWithManager(mgr mctrl.Manager) error {
	return mctrl.NewControllerManagedBy(mgr).
		For(&servicev1alpha1.Whoami{}).
		Named("whoami").
		Complete(r)
}
