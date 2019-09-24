package util

import (
	"os"

	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	fv1 "github.com/fission/fission/pkg/apis/fission.io/v1"
)

func GetIngressSpec(namespace string, trigger *fv1.HTTPTrigger) *v1beta1.Ingress {
	// TODO: remove backward compatibility
	host, path := trigger.Spec.Host, trigger.Spec.RelativeURL
	if len(trigger.Spec.IngressConfig.Rules) > 0 {
		host, path = trigger.Spec.IngressConfig.Rules[0].Host, trigger.Spec.IngressConfig.Rules[0].Path
	}

	ing := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Labels: GetDeployLabels(trigger),
			Name:   trigger.Metadata.Name,
			// The Ingress NS MUST be same as Router NS, check long discussion:
			// https://github.com/kubernetes/kubernetes/issues/17088
			// We need to revisit this in future, once Kubernetes supports cross namespace ingress
			Namespace:   namespace,
			Annotations: trigger.Spec.IngressConfig.Annotations,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: host,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Backend: v1beta1.IngressBackend{
										ServiceName: "router",
										ServicePort: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 80,
										},
									},
									Path: path,
								},
							},
						},
					},
				},
			},
		},
	}
	return ing
}

func GetDeployLabels(trigger *fv1.HTTPTrigger) map[string]string {
	return map[string]string{
		"triggerName":      trigger.Metadata.Name,
		"functionName":     trigger.Spec.FunctionReference.Name,
		"triggerNamespace": trigger.Metadata.Namespace,
	}
}

//func GetPodNamespace() string {
//	podNamespace := os.Getenv("POD_NAMESPACE")
//	if podNamespace == "" {
//		podNamespace = "fission"
//	}
//
//}
