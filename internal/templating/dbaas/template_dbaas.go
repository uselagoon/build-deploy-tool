package dbaas

import (
	"fmt"

	mariadbv1 "github.com/amazeeio/dbaas-operator/apis/mariadb/v1"
	mongodbv1 "github.com/amazeeio/dbaas-operator/apis/mongodb/v1"
	postgresv1 "github.com/amazeeio/dbaas-operator/apis/postgres/v1"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"sigs.k8s.io/yaml"
)

var dbaasTypes = []string{
	"mariadb-dbaas",
	"mongodb-dbaas",
	"postgres-dbaas",
}

// GenerateDBaaSTemplate generates the lagoon template to apply.
func GenerateDBaaSTemplate(
	lValues generator.BuildValues,
) ([]byte, error) {
	separator := []byte("---\n")
	var result []byte

	// add the default labels
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "build-deploy-tool",
		"lagoon.sh/project":            lValues.Project,
		"lagoon.sh/environment":        lValues.Environment,
		"lagoon.sh/environmentType":    lValues.EnvironmentType,
		"lagoon.sh/buildType":          lValues.BuildType,
	}

	// add the default annotations
	annotations := map[string]string{
		"lagoon.sh/version": lValues.LagoonVersion,
	}

	// add any additional labels
	additionalLabels := map[string]string{}
	additionalAnnotations := map[string]string{}
	if lValues.BuildType == "branch" {
		additionalAnnotations["lagoon.sh/branch"] = lValues.Branch
	} else if lValues.BuildType == "pullrequest" {
		additionalAnnotations["lagoon.sh/prNumber"] = lValues.PRNumber
		additionalAnnotations["lagoon.sh/prHeadBranch"] = lValues.PRHeadBranch
		additionalAnnotations["lagoon.sh/prBaseBranch"] = lValues.PRBaseBranch

	}

	for _, serviceValues := range lValues.Services {
		if helpers.Contains(dbaasTypes, serviceValues.Type) {
			var consumerBytes []byte
			additionalLabels["app.kubernetes.io/name"] = serviceValues.Type
			additionalLabels["app.kubernetes.io/instance"] = serviceValues.OverrideName
			additionalLabels["lagoon.sh/template"] = fmt.Sprintf("%s-%s", serviceValues.Type, "0.1.0")
			additionalLabels["lagoon.sh/service"] = serviceValues.OverrideName
			additionalLabels["lagoon.sh/service-type"] = serviceValues.Type
			switch serviceValues.Type {
			case "mariadb-dbaas":
				{
					mariaDBConsumer := &mariadbv1.MariaDBConsumer{
						TypeMeta: metav1.TypeMeta{
							Kind:       "MariaDBConsumer",
							APIVersion: mariadbv1.GroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: serviceValues.OverrideName,
						},
						Spec: mariadbv1.MariaDBConsumerSpec{
							Environment: serviceValues.DBaaSEnvironment,
						},
					}
					mariaDBConsumer.ObjectMeta.Labels = labels
					mariaDBConsumer.ObjectMeta.Annotations = annotations
					for key, value := range additionalLabels {
						mariaDBConsumer.ObjectMeta.Labels[key] = value
					}
					// add any additional annotations
					for key, value := range additionalAnnotations {
						mariaDBConsumer.ObjectMeta.Annotations[key] = value
					}
					// validate any annotations
					if err := apivalidation.ValidateAnnotations(mariaDBConsumer.ObjectMeta.Annotations, nil); err != nil {
						if len(err) != 0 {
							return nil, fmt.Errorf("the annotations for %s are not valid: %v", serviceValues.OverrideName, err)
						}
					}
					// validate any labels
					if err := metavalidation.ValidateLabels(mariaDBConsumer.ObjectMeta.Labels, nil); err != nil {
						if len(err) != 0 {
							return nil, fmt.Errorf("the labels for %s are not valid: %v", serviceValues.OverrideName, err)
						}
					}

					// check length of labels
					err := helpers.CheckLabelLength(mariaDBConsumer.ObjectMeta.Labels)
					if err != nil {
						return nil, err
					}
					consumerBytes, err = yaml.Marshal(mariaDBConsumer)
					if err != nil {
						return nil, err
					}
				}
			case "mongodb-dbaas":
				{
					mongodbConsumer := &mongodbv1.MongoDBConsumer{
						TypeMeta: metav1.TypeMeta{
							Kind:       "MongoDBConsumer",
							APIVersion: mongodbv1.GroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: serviceValues.OverrideName,
						},
						Spec: mongodbv1.MongoDBConsumerSpec{
							Environment: serviceValues.DBaaSEnvironment,
						},
					}
					mongodbConsumer.ObjectMeta.Labels = labels
					mongodbConsumer.ObjectMeta.Annotations = annotations
					for key, value := range additionalLabels {
						mongodbConsumer.ObjectMeta.Labels[key] = value
					}
					// add any additional annotations
					for key, value := range additionalAnnotations {
						mongodbConsumer.ObjectMeta.Annotations[key] = value
					}
					// validate any annotations
					if err := apivalidation.ValidateAnnotations(mongodbConsumer.ObjectMeta.Annotations, nil); err != nil {
						if len(err) != 0 {
							return nil, fmt.Errorf("the annotations for %s are not valid: %v", serviceValues.OverrideName, err)
						}
					}
					// validate any labels
					if err := metavalidation.ValidateLabels(mongodbConsumer.ObjectMeta.Labels, nil); err != nil {
						if len(err) != 0 {
							return nil, fmt.Errorf("the labels for %s are not valid: %v", serviceValues.OverrideName, err)
						}
					}
					// check length of labels
					err := helpers.CheckLabelLength(mongodbConsumer.ObjectMeta.Labels)
					if err != nil {
						return nil, err
					}
					consumerBytes, err = yaml.Marshal(mongodbConsumer)
					if err != nil {
						return nil, err
					}
				}
			case "postgres-dbaas":
				{
					postgresqlConsumer := &postgresv1.PostgreSQLConsumer{
						TypeMeta: metav1.TypeMeta{
							Kind:       "PostgreSQLConsumer",
							APIVersion: postgresv1.GroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: serviceValues.OverrideName,
						},
						Spec: postgresv1.PostgreSQLConsumerSpec{
							Environment: serviceValues.DBaaSEnvironment,
						},
					}
					postgresqlConsumer.ObjectMeta.Labels = labels
					postgresqlConsumer.ObjectMeta.Annotations = annotations
					for key, value := range additionalLabels {
						postgresqlConsumer.ObjectMeta.Labels[key] = value
					}
					// add any additional annotations
					for key, value := range additionalAnnotations {
						postgresqlConsumer.ObjectMeta.Annotations[key] = value
					}
					// validate any annotations
					if err := apivalidation.ValidateAnnotations(postgresqlConsumer.ObjectMeta.Annotations, nil); err != nil {
						if len(err) != 0 {
							return nil, fmt.Errorf("the annotations for %s are not valid: %v", serviceValues.OverrideName, err)
						}
					}
					// validate any labels
					if err := metavalidation.ValidateLabels(postgresqlConsumer.ObjectMeta.Labels, nil); err != nil {
						if len(err) != 0 {
							return nil, fmt.Errorf("the labels for %s are not valid: %v", serviceValues.OverrideName, err)
						}
					}

					// check length of labels
					err := helpers.CheckLabelLength(postgresqlConsumer.ObjectMeta.Labels)
					if err != nil {
						return nil, err
					}
					consumerBytes, err = yaml.Marshal(postgresqlConsumer)
					if err != nil {
						return nil, err
					}
				}

			}
			// @TODO: we should review this in the future when we stop doing `kubectl apply` in the builds :)
			// add the seperator to the template so that it can be `kubectl apply` in bulk as part
			// of the current build process
			// join all dbaas-consumer templates together
			restoreResult := append(separator[:], consumerBytes[:]...)
			result = append(result, restoreResult[:]...)
		}
	}
	return result, nil
}
