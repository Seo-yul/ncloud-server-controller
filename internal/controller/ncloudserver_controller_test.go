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
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	serverv1 "github.com/Seo-yul/ncloud-server-controller/api/v1"
)

var _ = Describe("NCloudServer Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		ncloudserver := &serverv1.NCloudServer{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind NCloudServer")
			err := k8sClient.Get(ctx, typeNamespacedName, ncloudserver)
			if err != nil && errors.IsNotFound(err) {
				resource := &serverv1.NCloudServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					// TODO(user): Specify other spec details if needed.
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &serverv1.NCloudServer{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance NCloudServer")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &NCloudServerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Mock CLI 경로 설정
			mockCliPath, err := filepath.Abs("../../test/mock_ncloud_cli.sh")
			Expect(err).NotTo(HaveOccurred())
			controllerReconciler.NCloudCliPath = mockCliPath

			// 테스트용 환경변수 설정
			err = os.Setenv("NCLOUD_CLI_PATH", mockCliPath)
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				err := os.Unsetenv("NCLOUD_CLI_PATH")
				Expect(err).NotTo(HaveOccurred())
			}()

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// 리소스 상태 확인
			updatedResource := &serverv1.NCloudServer{}
			err = k8sClient.Get(ctx, typeNamespacedName, updatedResource)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedResource.Status.Phase).To(Equal("Creating"))
		})
	})
})
