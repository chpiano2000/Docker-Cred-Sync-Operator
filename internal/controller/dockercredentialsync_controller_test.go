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

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	daxvov1 "dax.vo/docker-credentail-sync/api/v1"
)

var _ = Describe("DockerCredentialSync Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		dockercredentialsync := &daxvov1.DockerCredentialSync{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind DockerCredentialSync")
			err := k8sClient.Get(ctx, typeNamespacedName, dockercredentialsync)
			if err != nil && errors.IsNotFound(err) {
				resource := &daxvov1.DockerCredentialSync{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: daxvov1.DockerCredentialSyncSpec{
						SourceNamespace:        "cred-store",
						SourceSecretName:       "dockerhub-cred",
						TargetNamespacePrefix:  "dev-",
						RefreshIntervalSeconds: 3,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &daxvov1.DockerCredentialSync{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance DockerCredentialSync")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &DockerCredentialSyncReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying that the status was updated")
			err = k8sClient.Get(ctx, typeNamespacedName, dockercredentialsync)
			Expect(err).NotTo(HaveOccurred())
			Expect(dockercredentialsync.Status.Conditions.Type).To(Equal("Ready"))
			Expect(dockercredentialsync.Status.Conditions.Status).To(Equal(metav1.ConditionTrue))
		})

		It("should handle errors when source secret is not found", func() {
			By("Simulating source secret not found")
			controllerReconciler := &DockerCredentialSyncReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Create a DockerCredentialSync CR with an invalid source secret
			resource := &daxvov1.DockerCredentialSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-error-resource",
					Namespace: "default",
				},
				Spec: daxvov1.DockerCredentialSyncSpec{
					SourceNamespace:        "invalid-namespace",
					SourceSecretName:       "invalid-secret",
					TargetNamespacePrefix:  "dev-",
					RefreshIntervalSeconds: 3,
				},
			}

			// Create the invalid resource
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			// Run the reconcile logic
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-error-resource",
					Namespace: "default",
				},
			})

			// Assert the error case (source secret not found)
			Expect(err).To(HaveOccurred())

			// Clean up the resource after testing
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
	})
})
