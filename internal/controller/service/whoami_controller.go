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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mctrl "sigs.k8s.io/multicluster-runtime"

	servicev1alpha1 "github.com/ntnn/multicluster-runtime-tester/api/service/v1alpha1"

	"github.com/ntnn/multicluster-runtime-tester/pkg/reconciler"
)

// +kubebuilder:rbac:groups=service.ntnn.github.io,resources=whoamis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=service.ntnn.github.io,resources=whoamis/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=service.ntnn.github.io,resources=whoamis/finalizers,verbs=update

// WhoamiReconciler reconciles a Whoami object
func NewWhoamiReconciler() *reconciler.Reconciler[*servicev1alpha1.Whoami] {
	r := reconciler.NewReconciler[*servicev1alpha1.Whoami]("whoami")
	r.GetObject = func() *servicev1alpha1.Whoami {
		return &servicev1alpha1.Whoami{}
	}
	r.Ensure = func(ctx context.Context, rCtx reconciler.ReconcilerContext[*servicev1alpha1.Whoami]) error {
		log := logf.FromContext(ctx)
		log.Info("Creating Whoami")
		var errs error
		for _, obj := range GenWhoamiResources(rCtx.Request) {
			if err := r.Upsert(ctx, rCtx.Client, obj); err != nil {
				log.Error(err, "Failed to upsert Whoami resource", "resource", obj)
				errs = errors.Join(errs, err)
			}
		}
		return errs
	}
	r.Delete = func(ctx context.Context, rCtx reconciler.ReconcilerContext[*servicev1alpha1.Whoami]) error {
		log := logf.FromContext(ctx)
		log.Info("Deleting Whoami")

		var errs error
		for _, resource := range usedResourceTypes {
			if err := rCtx.Client.DeleteAllOf(ctx, resource,
				client.InNamespace(rCtx.Request.Namespace),
				client.MatchingLabels{"whoami": rCtx.Request.Name},
			); err != nil {
				log.Error(err, "Failed to delete Whoami resources", "resource", resource)
				err = errors.Join(errs, err)
			}
		}
		return errs

	}
	return r
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
