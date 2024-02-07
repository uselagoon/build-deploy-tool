package backups

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"

	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	k8upv1alpha1 "github.com/vshn/k8up/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"

	"sigs.k8s.io/yaml"
)

type PreBackupPodTmpl struct {
	Service   generator.ServiceValues
	Namespace string
}

func GeneratePreBackupPod(
	lValues generator.BuildValues,
) ([]byte, error) {
	// generate the template spec

	var result []byte
	separator := []byte("---\n")

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

	// create the prebackuppods
	for _, serviceValues := range lValues.Services {
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
		additionalLabels["app.kubernetes.io/name"] = serviceValues.Type
		additionalLabels["app.kubernetes.io/instance"] = serviceValues.OverrideName
		additionalLabels["lagoon.sh/service"] = serviceValues.OverrideName
		additionalLabels["lagoon.sh/service-type"] = serviceValues.Type
		if _, ok := preBackupPodSpecs[serviceValues.Type]; ok {
			switch lValues.Backup.K8upVersion {
			case "v1":
				prebackuppod := &k8upv1alpha1.PreBackupPod{
					TypeMeta: metav1.TypeMeta{
						Kind:       "PreBackupPod",
						APIVersion: k8upv1alpha1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: fmt.Sprintf("%s-prebackuppod", serviceValues.OverrideName),
					},
					Spec: k8upv1alpha1.PreBackupPodSpec{},
				}

				prebackuppod.ObjectMeta.Labels = labels
				prebackuppod.ObjectMeta.Annotations = annotations
				prebackuppod.ObjectMeta.Labels["prebackuppod"] = serviceValues.OverrideName

				var pbp bytes.Buffer
				tmpl, _ := template.New("").Funcs(funcMap).Parse(preBackupPodSpecs[serviceValues.Type])
				tmplVals := PreBackupPodTmpl{
					Service:   serviceValues,
					Namespace: lValues.Namespace,
				}
				err := tmpl.Execute(&pbp, tmplVals)
				if err != nil {
					return nil, err
				}
				k8upPBPSpec := k8upv1alpha1.PreBackupPodSpec{}
				err = yaml.Unmarshal(pbp.Bytes(), &k8upPBPSpec)
				if err != nil {
					return nil, err
				}

				prebackuppod.Spec = k8upPBPSpec

				if lValues.ImageCache != "" {
					imageCachedImage := fmt.Sprintf("%s%s", lValues.ImageCache, prebackuppod.Spec.Pod.Spec.Containers[0].Image)
					prebackuppod.Spec.Pod.Spec.Containers[0].Image = imageCachedImage
				}

				if prebackuppod.Spec.Pod.Spec.Containers[0].EnvFrom == nil && serviceValues.DBaasReadReplica {
					prebackuppod.Spec.Pod.Spec.Containers[0].Env = append(prebackuppod.Spec.Pod.Spec.Containers[0].Env, v1.EnvVar{
						Name: "BACKUP_DB_READREPLICA_HOSTS",
						ValueFrom: &v1.EnvVarSource{
							ConfigMapKeyRef: &v1.ConfigMapKeySelector{
								Key: fmt.Sprintf("%s_READREPLICA_HOSTS", varFix(serviceValues.OverrideName)),
								LocalObjectReference: v1.LocalObjectReference{
									Name: "lagoon-env",
								},
							},
						},
					})
				}

				for key, value := range additionalLabels {
					prebackuppod.ObjectMeta.Labels[key] = value
				}
				// add any additional annotations
				for key, value := range additionalAnnotations {
					prebackuppod.ObjectMeta.Annotations[key] = value
				}
				// validate any annotations
				if err := apivalidation.ValidateAnnotations(prebackuppod.ObjectMeta.Annotations, nil); err != nil {
					if len(err) != 0 {
						return nil, fmt.Errorf("the annotations for %s/%s are not valid: %v", "prebackuppod", serviceValues.Name, err)
					}
				}
				// validate any labels
				if err := metavalidation.ValidateLabels(prebackuppod.ObjectMeta.Labels, nil); err != nil {
					if len(err) != 0 {
						return nil, fmt.Errorf("the labels for %s/%s are not valid: %v", "prebackuppod", serviceValues.Name, err)
					}
				}

				// check length of labels
				err = helpers.CheckLabelLength(prebackuppod.ObjectMeta.Labels)
				if err != nil {
					return nil, err
				}
				// @TODO: we should review this in the future when we stop doing `kubectl apply` in the builds :)
				// marshal the resulting ingress
				prebackuppodBytes, err := yaml.Marshal(prebackuppod)
				if err != nil {
					return nil, err
				}

				pbpBytes, _ := RemoveYAML(prebackuppodBytes)
				// add the seperator to the template so that it can be `kubectl apply` in bulk as part
				// of the current build process
				restoreResult := append(separator[:], pbpBytes[:]...)
				result = append(result, restoreResult[:]...)
			case "v2":
				prebackuppod := &k8upv1.PreBackupPod{
					TypeMeta: metav1.TypeMeta{
						Kind:       "PreBackupPod",
						APIVersion: k8upv1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: fmt.Sprintf("%s-prebackuppod", serviceValues.Name),
					},
					Spec: k8upv1.PreBackupPodSpec{},
				}

				prebackuppod.ObjectMeta.Labels = labels
				prebackuppod.ObjectMeta.Annotations = annotations
				prebackuppod.ObjectMeta.Labels["prebackuppod"] = serviceValues.Name

				var pbp bytes.Buffer
				tmpl, _ := template.New("").Funcs(funcMap).Parse(preBackupPodSpecs[serviceValues.Type])
				tmplVals := PreBackupPodTmpl{
					Service:   serviceValues,
					Namespace: lValues.Namespace,
				}
				err := tmpl.Execute(&pbp, tmplVals)
				if err != nil {
					return nil, err
				}
				k8upPBPSpec := k8upv1.PreBackupPodSpec{}
				err = yaml.Unmarshal(pbp.Bytes(), &k8upPBPSpec)
				if err != nil {
					return nil, err
				}

				prebackuppod.Spec = k8upPBPSpec

				if lValues.ImageCache != "" {
					imageCachedImage := fmt.Sprintf("%s%s", lValues.ImageCache, prebackuppod.Spec.Pod.Spec.Containers[0].Image)
					prebackuppod.Spec.Pod.Spec.Containers[0].Image = imageCachedImage
				}

				if prebackuppod.Spec.Pod.Spec.Containers[0].EnvFrom == nil && serviceValues.DBaasReadReplica {
					prebackuppod.Spec.Pod.Spec.Containers[0].Env = append(prebackuppod.Spec.Pod.Spec.Containers[0].Env, v1.EnvVar{
						Name: "BACKUP_DB_READREPLICA_HOSTS",
						ValueFrom: &v1.EnvVarSource{
							ConfigMapKeyRef: &v1.ConfigMapKeySelector{
								Key: fmt.Sprintf("%s_READREPLICA_HOSTS", varFix(serviceValues.OverrideName)),
								LocalObjectReference: v1.LocalObjectReference{
									Name: "lagoon-env",
								},
							},
						},
					})
				}

				for key, value := range additionalLabels {
					prebackuppod.ObjectMeta.Labels[key] = value
				}
				// add any additional annotations
				for key, value := range additionalAnnotations {
					prebackuppod.ObjectMeta.Annotations[key] = value
				}
				// validate any annotations
				if err := apivalidation.ValidateAnnotations(prebackuppod.ObjectMeta.Annotations, nil); err != nil {
					if len(err) != 0 {
						return nil, fmt.Errorf("the annotations for %s/%s are not valid: %v", "prebackuppod", serviceValues.Name, err)
					}
				}
				// validate any labels
				if err := metavalidation.ValidateLabels(prebackuppod.ObjectMeta.Labels, nil); err != nil {
					if len(err) != 0 {
						return nil, fmt.Errorf("the labels for %s/%s are not valid: %v", "prebackuppod", serviceValues.Name, err)
					}
				}

				// check length of labels
				err = helpers.CheckLabelLength(prebackuppod.ObjectMeta.Labels)
				if err != nil {
					return nil, err
				}
				// @TODO: we should review this in the future when we stop doing `kubectl apply` in the builds :)
				// marshal the resulting ingress
				prebackuppodBytes, err := yaml.Marshal(prebackuppod)
				if err != nil {
					return nil, err
				}
				pbpBytes, _ := RemoveYAML(prebackuppodBytes)
				// add the seperator to the template so that it can be `kubectl apply` in bulk as part
				// of the current build process
				restoreResult := append(separator[:], pbpBytes[:]...)
				result = append(result, restoreResult[:]...)
			}
		}
	}
	return result, nil
}

// helper function to remove the creationtimestamp from the prebackuppod pod spec so that kubectl will apply without validation errors
func RemoveYAML(a []byte) ([]byte, error) {
	tmpMap := map[string]interface{}{}
	yaml.Unmarshal(a, &tmpMap)
	if _, ok := tmpMap["spec"].(map[string]interface{})["pod"].(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"]; ok {
		delete(tmpMap["spec"].(map[string]interface{})["pod"].(map[string]interface{})["metadata"].(map[string]interface{}), "creationTimestamp")
		b, _ := yaml.Marshal(tmpMap)
		return b, nil
	}
	return a, nil
}

var funcMap = template.FuncMap{
	"VarFix": varFix,
}

// varfix just uppercases and replaces - with _ for variable names
func varFix(s string) string {
	return fmt.Sprintf("%s", strings.ToUpper(strings.Replace(s, "-", "_", -1)))
}

// this is just the first run at doing this, once the service template generator is introduced, this will need to be re-evaluated
type PreBackupPods map[string]string

// this is just the first run at doing this, once the service template generator is introduced, this will need to be re-evaluated
var preBackupPodSpecs = PreBackupPods{
	"mariadb-dbaas": `backupCommand: >
  /bin/sh -c "if [ ! -z $BACKUP_DB_READREPLICA_HOSTS ]; then
  BACKUP_DB_HOST=$(echo $BACKUP_DB_READREPLICA_HOSTS | cut -d ',' -f1);
  fi &&
  dump=$(mktemp)
  && mysqldump --max-allowed-packet=500M --events --routines --quick
  --add-locks --no-autocommit --single-transaction --no-create-db
  --no-data --no-tablespaces
  -h $BACKUP_DB_HOST
  -u $BACKUP_DB_USERNAME
  -p$BACKUP_DB_PASSWORD
  $BACKUP_DB_DATABASE
  > $dump
  && mysqldump --max-allowed-packet=500M --events --routines --quick
  --add-locks --no-autocommit --single-transaction --no-create-db
  --ignore-table=$BACKUP_DB_DATABASE.watchdog
  --no-create-info --no-tablespaces --skip-triggers
  -h $BACKUP_DB_HOST
  -u $BACKUP_DB_USERNAME
  -p$BACKUP_DB_PASSWORD
  $BACKUP_DB_DATABASE
  >> $dump
  && cat $dump && rm $dump"
fileExtension: .{{ .Service.Name }}.sql
pod:
  spec:
    containers:
    - args:
      - sleep
      - infinity
      env:
      - name: BACKUP_DB_HOST
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_HOST
            name: lagoon-env
      - name: BACKUP_DB_USERNAME
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_USERNAME
            name: lagoon-env
      - name: BACKUP_DB_PASSWORD
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_PASSWORD
            name: lagoon-env
      - name: BACKUP_DB_DATABASE
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_DATABASE
            name: lagoon-env
      image: uselagoon/database-tools:latest
      imagePullPolicy: Always
      name: {{ .Service.Name }}-prebackuppod`,
	"postgres-dbaas": `backupCommand: >
  /bin/sh -c  "if [ ! -z $BACKUP_DB_READREPLICA_HOSTS ]; then
  BACKUP_DB_HOST=$(echo $BACKUP_DB_READREPLICA_HOSTS | cut -d ',' -f1);
  fi && PGPASSWORD=$BACKUP_DB_PASSWORD pg_dump
  --host=$BACKUP_DB_HOST
  --port=$BACKUP_DB_PORT
  --dbname=$BACKUP_DB_DATABASE
  --username=$BACKUP_DB_USERNAME
  --format=t -w"
fileExtension: .{{ .Service.Name }}.tar
pod:
  spec:
    containers:
    - args:
      - sleep
      - infinity
      env:
      - name: BACKUP_DB_HOST
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_HOST
            name: lagoon-env
      - name: BACKUP_DB_USERNAME
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_USERNAME
            name: lagoon-env
      - name: BACKUP_DB_PASSWORD
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_PASSWORD
            name: lagoon-env
      - name: BACKUP_DB_DATABASE
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_DATABASE
            name: lagoon-env
      image: uselagoon/database-tools:latest
      imagePullPolicy: Always
      name: {{ .Service.Name }}-prebackuppod`,
	"mongodb-dbaas": `backupCommand: /bin/sh -c "dump=$(mktemp) && mongodump --quiet --ssl --tlsInsecure --username=${BACKUP_DB_USERNAME} --password=${BACKUP_DB_PASSWORD} --host=${BACKUP_DB_HOST}:${BACKUP_DB_PORT} --db=${BACKUP_DB_DATABASE} --authenticationDatabase=${BACKUP_DB_AUTHSOURCE} --authenticationMechanism=${BACKUP_DB_AUTHMECHANISM} --archive=$dump && cat $dump && rm $dump"
fileExtension: .{{ .Service.Name }}.bson
pod:
  spec:
    containers:
    - args:
      - sleep
      - infinity
      env:
      - name: BACKUP_DB_HOST
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_HOST
            name: lagoon-env
      - name: BACKUP_DB_USERNAME
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_USERNAME
            name: lagoon-env
      - name: BACKUP_DB_PASSWORD
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_PASSWORD
            name: lagoon-env
      - name: BACKUP_DB_DATABASE
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_DATABASE
            name: lagoon-env
      - name: BACKUP_DB_PORT
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_PORT
            name: lagoon-env
      - name: BACKUP_DB_AUTHSOURCE
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_AUTHSOURCE
            name: lagoon-env
      - name: BACKUP_DB_AUTHMECHANISM
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_AUTHMECHANISM
            name: lagoon-env
      - name: BACKUP_DB_AUTHTLS
        valueFrom:
          configMapKeyRef:
            key: {{ .Service.Name | VarFix }}_AUTHTLS
            name: lagoon-env
      image: uselagoon/database-tools:latest
      imagePullPolicy: Always
      name: {{ .Service.Name }}-prebackuppod`,
}
