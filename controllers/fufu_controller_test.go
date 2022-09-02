package controllers

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	catv1alpha2 "github.com/ZhengjunHUO/kubebuilder/api/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	asv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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