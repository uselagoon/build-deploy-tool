package templating

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkv1 "k8s.io/api/networking/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)


func GenerateOauth2ProxyTemplate(routes []lagoon.RouteV2, lValues generator.BuildValues) (*metav1.List, error) {
	o2pName := lValues.Namespace + "-oauth2proxy"
	o2pHost := "o2p." + routes[0].Domain
	fmt.Printf("[DEBUG] o2pHost: %s\n", o2pHost)

	ingress := &networkv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      o2pName,
			Labels: map[string]string{
				"app.kubernetes.io/name":  "lagoon-oauth2proxy",
			},
			Annotations: map[string]string{
				"acme.cert-manager.io/http01-ingress-class":   "nginx",
				"kubernetes.io/tls-acme":                       "true",
				"nginx.ingress.kubernetes.io/proxy-buffer-size": "16k",
				"nginx.ingress.kubernetes.io/ssl-redirect":      "false",
			},
		},
		Spec: networkv1.IngressSpec {
			IngressClassName: func() *string { s := "nginx"; return &s }(),
			TLS: []networkv1.IngressTLS{
				{
					Hosts:      []string{o2pHost},
					SecretName: "oauth2proxy-tls",
				},
			},
			Rules: []networkv1.IngressRule {
				{
					Host: o2pHost,
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: func() *networkv1.PathType { pt := networkv1.PathTypePrefix; return &pt }(),
									Backend: networkv1.IngressBackend{
										Service: &networkv1.IngressServiceBackend{
											Name: o2pName,
											Port: networkv1.ServiceBackendPort{
												Number: 4180,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      o2pName,
			Labels: map[string]string{
				"app.kubernetes.io/name": o2pName,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/name": o2pName,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http-4180",
					Port:       4180,
					TargetPort: intstr.FromInt(4180),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lagoon-oauth2proxy",
			Labels: map[string]string{
				"app.kubernetes.io/name":       "lagoon-oauth2proxy",
			},
			Annotations: map[string]string{
				"meta.helm.sh/release-name":      "lagoon-oauth2proxy",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":      "lagoon-oauth2proxy",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":      "lagoon-oauth2proxy",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "oauth2proxy",
							Image: "registry.172.19.0.240.nip.io/library/oauth2-proxy:o2p-authentication",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 4180,
									Name:          "http-4180",
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ping",
										Port: intstr.FromString("http-4180"),
									},
								},
								PeriodSeconds: 10,
								TimeoutSeconds: 1,
							},
							Env: []corev1.EnvVar{
								{Name: "OAUTH2_PROXY_LAGOON_ENDPOINT", Value: "https://lagoon-api.172.19.0.240.nip.io/graphql"},
								{Name: "OAUTH2_PROXY_INSECURE_OIDC_SKIP_ISSUER_VERIFICATION", Value: "true"},
								{Name: "OAUTH2_PROXY_INSECURE_OIDC_ALLOW_UNVERIFIED_EMAIL", Value: "true"},
								//{
								//	Name: "OAUTH2_PROXY_CLIENT_SECRET",
								//	ValueFrom: &corev1.EnvVarSource{
								//		SecretKeyRef: &corev1.SecretKeySelector{
								//			LocalObjectReference: corev1.LocalObjectReference{Name: "lagoon-core-keycloak"},
								//			Key:                  "KEYCLOAK_LAGOON_O2P_CLIENT_SECRET",
								//		},
								//	},
								//},
								//{
								//	Name: "OAUTH2_PROXY_COOKIE_SECRET",
								//	ValueFrom: &corev1.EnvVarSource{
								//		SecretKeyRef: &corev1.SecretKeySelector{
								//			LocalObjectReference: corev1.LocalObjectReference{Name: "lagoon-core-oauth2proxy"},
								//			Key:                  "OAUTH2_PROXY_COOKIE_SECRET",
								//		},
								//	},
								//},
								{Name: "OAUTH2_PROXY_PROVIDER", Value: "oidc"},
								{Name: "OAUTH2_PROXY_CLIENT_ID", Value: "lagoon-oauth2proxy"},
								{Name: "OAUTH2_PROXY_HTTP_ADDRESS", Value: "0.0.0.0:4180"},
								{Name: "OAUTH2_PROXY_OIDC_ISSUER_URL", Value: "https://lagoon-keycloak.172.19.0.240.nip.io/auth/realms/lagoon"},
								{Name: "OAUTH2_PROXY_EMAIL_DOMAINS", Value: "*"},
								{Name: "OAUTH2_PROXY_COOKIE_SECURE", Value: "false"},
								{Name: "OAUTH2_PROXY_COOKIE_DOMAINS", Value: ".nip.io,.sslip.io"},
								{Name: "OAUTH2_PROXY_COOKIE_CSRF_PER_REQUEST", Value: "true"},
								{Name: "OAUTH2_PROXY_WHITELIST_DOMAINS", Value: ".nip.io,.sslip.io"},
								{Name: "OAUTH2_PROXY_SCOPE", Value: "openid email profile"},
								{Name: "OAUTH2_PROXY_CODE_CHALLENGE_METHOD", Value: "S256"},
								{Name: "OAUTH2_PROXY_REVERSE_PROXY", Value: "false"},
								{Name: "OAUTH2_PROXY_PING_PATH", Value: "/ping"},
								{Name: "OAUTH2_PROXY_SILENCE_PING_LOGGING", Value: "true"},
							},
						},
					},
				},
			},
		},
	}
	//deployment := &appsv1.Deployment{
	//	TypeMeta: metav1.TypeMeta{
	//		APIVersion: "apps/v1",
	//		Kind:       "Deployment",
	//    },
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "oauth2-proxy",
	//		Labels:    map[string]string{
	//			"test-label": "test",
	//		},
	//	},
	//	Spec: appsv1.DeploymentSpec{
	//		Selector: &metav1.LabelSelector{
	//			MatchLabels: map[string]string{
	//				"app.kubernetes.io/name":     "test",
	//				"app.kubernetes.io/instance": "test",
	//			},
	//		},
	//		Template: corev1.PodTemplateSpec{
	//			ObjectMeta: metav1.ObjectMeta{
	//				Labels: map[string]string{
	//					"app.kubernetes.io/name":     "test",
	//					"app.kubernetes.io/instance": "test",
	//				},
	//			},
	//			Spec: corev1.PodSpec{
	//				Containers: []corev1.Container{
	//					{
	//						Name:  "oauth2-proxy",
	//						Image: "quay.io/oauth2-proxy/oauth2-proxy:v7.5.1", // Adjust version as needed
	//						Args: []string{
	//							"--provider=lagoon",
	//							"--email-domain=*",
	//							"--upstream=http://127.0.0.1:8080/",
	//							"--http-address=0.0.0.0:4180",
	//						},
	//						Ports: []corev1.ContainerPort{
	//							{
	//								ContainerPort: 4180,
	//							},
	//						},
	//						ReadinessProbe: &corev1.Probe{
	//							ProbeHandler: corev1.ProbeHandler{
	//								HTTPGet: &corev1.HTTPGetAction{
	//									Path: "/ping",
	//									Port: intstr.FromInt(4180),
	//								},
	//							},
	//							InitialDelaySeconds: 5,
	//							PeriodSeconds:       10,
	//						},
	//					},
	//				},
	//			},
	//		},
	//	},
	//}

	templ := &metav1.List{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
		Items: []runtime.RawExtension{
			{Object: deployment},
			{Object: service},
			{Object: ingress},
		},
	}

	return templ, nil
}


func TemplateOauth2Proxy(o2p *metav1.List) ([]byte, error) {
	separator := []byte("---\n")
	var templateYAML []byte
	iBytes, err := yaml.Marshal(o2p)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate template: %v", err)
	}
	restoreResult := append(separator[:], iBytes[:]...)
	templateYAML = append(templateYAML, restoreResult[:]...)
	return templateYAML, nil
}
