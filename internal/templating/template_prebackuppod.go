package templating

import (
	"fmt"
	"strings"

	"github.com/uselagoon/build-deploy-tool/internal/generator"

	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	k8upv1alpha1 "github.com/vshn/k8up/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// Enum of allowed service backup types
const (
	MariaDB  string = "mariadb-dbaas"
	Postgres string = "postgres-dbaas"
	MongoDB  string = "mongodb-dbaas"
)

func isUnsupported(serviceType string) bool {
	switch serviceType {
	case MariaDB, Postgres, MongoDB:
		return false
	}
	return true
}

// Serialize a list of prebackup pods into a YAML bytestream.
func TemplatePreBackupPods(pods []k8upv1.PreBackupPod) ([]byte, error) {
	separator := []byte("---\n")
	var templateYAML []byte
	for _, pod := range pods {
		podBytes, err := yaml.Marshal(pod)
		if err != nil {
			return nil, fmt.Errorf("couldn't generate template: %v", err)
		}
		podBytes, _ = RemoveCreationTimestamp(podBytes)
		restoreResult := append(separator[:], podBytes[:]...)
		templateYAML = append(templateYAML, restoreResult[:]...)
	}
	return templateYAML, nil
}

// Build a list of PreBackup Pods
func GeneratePreBackupPod(buildValues generator.BuildValues) ([]k8upv1.PreBackupPod, error) {
	var pods []k8upv1.PreBackupPod

	defaultLabels := map[string]string{
		"app.kubernetes.io/managed-by": "build-deploy-tool",
		"lagoon.sh/project":            buildValues.Project,
		"lagoon.sh/environment":        buildValues.Environment,
		"lagoon.sh/environmentType":    buildValues.EnvironmentType,
		"lagoon.sh/buildType":          buildValues.BuildType,
	}
	defaultAnnotations := map[string]string{
		"lagoon.sh/version": buildValues.LagoonVersion,
	}

	for _, serviceValues := range buildValues.Services {
		var pod k8upv1.PreBackupPod

		if isUnsupported(serviceValues.Type) {
			continue
		}

		labels := make(map[string]string, len(defaultLabels))
		for k, v := range defaultLabels {
			labels[k] = v
		}
		labels["app.kubernetes.io/name"] = serviceValues.Type
		labels["app.kubernetes.io/instance"] = serviceValues.Name
		labels["lagoon.sh/service"] = serviceValues.Name
		labels["lagoon.sh/service-type"] = serviceValues.Type
		labels["prebackuppod"] = serviceValues.Name

		annotations := make(map[string]string, len(defaultAnnotations))
		for k, v := range defaultAnnotations {
			annotations[k] = v
		}
		if buildValues.BuildType == "branch" {
			annotations["lagoon.sh/branch"] = buildValues.Branch
		}
		if buildValues.BuildType == "pullrequest" {
			annotations["lagoon.sh/prNumber"] = buildValues.PRNumber
			annotations["lagoon.sh/prHeadBranch"] = buildValues.PRHeadBranch
			annotations["lagoon.sh/prBaseBranch"] = buildValues.PRBaseBranch
		}

		version := buildValues.Backup.K8upVersion
		if version == "v1" {
			pod.TypeMeta = metav1.TypeMeta{
				Kind:       "PreBackupPod",
				APIVersion: k8upv1alpha1.GroupVersion.String(),
			}
		} else if version == "v2" {
			pod.TypeMeta = metav1.TypeMeta{
				Kind:       "PreBackupPod",
				APIVersion: k8upv1.GroupVersion.String(),
			}
		} else {
			return nil, fmt.Errorf("invalid K8up version: %s", version)
		}

		pod.ObjectMeta = metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-prebackuppod", serviceValues.Name),
			Labels:      labels,
			Annotations: annotations,
		}

		backupCommand, err := getBackupCommand(serviceValues.Type)
		if err != nil {
			return nil, err
		}
		fileExtension, err := getFileExtension(serviceValues)
		if err != nil {
			return nil, err
		}
		podSpecs, err := getPodSpecs(serviceValues)
		if err != nil {
			return nil, err
		}
		if buildValues.ImageCache != "" {
			imageCachedImage := fmt.Sprintf("%s%s", buildValues.ImageCache, podSpecs.Spec.Containers[0].Image)
			podSpecs.Spec.Containers[0].Image = imageCachedImage
		}

		pod.Spec = k8upv1.PreBackupPodSpec{
			BackupCommand: backupCommand,
			FileExtension: fileExtension,
			Pod:           podSpecs,
		}

		pods = append(pods, pod)
	}

	return pods, nil
}

func getBackupCommand(serviceType string) (string, error) {
	var cmd string
	switch serviceType {
	case MariaDB:
		cmd = mariadbBackupCommand
	case Postgres:
		cmd = postgresBackupCommand
	case MongoDB:
		cmd = mongoBackupCommand
	default:
		return "", fmt.Errorf("unknown service type %s passed to backup command getter", serviceType)
	}

	return cmd, nil
}

func getFileExtension(serviceValues generator.ServiceValues) (string, error) {
	var extension string
	switch serviceValues.Type {
	case MariaDB:
		extension = "sql"
	case Postgres:
		extension = "tar"
	case MongoDB:
		extension = "bson"
	default:
		return "", fmt.Errorf("unknown service type %s passed to file extension getter", serviceValues.Type)
	}

	fileExtension := "." + serviceValues.Name + "." + extension
	return fileExtension, nil
}

func getPodSpecs(serviceValues generator.ServiceValues) (*k8upv1.Pod, error) {
	var pod k8upv1.Pod

	env, err := getEnvVars(serviceValues)
	if err != nil {
		return nil, err
	}

	pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
		Args:            []string{"sleep", "infinity"},
		Image:           "uselagoon/database-tools:latest",
		ImagePullPolicy: corev1.PullAlways,
		Name:            serviceValues.Name + "-prebackuppod",
		Env:             env,
	})

	return &pod, nil
}

func getEnvVars(serviceValues generator.ServiceValues) ([]corev1.EnvVar, error) {
	// default env vars present in all db types
	env := []corev1.EnvVar{
		{
			Name: "BACKUP_DB_HOST",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "lagoon-env",
					},
					Key: varFix(serviceValues.Name) + "_HOST",
				},
			},
		},
		{
			Name: "BACKUP_DB_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "lagoon-env",
					},
					Key: varFix(serviceValues.Name) + "_USERNAME",
				},
			},
		},
		{
			Name: "BACKUP_DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "lagoon-env",
					},
					Key: varFix(serviceValues.Name) + "_PASSWORD",
				},
			},
		},
		{
			Name: "BACKUP_DB_DATABASE",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "lagoon-env",
					},
					Key: varFix(serviceValues.Name) + "_DATABASE",
				},
			},
		},
	}

	if serviceValues.Type == MongoDB {
		env = append(env, []corev1.EnvVar{
			{
				Name: "BACKUP_DB_PORT",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "lagoon-env",
						},
						Key: varFix(serviceValues.Name) + "_PORT",
					},
				},
			},
			{
				Name: "BACKUP_DB_AUTHSOURCE",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "lagoon-env",
						},
						Key: varFix(serviceValues.Name) + "_AUTHSOURCE",
					},
				},
			},
			{
				Name: "BACKUP_DB_AUTHMECHANISM",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "lagoon-env",
						},
						Key: varFix(serviceValues.Name) + "_AUTHMECHANISM",
					},
				},
			},
			{
				Name: "BACKUP_DB_AUTHTLS",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "lagoon-env",
						},
						Key: varFix(serviceValues.Name) + "_AUTHTLS",
					},
				},
			},
		}...)
	}

	if serviceValues.DBaasReadReplica {
		env = append(env, corev1.EnvVar{
			Name: "BACKUP_DB_READREPLICA_HOSTS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: fmt.Sprintf("%s_READREPLICA_HOSTS", varFix(serviceValues.OverrideName)),
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "lagoon-env",
					},
				},
			},
		})
	}

	return env, nil
}

// Removes "creationTimestamp: null" from 'spec.pod.metadata'
func RemoveCreationTimestamp(a []byte) ([]byte, error) {
	tmpMap := map[string]interface{}{}
	if err := yaml.Unmarshal(a, &tmpMap); err != nil {
		return nil, err
	}

	spec, ok := tmpMap["spec"].(map[string]interface{})
	if !ok {
		return a, nil
	}
	pod, ok := spec["pod"].(map[string]interface{})
	if !ok {
		return a, nil
	}
	meta, ok := pod["metadata"].(map[string]interface{})
	if !ok {
		return a, nil
	}

	delete(meta, "creationTimestamp")
	return yaml.Marshal(tmpMap)
}

var mariadbBackupCommand = `/bin/sh -c "if [ ! -z $BACKUP_DB_READREPLICA_HOSTS ]; then
BACKUP_DB_HOST=$(echo $BACKUP_DB_READREPLICA_HOSTS | cut -d ',' -f1);
fi &&
dump=$(mktemp)
&& mysqldump --max-allowed-packet=1G --events --routines --quick
--add-locks --no-autocommit --single-transaction --no-create-db
--no-data --no-tablespaces
-h $BACKUP_DB_HOST
-u $BACKUP_DB_USERNAME
-p$BACKUP_DB_PASSWORD
$BACKUP_DB_DATABASE
> $dump
&& mysqldump --max-allowed-packet=1G --events --routines --quick
--add-locks --no-autocommit --single-transaction --no-create-db
--ignore-table=$BACKUP_DB_DATABASE.watchdog
--no-create-info --no-tablespaces --skip-triggers
-h $BACKUP_DB_HOST
-u $BACKUP_DB_USERNAME
-p$BACKUP_DB_PASSWORD
$BACKUP_DB_DATABASE
>> $dump
&& cat $dump && rm $dump"`

var postgresBackupCommand = `/bin/sh -c "if [ ! -z $BACKUP_DB_READREPLICA_HOSTS ]; then
BACKUP_DB_HOST=$(echo $BACKUP_DB_READREPLICA_HOSTS | cut -d ',' -f1);
fi && PGPASSWORD=$BACKUP_DB_PASSWORD pg_dump
--host=$BACKUP_DB_HOST
--port=$BACKUP_DB_PORT
--dbname=$BACKUP_DB_DATABASE
--username=$BACKUP_DB_USERNAME
--format=t -w"`

var mongoBackupCommand = `/bin/sh -c "dump=$(mktemp) && mongodump \
--quiet \
--ssl \
--tlsInsecure \
--username=${BACKUP_DB_USERNAME} \
--password=${BACKUP_DB_PASSWORD} \
--host=${BACKUP_DB_HOST}:${BACKUP_DB_PORT} \
--db=${BACKUP_DB_DATABASE} \
--authenticationDatabase=${BACKUP_DB_AUTHSOURCE} \
--authenticationMechanism=${BACKUP_DB_AUTHMECHANISM} \
--archive=$dump \
&& cat $dump && rm $dump"`

// varfix just uppercases and replaces - with _ for variable names
func varFix(s string) string {
	return strings.ToUpper(strings.Replace(s, "-", "_", -1))
}
