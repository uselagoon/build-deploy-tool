package services

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/servicetypes"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

var separator = []byte("---\n")

// GenerateServiceTemplate generates the lagoon template to apply.
func GenerateServiceTemplate(
	buildValues generator.BuildValues,
) ([]byte, error) {

	var result []byte

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
	additionalLabels := map[string]string{}
	additionalAnnotations := map[string]string{}
	if buildValues.BuildType == "branch" {
		additionalAnnotations["lagoon.sh/branch"] = buildValues.Branch
	} else if buildValues.BuildType == "pullrequest" {
		additionalAnnotations["lagoon.sh/prNumber"] = buildValues.PRNumber
		additionalAnnotations["lagoon.sh/prHeadBranch"] = buildValues.PRHeadBranch
		additionalAnnotations["lagoon.sh/prBaseBranch"] = buildValues.PRBaseBranch

	}

	// check linked services
	checkedServices := LinkedServiceCalculator(buildValues.Services)

	// for all the services that the build values generated
	// iterate over them and generate any kubernetes services
	for _, serviceValues := range checkedServices {
		if val, ok := servicetypes.ServiceTypes[serviceValues.Type]; ok {
			serviceType := &servicetypes.ServiceType{}
			helpers.DeepCopy(val, serviceType)
			restoreResult, err := GenerateService(result, serviceType, serviceValues, labels, annotations, additionalLabels, additionalAnnotations)
			if err != nil {
				return nil, err
			}
			result = append(result, restoreResult[:]...)
		}
	}
	return result, nil
}

func GenerateService(result []byte, serviceType *servicetypes.ServiceType, serviceValues generator.ServiceValues, labels, annotations, additionalLabels, additionalAnnotations map[string]string) ([]byte, error) {
	if serviceValues.AdditionalServicePorts == nil && serviceType.Ports.Ports == nil {
		// there are no additional ports provided, and this servicetype has no default ports associated to it
		// just drop out
		return nil, nil
	}
	var serviceBytes []byte

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
	service.ObjectMeta.Labels = labels
	service.ObjectMeta.Annotations = annotations
	for key, value := range additionalLabels {
		service.ObjectMeta.Labels[key] = value
	}
	// add any additional annotations
	for key, value := range additionalAnnotations {
		service.ObjectMeta.Annotations[key] = value
	}
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
				Name:     addPort.ServiceName,
				TargetPort: intstr.IntOrString{
					StrVal: addPort.ServiceName,
					Type:   intstr.String,
				},
				Port: int32(addPort.ServicePort.Target),
			}
			// set protocol to anything but tcp if required
			switch addPort.ServicePort.Protocol {
			case "udp":
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

	serviceBytes, err = yaml.Marshal(service)
	if err != nil {
		return nil, err
	}
	// @TODO: we should review this in the future when we stop doing `kubectl apply` in the builds :)
	// add the seperator to the template so that it can be `kubectl apply` in bulk as part
	// of the current build process
	// join all dbaas-consumer templates together
	restoreResult := append(separator[:], serviceBytes[:]...)
	return restoreResult, nil
}
