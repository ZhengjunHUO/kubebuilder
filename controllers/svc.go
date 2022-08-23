package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	catv1alpha2 "github.com/ZhengjunHUO/kubebuilder/api/v1alpha2"
)

func (r *FufuReconciler) updateSvc(fufu *catv1alpha2.Fufu, ctx context.Context) error {
	loggr := log.FromContext(ctx)

	wanted := r.createSvc(fufu)

	had := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: wanted.ObjectMeta.Name, Namespace: wanted.ObjectMeta.Namespace}, had); err == nil {
		if !equality.Semantic.DeepDerivative(wanted.Spec.Selector, had.Spec.Selector) {
			loggr.Info("A diff was found, update svc ...")
			ctrutil.SetControllerReference(fufu, wanted, r.Scheme)
			if err = r.Update(ctx, wanted); err != nil {
				return err
			}
			//r.Recorder.Event(fufu, corev1.EventTypeNormal, "svc-updated", "Service updated")
			loggr.Info("Service updated")
		}
		return nil
	} else {
		if err = client.IgnoreNotFound(err); err != nil {
			return err
		}

		loggr.Info("Create svc ...")
		ctrutil.SetControllerReference(fufu, wanted, r.Scheme)
		if err = r.Create(ctx, wanted); err != nil {
			loggr.Error(err, "failed to create svc")
		}

		//r.Recorder.Event(fufu, corev1.EventTypeNormal, "svc-created", "Service created")
		loggr.Info("Service created")
		return nil
	}
}

func (r *FufuReconciler) createSvc(fufu *catv1alpha2.Fufu) *corev1.Service {
	name := fufu.Name + "-svc"
	selectName := fufu.Name + "-deploy"
	labels := map[string]string{
		"app": selectName,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: fufu.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}
}
