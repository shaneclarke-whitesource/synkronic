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

var (
	jobOwnerKey = ".metadata.controller"
	//apiGVStr    = kyaninusv1.GroupVersion.String()
)

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
	log := log.FromContext(ctx)

	var deploymentVersion kyaninusv1.DeploymentVersion
	if err := r.Get(ctx, req.NamespacedName, &deploymentVersion); err != nil {
		log.Error(err, "Unable to fetch DeploymentVersion")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var existingDeploy appsv1.Deployment
	err := r.Get(ctx, req.NamespacedName, &existingDeploy)

	haveDeploy := true
	if err != nil {
		log.Info("Error getting existing Deploy")
		haveDeploy = false
	}

	log.Info("Have DeploymentVersion")

	baseDeploy := &appsv1.Deployment{}
	baseDeployName := types.NamespacedName{Namespace: deploymentVersion.Spec.Namespace, Name: deploymentVersion.Spec.Name}

	if err := r.Client.Get(ctx, baseDeployName, baseDeploy); err != nil {
		log.Error(err, "Unable to fetch base Deployment")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	newDeploy := baseDeploy.DeepCopy()

	if err := mergo.Merge(&newDeploy.Spec, deploymentVersion.Spec.DeploymentSpec, mergo.WithOverride); err != nil {
		log.Error(err, "Error merging configuration")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	newDeploy.Name = deploymentVersion.Name
	newDeploy.Namespace = deploymentVersion.Namespace
	newDeploy.ResourceVersion = ""

	if haveDeploy {
		if err := r.Client.Update(ctx, newDeploy); err != nil {
			log.Error(err, "Error updating existing deployment")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	} else {
		if err := r.Client.Create(ctx, newDeploy); err != nil {
			log.Error(err, "Error creating new deployment")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		if err := ctrl.SetControllerReference(&deploymentVersion, newDeploy, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
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
