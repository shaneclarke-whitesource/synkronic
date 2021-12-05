/*
Copyright 2021.

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

package controllers

import (
	"context"

	"github.com/imdario/mergo"
	appsv1 "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kyaninusv1 "codepraxis.com/kyaninus/api/v1"
)

// DeploymentVersionReconciler reconciles a DeploymentVersion object
type DeploymentVersionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kyaninus.codepraxis.com,resources=deploymentversions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kyaninus.codepraxis.com,resources=deploymentversions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kyaninus.codepraxis.com,resources=deploymentversions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DeploymentVersion object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *DeploymentVersionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var deploymentVersion kyaninusv1.DeploymentVersion
	if err := r.Get(ctx, req.NamespacedName, &deploymentVersion); err != nil {
		log.Log.Error(err, "Unable to fetch DeploymentVersion")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	baseDeploy := &appsv1.Deployment{}
	baseDeployName := types.NamespacedName{Namespace: deploymentVersion.Spec.Namespace, Name: deploymentVersion.Spec.Name}

	if err := r.Client.Get(ctx, baseDeployName, baseDeploy); err != nil {
		log.Log.Error(err, "Unable to fetch base Deployment")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	newDeploy := baseDeploy.DeepCopy()

	log.Log.Info("have DeploymentVersion!")

	if err := mergo.Merge(&newDeploy.Spec, deploymentVersion.Spec.DeploymentSpec, mergo.WithOverride); err != nil {

		log.Log.Error(err, "Error merging configuration")

		return ctrl.Result{}, client.IgnoreNotFound(err)

	}

	newDeploy.Name = newDeploy.Name + "1"
	newDeploy.ResourceVersion = ""

	if err := r.Client.Create(ctx, newDeploy); err != nil {
		log.Log.Error(err, "Error creating new deployment")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentVersionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kyaninusv1.DeploymentVersion{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
