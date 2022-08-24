package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	catv1alpha2 "github.com/ZhengjunHUO/kubebuilder/api/v1alpha2"
	asv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *FufuReconciler) updateHpa(fufu *catv1alpha2.Fufu, ctx context.Context) error {
	loggr := log.FromContext(ctx)

	wanted := r.createHpa(fufu)

	had := &asv1.HorizontalPodAutoscaler{}
	if err := r.Get(ctx, types.NamespacedName{Name: wanted.ObjectMeta.Name, Namespace: wanted.ObjectMeta.Namespace}, had); err == nil {
		if !equality.Semantic.DeepDerivative(wanted.Spec, had.Spec) {
			loggr.Info("A diff was found, update hpa ...")
			ctrutil.SetControllerReference(fufu, wanted, r.Scheme)
			if err = r.Update(ctx, wanted); err != nil {
				return err
			}
			loggr.Info("Hpa updated")
		}
		return nil
	} else {
		if err = client.IgnoreNotFound(err); err != nil {
			return err
		}

		loggr.Info("Create hpa ...")
		ctrutil.SetControllerReference(fufu, wanted, r.Scheme)
		if err = r.Create(ctx, wanted); err != nil {
			loggr.Error(err, "failed to create hpa")
		}

		loggr.Info("Hpa created")
		return nil
	}
}

func (r *FufuReconciler) createHpa(fufu *catv1alpha2.Fufu) *asv1.HorizontalPodAutoscaler {
	name := fufu.Name + "-hpa"
	deployName := fufu.Name + "-deploy"
	var minReplicas int32 = 2
	var cpuThreshold int32 = 60

	return &asv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: fufu.Namespace,
		},
		Spec: asv1.HorizontalPodAutoscalerSpec{
			MinReplicas:                    &minReplicas,
			MaxReplicas:                    5,
			TargetCPUUtilizationPercentage: &cpuThreshold,
			ScaleTargetRef: asv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
				Name:       deployName,
			},
		},
	}
}
