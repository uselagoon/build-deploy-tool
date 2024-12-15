package templating

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/servicetypes"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

// GenerateServiceTemplate generates the lagoon template to apply.
func GenerateServiceTemplate(
	buildValues generator.BuildValues,
) ([]corev1.Service, error) {
	var services []corev1.Service
	// add the default labels
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "build-deploy-tool",
		"lagoon.sh/project":            buildValues.Project,
		"lagoon.sh/environment":        buildValues.Environment,
		"lagoon.sh/environmentType":    buildValues.EnvironmentType,
		"lagoon.sh/buildType":          buildValues.BuildType,
	}

	// add the default annotations
	annotations := map[string]string{
		"lagoon.sh/version": buildValues.LagoonVersion,
	}

	// add any additional labels
	if buildValues.BuildType == "branch" {
		annotations["lagoon.sh/branch"] = buildValues.Branch
	} else if buildValues.BuildType == "pullrequest" {
		annotations["lagoon.sh/prNumber"] = buildValues.PRNumber
		annotations["lagoon.sh/prHeadBranch"] = buildValues.PRHeadBranch
		annotations["lagoon.sh/prBaseBranch"] = buildValues.PRBaseBranch

	}

	// check linked services
	checkedServices := LinkedServiceCalculator(buildValues.Services)

	// for all the services that the build values generated
	// iterate over them and generate any kubernetes services
	for _, serviceValues := range checkedServices {
		if val, ok := servicetypes.ServiceTypes[serviceValues.Type]; ok {
			serviceType := &servicetypes.ServiceType{}
			helpers.DeepCopy(val, serviceType)
			service, err := GenerateService(serviceType, serviceValues, labels, annotations)
			if err != nil {
				return nil, err
			}
			if service != nil {
				services = append(services, *service)
			}
		}
	}
	return services, nil
}

func GenerateService(serviceType *servicetypes.ServiceType, serviceValues generator.ServiceValues, labels, annotations map[string]string) (*corev1.Service, error) {
	if serviceValues.AdditionalServicePorts == nil && serviceType.Ports.Ports == nil {
		// there are no additional ports provided, and this servicetype has no default ports associated to it
		// just drop out
		return nil, nil
	}
	additionalLabels := map[string]string{}
	additionalAnnotations := map[string]string{}

	additionalLabels["app.kubernetes.io/name"] = serviceType.Name
	additionalLabels["app.kubernetes.io/instance"] = serviceValues.OverrideName
	additionalLabels["lagoon.sh/template"] = fmt.Sprintf("%s-%s", serviceType.Name, "0.1.0")
	additionalLabels["lagoon.sh/service"] = serviceValues.OverrideName
	additionalLabels["lagoon.sh/service-type"] = serviceType.Name
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: corev1.SchemeGroupVersion.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceValues.OverrideName,
		},
	}

	labelsCopy := &map[string]string{}
	helpers.DeepCopy(labels, labelsCopy)
	annotationsCopy := &map[string]string{}
	helpers.DeepCopy(annotations, annotationsCopy)

	for key, value := range additionalLabels {
		(*labelsCopy)[key] = value
	}
	// add any additional annotations
	for key, value := range additionalAnnotations {
		(*annotationsCopy)[key] = value
	}
	service.ObjectMeta.Labels = *labelsCopy
	service.ObjectMeta.Annotations = *annotationsCopy
	// validate any annotations
	if err := apivalidation.ValidateAnnotations(service.ObjectMeta.Annotations, nil); err != nil {
		if len(err) != 0 {
			return nil, fmt.Errorf("the annotations for %s are not valid: %v", serviceValues.OverrideName, err)
		}
	}
	// validate any labels
	if err := metavalidation.ValidateLabels(service.ObjectMeta.Labels, nil); err != nil {
		if len(err) != 0 {
			return nil, fmt.Errorf("the labels for %s are not valid: %v", serviceValues.OverrideName, err)
		}
	}
	// check length of labels
	err := helpers.CheckLabelLength(service.ObjectMeta.Labels)
	if err != nil {
		return nil, err
	}

	// start service template
	if serviceType.Ports.CanChangePort {
		if serviceValues.ServicePort != 0 {
			serviceType.Ports.Ports[0].Port = serviceValues.ServicePort
		}
	}

	// start compose service port override templating here
	if serviceValues.AdditionalServicePorts != nil {
		// blank out the provided ports from the servicetype
		serviceType.Ports.Ports = nil
		// if the service is set to consume the additional services only, then generate those here
		for _, addPort := range serviceValues.AdditionalServicePorts {
			port := corev1.ServicePort{
				Protocol: corev1.ProtocolTCP,
				Name:     fmt.Sprintf("tcp-%d", addPort.ServicePort.Target),
				TargetPort: intstr.IntOrString{
					StrVal: fmt.Sprintf("tcp-%d", addPort.ServicePort.Target),
					Type:   intstr.String,
				},
				Port: int32(addPort.ServicePort.Target),
			}
			// set protocol to anything but tcp if required
			switch addPort.ServicePort.Protocol {
			case "udp":
				port.Name = fmt.Sprintf("udp-%d", addPort.ServicePort.Target)
				port.TargetPort = intstr.IntOrString{
					StrVal: fmt.Sprintf("udp-%d", addPort.ServicePort.Target),
					Type:   intstr.String,
				}
				port.Protocol = corev1.ProtocolUDP
			}
			serviceType.Ports.Ports = append(serviceType.Ports.Ports, port)
		}
	}
	// end compose service port override templating here

	service.Spec = corev1.ServiceSpec{
		Ports: serviceType.Ports.Ports,
	}

	service.Spec.Selector = map[string]string{
		"app.kubernetes.io/name":     serviceType.Name,
		"app.kubernetes.io/instance": serviceValues.OverrideName,
	}
	// end service template
	return service, nil
}

func GenerateServiceBackendPort(addPort generator.AdditionalServicePort) networkv1.ServiceBackendPort {
	switch addPort.ServicePort.Protocol {
	case "udp":
		return networkv1.ServiceBackendPort{
			Name: fmt.Sprintf("udp-%d", addPort.ServicePort.Target),
		}
	default:
		return networkv1.ServiceBackendPort{
			Name: fmt.Sprintf("tcp-%d", addPort.ServicePort.Target),
		}
	}
}

func TemplateService(item corev1.Service) ([]byte, error) {
	separator := []byte("---\n")
	iBytes, err := yaml.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate template: %v", err)
	}
	templateYAML := append(separator[:], iBytes[:]...)
	return templateYAML, nil
}
