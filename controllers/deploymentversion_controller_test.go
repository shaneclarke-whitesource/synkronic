package controllers

import (
	"context"
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

			createdDeploy := &appsv1.Deployment{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: DeployName, Namespace: DeployNamespace}, createdDeploy)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("By creating a new DeploymentVersion")

			deploymentVersion := &kyaninusv1.DeploymentVersion{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "kyaninus.codepraxis.com/v1",
					Kind:       "DeploymentVersion",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deployversion1",
					Namespace: DeployNamespace,
				},
				Spec: kyaninusv1.DeploymentVersionSpec{
					Name:      DeployName,
					Namespace: DeployNamespace,
					DeploymentSpec: appsv1.DeploymentSpec{
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
				},
			}

			/*
				s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme,
					scheme.Scheme)

				var b bytes.Buffer
				buffwriter := bufio.NewWriter(&b)

				s.Encode(deploymentVersion, buffwriter)

				buffwriter.Flush()
				fmt.Println(b.String())

				deploymentVersion2 := &kyaninusv1.DeploymentVersion{}

				s.Decode(b.Bytes(), &schema.GroupVersionKind{Group: "kyaninus.codepraxis.com", Version: "v1", Kind: "DeploymentVersion"}, deploymentVersion2)
			*/
			/*
				bytes, err := json.Marshal(deploymentVersion)
				if err != nil {
					fmt.Println("Can't serislize", deploymentVersion)
				}

				asdf := string(bytes)
				fmt.Printf("%v => %v, '%v'\n", deploymentVersion, bytes, asdf)

					kind := reflect.TypeOf(kyaninusv1.DeploymentVersion{}).Name()

					gvk := kyaninusv1.GroupVersion.WithKind(kind)

					metav1.NewControllerRef(deploymentVersion, gvk)
			*/

			//err1 := k8sClient.Create(ctx, deploymentVersion)
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

			//Check new deployment has been created per spec
			createdDeploy2 := &appsv1.Deployment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: DeployName + "1", Namespace: DeployNamespace}, createdDeploy2)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})
	})

})

/*
	After writing all this code, you can run `go test ./...` in your `controllers/` directory again to run your new test!
*/
