package controllers

import (
	"context"
	"fmt"

	"github.com/imdario/mergo"
	appsv1 "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create

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
	deployVersionRef := &deploymentVersion

	if err := r.Get(ctx, req.NamespacedName, deployVersionRef); err != nil {
		log.Error(err, "Unable to fetch DeploymentVersion")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Have DeploymentVersion")

	myFinalizerName := "codepraxis.com/finalizer"

	if deploymentVersion.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(deployVersionRef, myFinalizerName) {
			controllerutil.AddFinalizer(deployVersionRef, myFinalizerName)
			if err := r.Update(ctx, deployVersionRef); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(deployVersionRef, myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, deployVersionRef); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(deployVersionRef, myFinalizerName)
			if err := r.Update(ctx, deployVersionRef); err != nil {
				return ctrl.Result{}, err
			}
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	var existingDeploy appsv1.Deployment
	err := r.Get(ctx, req.NamespacedName, &existingDeploy)

	haveDeploy := true
	if err != nil {
		log.Info("Error getting existing Deploy")
		haveDeploy = false
	}

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

		if err := ctrl.SetControllerReference(deployVersionRef, newDeploy, r.Scheme); err != nil {
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

func (r *DeploymentVersionReconciler) deleteExternalResources(ctx context.Context, deploymentVersion *kyaninusv1.DeploymentVersion) error {
	//
	// delete any external resources associated with the deploymentVersion
	//
	log := log.FromContext(ctx)

	var existingDeploy *appsv1.Deployment
	existingDeployName := types.NamespacedName{Namespace: deploymentVersion.Spec.Namespace, Name: deploymentVersion.Spec.Name}

	err := r.Get(ctx, existingDeployName, existingDeploy)
	if err != nil {
		log.Info(fmt.Sprintf("%s %s", "Removing deployment for version", existingDeployName.Name))
		delErr := r.Client.Delete(ctx, existingDeploy)
		if delErr != nil {
			log.Error(delErr, fmt.Sprintf("%s %s", "Error removing deployment: ", delErr))
		}
	}
	return nil
}
