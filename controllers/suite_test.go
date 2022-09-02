/*
Copyright 2022 huo.

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
	"path/filepath"
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	catv1alpha2 "github.com/ZhengjunHUO/kubebuilder/api/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	asv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = catv1alpha2.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sMgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&FufuReconciler{
		Client: k8sMgr.GetClient(),
		Scheme: k8sMgr.GetScheme(),
	}).SetupWithManager(k8sMgr)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sMgr.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

}, 60)

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Test controller", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Second * 1
	)

	var (
		nsn = types.NamespacedName{
			Name:      "fufu",
			Namespace: "default",
		}
		deployNsn = types.NamespacedName{
			Name:      "fufu-deploy",
			Namespace: "default",
		}
		svcNsn = types.NamespacedName{
			Name:      "fufu-svc",
			Namespace: "default",
		}

		hpaNsn = types.NamespacedName{
			Name:      "fufu-hpa",
			Namespace: "default",
		}
	)

	When("create custom resource fufu", func() {
		var (
			created                catv1alpha2.Fufu
			expectedOwnerReference metav1.OwnerReference
		)

		BeforeEach(func() {
			created = catv1alpha2.Fufu{
				ObjectMeta: metav1.ObjectMeta{
					Name:      nsn.Name,
					Namespace: nsn.Namespace,
				},
				Spec: catv1alpha2.FufuSpec{
					Color:  "orange",
					Weight: "5kg",
					Age:    6,
					Info: catv1alpha2.AdditionalInfo{
						Breed:      "stray",
						Vaccinated: false,
					},
				},
			}

			Expect(k8sClient.Create(ctx, &created)).Should(Succeed())

			expectedOwnerReference = metav1.OwnerReference{
				Kind:               "Fufu",
				APIVersion:         "cat.huozj.io/v1alpha2",
				Name:               nsn.Name,
				UID:                created.UID,
				Controller:         func(v bool) *bool { return &v }(true),
				BlockOwnerDeletion: func(v bool) *bool { return &v }(true),
			}
		})

		AfterEach(func() {
			k8sClient.Delete(ctx, &created)
		})

		Specify("create full fufu stack", func() {
			By("create deploy for fufu", func() {
				var deploy appsv1.Deployment
				Eventually(func() error {
					return k8sClient.Get(ctx, deployNsn, &deploy)
				}, timeout, interval).Should(BeNil())
				Expect(deploy.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))
			})

			By("create associated svc for deploy", func() {
				var svc corev1.Service
				Eventually(func() error {
					return k8sClient.Get(ctx, svcNsn, &svc)
				}, timeout, interval).Should(BeNil())
				Expect(svc.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))
			})

			By("create associated hpa for deploy", func() {
				var hpa asv1.HorizontalPodAutoscaler
				Eventually(func() error {
					return k8sClient.Get(ctx, hpaNsn, &hpa)
				}, timeout, interval).Should(BeNil())
				Expect(hpa.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))
			})
		})
	})
})
