package templating

import (
	"fmt"

	mariadbv1 "github.com/amazeeio/dbaas-operator/apis/mariadb/v1"
	mongodbv1 "github.com/amazeeio/dbaas-operator/apis/mongodb/v1"
	postgresv1 "github.com/amazeeio/dbaas-operator/apis/postgres/v1"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"sigs.k8s.io/yaml"

	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
)

var dbaasTypes = []string{
	"mariadb-dbaas",
	"mongodb-dbaas",
	"postgres-dbaas",
}

type DBaaSTemplates struct {
	MariaDB    []mariadbv1.MariaDBConsumer
	MongoDB    []mongodbv1.MongoDBConsumer
	PostgreSQL []postgresv1.PostgreSQLConsumer
}

// GenerateDBaaSTemplate generates the lagoon template to apply.
func GenerateDBaaSTemplate(
	lValues generator.BuildValues,
) (*DBaaSTemplates, error) {
	// separator := []byte("---\n")
	// var result []byte
	var dbaasTemplates DBaaSTemplates

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
			additionalLabels["app.kubernetes.io/name"] = serviceValues.Type
			additionalLabels["app.kubernetes.io/instance"] = serviceValues.Name
			additionalLabels["lagoon.sh/template"] = fmt.Sprintf("%s-%s", serviceValues.Type, "0.1.0")
			additionalLabels["lagoon.sh/service"] = serviceValues.Name
			additionalLabels["lagoon.sh/service-type"] = serviceValues.Type
			switch serviceValues.Type {
			case "mariadb-dbaas":
				{
					mariaDBConsumer := mariadbv1.MariaDBConsumer{
						TypeMeta: metav1.TypeMeta{
							Kind:       "MariaDBConsumer",
							APIVersion: mariadbv1.GroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: serviceValues.Name,
						},
						Spec: mariadbv1.MariaDBConsumerSpec{
							Environment: serviceValues.DBaaSEnvironment,
						},
					}
					mariaDBConsumer.ObjectMeta.Labels = map[string]string{}
					mariaDBConsumer.ObjectMeta.Annotations = map[string]string{}
					for key, value := range labels {
						mariaDBConsumer.ObjectMeta.Labels[key] = value
					}
					for key, value := range annotations {
						mariaDBConsumer.ObjectMeta.Annotations[key] = value
					}
					for key, value := range additionalLabels {
						mariaDBConsumer.ObjectMeta.Labels[key] = value
					}
					for key, value := range additionalAnnotations {
						mariaDBConsumer.ObjectMeta.Annotations[key] = value
					}
					// validate any annotations
					if err := apivalidation.ValidateAnnotations(mariaDBConsumer.ObjectMeta.Annotations, nil); err != nil {
						if len(err) != 0 {
							return nil, fmt.Errorf("the annotations for %s are not valid: %v", serviceValues.Name, err)
						}
					}
					// validate any labels
					if err := metavalidation.ValidateLabels(mariaDBConsumer.ObjectMeta.Labels, nil); err != nil {
						if len(err) != 0 {
							return nil, fmt.Errorf("the labels for %s are not valid: %v", serviceValues.Name, err)
						}
					}

					// check length of labels
					err := helpers.CheckLabelLength(mariaDBConsumer.ObjectMeta.Labels)
					if err != nil {
						return nil, err
					}
					dbaasTemplates.MariaDB = append(dbaasTemplates.MariaDB, mariaDBConsumer)
				}
			case "mongodb-dbaas":
				{
					mongodbConsumer := &mongodbv1.MongoDBConsumer{
						TypeMeta: metav1.TypeMeta{
							Kind:       "MongoDBConsumer",
							APIVersion: mongodbv1.GroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: serviceValues.Name,
						},
						Spec: mongodbv1.MongoDBConsumerSpec{
							Environment: serviceValues.DBaaSEnvironment,
						},
					}
					mongodbConsumer.ObjectMeta.Labels = map[string]string{}
					mongodbConsumer.ObjectMeta.Annotations = map[string]string{}
					for key, value := range labels {
						mongodbConsumer.ObjectMeta.Labels[key] = value
					}
					for key, value := range annotations {
						mongodbConsumer.ObjectMeta.Annotations[key] = value
					}
					for key, value := range additionalLabels {
						mongodbConsumer.ObjectMeta.Labels[key] = value
					}
					for key, value := range additionalAnnotations {
						mongodbConsumer.ObjectMeta.Annotations[key] = value
					}
					// validate any annotations
					if err := apivalidation.ValidateAnnotations(mongodbConsumer.ObjectMeta.Annotations, nil); err != nil {
						if len(err) != 0 {
							return nil, fmt.Errorf("the annotations for %s are not valid: %v", serviceValues.Name, err)
						}
					}
					// validate any labels
					if err := metavalidation.ValidateLabels(mongodbConsumer.ObjectMeta.Labels, nil); err != nil {
						if len(err) != 0 {
							return nil, fmt.Errorf("the labels for %s are not valid: %v", serviceValues.Name, err)
						}
					}
					// check length of labels
					err := helpers.CheckLabelLength(mongodbConsumer.ObjectMeta.Labels)
					if err != nil {
						return nil, err
					}
					dbaasTemplates.MongoDB = append(dbaasTemplates.MongoDB, *mongodbConsumer)
				}
			case "postgres-dbaas":
				{
					postgresqlConsumer := &postgresv1.PostgreSQLConsumer{
						TypeMeta: metav1.TypeMeta{
							Kind:       "PostgreSQLConsumer",
							APIVersion: postgresv1.GroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: serviceValues.Name,
						},
						Spec: postgresv1.PostgreSQLConsumerSpec{
							Environment: serviceValues.DBaaSEnvironment,
						},
					}
					postgresqlConsumer.ObjectMeta.Labels = map[string]string{}
					postgresqlConsumer.ObjectMeta.Annotations = map[string]string{}
					for key, value := range labels {
						postgresqlConsumer.ObjectMeta.Labels[key] = value
					}
					for key, value := range annotations {
						postgresqlConsumer.ObjectMeta.Annotations[key] = value
					}
					for key, value := range additionalLabels {
						postgresqlConsumer.ObjectMeta.Labels[key] = value
					}
					for key, value := range additionalAnnotations {
						postgresqlConsumer.ObjectMeta.Annotations[key] = value
					}
					// validate any annotations
					if err := apivalidation.ValidateAnnotations(postgresqlConsumer.ObjectMeta.Annotations, nil); err != nil {
						if len(err) != 0 {
							return nil, fmt.Errorf("the annotations for %s are not valid: %v", serviceValues.Name, err)
						}
					}
					// validate any labels
					if err := metavalidation.ValidateLabels(postgresqlConsumer.ObjectMeta.Labels, nil); err != nil {
						if len(err) != 0 {
							return nil, fmt.Errorf("the labels for %s are not valid: %v", serviceValues.Name, err)
						}
					}

					// check length of labels
					err := helpers.CheckLabelLength(postgresqlConsumer.ObjectMeta.Labels)
					if err != nil {
						return nil, err
					}
					dbaasTemplates.PostgreSQL = append(dbaasTemplates.PostgreSQL, *postgresqlConsumer)
				}
			}
		}
	}
	return &dbaasTemplates, nil
}

func TemplateConsumers(dbaas *DBaaSTemplates) ([]byte, error) {
	separator := []byte("---\n")
	var templateYAML []byte
	for _, db := range dbaas.MariaDB {
		dbBytes, err := yaml.Marshal(db)
		if err != nil {
			return nil, fmt.Errorf("couldn't generate template: %v", err)
		}
		restoreResult := append(separator[:], dbBytes[:]...)
		templateYAML = append(templateYAML, restoreResult[:]...)
	}
	for _, db := range dbaas.MongoDB {
		dbBytes, err := yaml.Marshal(db)
		if err != nil {
			return nil, fmt.Errorf("couldn't generate template: %v", err)
		}
		restoreResult := append(separator[:], dbBytes[:]...)
		templateYAML = append(templateYAML, restoreResult[:]...)
	}
	for _, db := range dbaas.PostgreSQL {
		dbBytes, err := yaml.Marshal(db)
		if err != nil {
			return nil, fmt.Errorf("couldn't generate template: %v", err)
		}
		restoreResult := append(separator[:], dbBytes[:]...)
		templateYAML = append(templateYAML, restoreResult[:]...)
	}
	return templateYAML, nil
}
