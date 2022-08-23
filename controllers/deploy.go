package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	catv1alpha2 "github.com/ZhengjunHUO/kubebuilder/api/v1alpha2"
)

func (r *FufuReconciler) updateDeploy(fufu *catv1alpha2.Fufu, ctx context.Context) error {
	loggr := log.FromContext(ctx)

	wanted := r.createDeploy(fufu)

	had := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: wanted.ObjectMeta.Name, Namespace: wanted.ObjectMeta.Namespace}, had); err == nil {
		if had.Status.Replicas != fufu.Status.Replicas {
			loggr.Info(fmt.Sprintf("Fufu's current replicas: %d", had.Status.Replicas))
			fufu.Status.Replicas = had.Status.Replicas
			if err = r.Status().Update(ctx, fufu); err != nil {
				return err
			}
			//r.Recorder.Eventf(fufu, corev1.EventTypeNormal, "replicas-updated", "Replicas updated to %d", had.Status.Replicas)
			loggr.Info(fmt.Sprintf("Replicas updated to %d", had.Status.Replicas))
		}

		if !equality.Semantic.DeepDerivative(wanted.Spec, had.Spec) {
			loggr.Info("A diff was found, update deploy ...")
			ctrutil.SetControllerReference(fufu, wanted, r.Scheme)
			if err = r.Update(ctx, wanted); err != nil {
				return err
			}
			//r.Recorder.Event(fufu, corev1.EventTypeNormal, "deploy-updated", "Deployment updated")
			loggr.Info("Deployment updated")
		}

		return nil
	} else {
		if err = client.IgnoreNotFound(err); err != nil {
			return err
		}

		loggr.Info("Create deploy ...")
		ctrutil.SetControllerReference(fufu, wanted, r.Scheme)
		if err = r.Create(ctx, wanted); err != nil {
			loggr.Error(err, "failed to create deploy")
		}

		//r.Recorder.Event(fufu, corev1.EventTypeNormal, "deploy-created", "Deployment created")
		loggr.Info("Deployment created")
		return nil
	}
}

func (r *FufuReconciler) createDeploy(fufu *catv1alpha2.Fufu) *appsv1.Deployment {
	name := fufu.Name + "-deploy"
	labels := map[string]string{
		"app": name,
	}
	volName := "homedir"

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: fufu.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: volName,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:  "prepare-webcontent",
							Image: "alpine",
							Command: []string{
								"/bin/sh",
								"-c",
							},
							Args: []string{
								"wget https://raw.githubusercontent.com/ZhengjunHUO/kubebuilder/main/k8s/nginx/index.html.tmpl     && apk add gettext && envsubst '$FUR_COLOR $BREED $AGE $WEIGHT' < index.html.tmpl > /mnt/index.html",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "FUR_COLOR",
									Value: fufu.Spec.Color,
								},
								{
									Name:  "BREED",
									Value: fufu.Spec.Info.Breed,
								},
								{
									Name:  "AGE",
									Value: fmt.Sprintf("%d", fufu.Spec.Age),
								},
								{
									Name:  "WEIGHT",
									Value: fufu.Spec.Weight,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      volName,
									MountPath: "/mnt",
									ReadOnly:  false,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "web",
							Image: "nginx",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      volName,
									MountPath: "/usr/share/nginx/html/index.html",
									SubPath:   "index.html",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
		},
	}
}
