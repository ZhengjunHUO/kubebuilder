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

		When("the service is up", func() {
			var (
				deploy appsv1.Deployment
				svc    corev1.Service
				hpa    asv1.HorizontalPodAutoscaler
			)

			BeforeEach(func() {
				Eventually(func() error {
					return k8sClient.Get(ctx, deployNsn, &deploy)
				}, timeout, interval).Should(BeNil())

				Eventually(func() error {
					return k8sClient.Get(ctx, svcNsn, &svc)
				}, timeout, interval).Should(BeNil())

				Eventually(func() error {
					return k8sClient.Get(ctx, hpaNsn, &hpa)
				}, timeout, interval).Should(BeNil())
			})

			When("the deploy's replicas changed", func() {
				const replicas = 3

				BeforeEach(func() {
					deploy.Status.Replicas = replicas
					Expect(k8sClient.Status().Update(ctx, &deploy)).To(Succeed())
				})

				Specify("Replicas in Fufu's status changed", func() {
					Eventually(func() bool {
						fufu := &catv1alpha2.Fufu{}
						if err := k8sClient.Get(ctx, nsn, fufu); err != nil {
							return false
						}
						return fufu.Status.Replicas == replicas
					}, timeout, interval).Should(BeTrue())
				})
			})

			When("the svc's external ip changed", func() {
				const extIP = "10.10.10.10"

				BeforeEach(func() {
					svc.Status.LoadBalancer = corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{
								IP: extIP,
							},
						},
					}
					Expect(k8sClient.Status().Update(ctx, &svc)).To(Succeed())
				})

				Specify("ExternalIP in Fufu's status changed", func() {
					Eventually(func() bool {
						fufu := &catv1alpha2.Fufu{}
						if err := k8sClient.Get(ctx, nsn, fufu); err != nil {
							return false
						}
						return fufu.Status.ExternalIP == extIP
					}, timeout, interval).Should(BeTrue())
				})
			})

			Context("Check deploy strategy", func() {
				const (
					originalStrategy = appsv1.RollingUpdateDeploymentStrategyType
					modifiedStrategy = appsv1.RecreateDeploymentStrategyType
				)

				BeforeEach(func() {
					Expect(deploy.Spec.Strategy.Type).To(Equal(originalStrategy))
				})

				When("deploy's strategy changed manually", func() {
					BeforeEach(func() {
						deploy.Spec.Strategy = appsv1.DeploymentStrategy{
							Type: modifiedStrategy,
						}
						Expect(k8sClient.Update(ctx, &deploy)).To(Succeed())
					})

					It("deploy's strategy restored by controller", func() {
						Eventually(func() bool {
							d := &appsv1.Deployment{}
							if err := k8sClient.Get(ctx, deployNsn, d); err != nil {
								return false
							}
							return d.Spec.Strategy.Type == originalStrategy
						}, timeout, interval).Should(BeTrue())

					})
				})
			})

			Context("Check service port", func() {
				const (
					originalPort int32 = 80
					modifiedPort int32 = 8000
				)

				BeforeEach(func() {
					Expect(svc.Spec.Ports[0].Port).To(Equal(originalPort))
				})

				When("service port changed manually", func() {
					BeforeEach(func() {
						svc.Spec.Ports[0].Port = modifiedPort
						Expect(k8sClient.Update(ctx, &svc)).To(Succeed())
					})

					It("service port restored by controller", func() {
						Eventually(func() bool {
							s := &corev1.Service{}
							if err := k8sClient.Get(ctx, svcNsn, s); err != nil {
								return false
							}
							return s.Spec.Ports[0].Port == originalPort
						}, timeout, interval).Should(BeTrue())
					})
				})
			})

			Context("Check hpa's replica", func() {
				var (
					originalMinRep int32 = 2
					modifiedMinRep int32 = 3
				)

				BeforeEach(func() {
					Expect(*hpa.Spec.MinReplicas).To(Equal(originalMinRep))
				})

				When("hpa's replica changed manually", func() {
					BeforeEach(func() {
						hpa.Spec.MinReplicas = &modifiedMinRep
						Expect(k8sClient.Update(ctx, &hpa)).To(Succeed())
					})

					It("hpa's replica restored by controller", func() {
						Eventually(func() bool {
							h := &asv1.HorizontalPodAutoscaler{}
							if err := k8sClient.Get(ctx, hpaNsn, h); err != nil {
								return false
							}
							return *h.Spec.MinReplicas == originalMinRep
						}, timeout, interval).Should(BeTrue())
					})
				})
			})
		})
	})
})
