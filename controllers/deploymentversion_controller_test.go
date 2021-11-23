/*
Ideally, we should have one `<kind>_controller_test.go` for each controller scaffolded and called in the `suite_test.go`.
So, let's write our example test for the CronJob controller (`cronjob_controller_test.go.`)
*/

/*

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
// +kubebuilder:docs-gen:collapse=Apache License

/*
As usual, we start with the necessary imports. We also define some utility variables.
*/
package controllers

import (
	"context"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kyaninusv1 "codepraxis.com/kyaninus/api/v1"
	appsv1 "k8s.io/api/apps/v1"
)

// +kubebuilder:docs-gen:collapse=Imports

/*
The first step to writing a simple integration test is to actually create an instance of CronJob you can run tests against.
Note that to create a CronJob, you’ll need to create a stub CronJob struct that contains your CronJob’s specifications.

Note that when we create a stub CronJob, the CronJob also needs stubs of its required downstream objects.
Without the stubbed Job template spec and the Pod template spec below, the Kubernetes API will not be able to
create the CronJob.
*/
var _ = Describe("DeploymentVersion controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		DeployName      = "myappdeploy"
		DeployNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When reconciling DeploymentVersion", func() {
		It("Something should happen", func() {

			By("By creating an initial User Deployment")
			ctx := context.Background()

			matchLabels := map[string]string{"app": "version1"}
			labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}

			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: DeployName, Namespace: DeployNamespace},
				Spec: appsv1.DeploymentSpec{
					Selector: &labelSelector,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Name: DeployName, Namespace: DeployNamespace, Labels: matchLabels},
						Spec: v1.PodSpec{
							// For simplicity, we only fill out the required fields.
							Containers: []v1.Container{
								{
									Name:  "test-container",
									Image: "test-image",
								},
							},
							RestartPolicy: v1.RestartPolicyAlways,
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, deployment)).Should(Succeed())

			By("By creating a new DeploymentVersion")
			deploymentVersion := &kyaninusv1.DeploymentVersion{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "deploymentversions.kyaninus.codepraxis.com/v1",
					Kind:       "DeploymentVersion",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deployversion1",
					Namespace: DeployNamespace,
				},
				Spec: kyaninusv1.DeploymentVersionSpec{},
			}

			kind := reflect.TypeOf(kyaninusv1.DeploymentVersion{}).Name()
			gvk := kyaninusv1.GroupVersion.WithKind(kind)

			metav1.NewControllerRef(deploymentVersion, gvk)

			Expect(k8sClient.Create(ctx, deploymentVersion)).Should(Succeed())

			/*
				After creating this DeploymentVersion, let's check that the DeploymentVersion's Spec fields match what we passed in.
				Note that, because the k8s apiserver may not have finished creating a CronJob after our `Create()` call from earlier, we will use Gomega’s Eventually() testing function instead of Expect() to give the apiserver an opportunity to finish creating our CronJob.

				`Eventually()` will repeatedly run the function provided as an argument every interval seconds until
				(a) the function’s output matches what’s expected in the subsequent `Should()` call, or
				(b) the number of attempts * interval period exceed the provided timeout value.

				In the examples below, timeout and interval are Go Duration values of our choosing.
			*/

			deployVersionLookupKey := types.NamespacedName{Name: "deployversion1", Namespace: DeployNamespace}
			createdDeployVersion := &kyaninusv1.DeploymentVersion{}

			// We'll need to retry getting this newly created DeploymentVersion, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, deployVersionLookupKey, createdDeployVersion)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			//Expect(createdDeployVersion.Spec.Schedule).Should(Equal("1 * * * *"))

			/*
				Now that we've created a DeploymentVersion in our test cluster, the next step is to write a test that actually tests our CronJob controller’s behavior.
				Let’s test the DeploymentVersion controller’s logic
			*/
			/*
				By("By checking the CronJob has zero active Jobs")
				Consistently(func() (int, error) {
					err := k8sClient.Get(ctx, deployVersionLookupKey, createdDeployVersion)
					if err != nil {
						return -1, err
					}
					//return len(createdDeployVersion.Status.Active), nil
					return 1, nil
				}, duration, interval).Should(Equal(0))
			*/
			/*
				Next, we actually create a stubbed Deployment that will belong to our DeploymentVersion, as well as its downstream template specs.

				We then take the stubbed Deployment and set its owner reference to point to our test DeploymentVersion.
				This ensures that the test Deployment belongs to, and is tracked by, our test DeploymentVersion.
				Once that’s done, we create our new Deployment instance.
			*/
			By("By creating a new Deployment")
			testDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deploy1",
					Namespace: DeployNamespace,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &labelSelector,
					Template: v1.PodTemplateSpec{
						// For simplicity, we only ll out the required fields.
						ObjectMeta: metav1.ObjectMeta{Name: DeployName, Namespace: DeployNamespace, Labels: matchLabels},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "test-container",
									Image: "test-image",
								},
							},
							RestartPolicy: v1.RestartPolicyAlways,
						},
					},
				},
				Status: appsv1.DeploymentStatus{
					Replicas: 1,
				},
			}

			// Note that your CronJob’s GroupVersionKind is required to set up this owner reference.
			//kind := reflect.TypeOf(kyaninusv1.DeploymentVersion{}).Name()
			//gvk := kyaninusv1.GroupVersion.WithKind(kind)

			//controllerRef := metav1.NewControllerRef(createdDeployVersion, gvk)

			//testDeployment.SetOwnerReferences([]metav1.OwnerReference{*controllerRef})
			Expect(k8sClient.Create(ctx, testDeployment)).Should(Succeed())
			/*
				Adding this Job to our test CronJob should trigger our controller’s reconciler logic.
				After that, we can write a test that evaluates whether our controller eventually updates our CronJob’s Status field as expected!
			*/
			/*
				By("By checking that the CronJob has one active Job")
				Eventually(func() ([]string, error) {
					err := k8sClient.Get(ctx, deployVersionLookupKey, createdDeployVersion)
					if err != nil {
						return nil, err
					}

					names := []string{}
					for _, job := range createdDeployVersion.Status.Active {
						names = append(names, job.Name)
					}
					return names, nil
				}, timeout, interval).Should(ConsistOf(JobName), "should list our active job %s in the active jobs list in status", JobName)
			*/
		})
	})

})

/*
	After writing all this code, you can run `go test ./...` in your `controllers/` directory again to run your new test!
*/
