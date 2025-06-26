package collector

import (
	"context"
	"fmt"
	"os"
	"strings"

	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	k8upv1alpha1 "github.com/vshn/k8up/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"

	mariadbv1 "github.com/amazeeio/dbaas-operator/apis/mariadb/v1"
	mongodbv1 "github.com/amazeeio/dbaas-operator/apis/mongodb/v1"
	postgresv1 "github.com/amazeeio/dbaas-operator/apis/postgres/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type Collector struct {
	Client client.Client
}

type LagoonEnvState struct {
	Deployments           *appsv1.DeploymentList             `json:"deployments"`
	Cronjobs              *batchv1.CronJobList               `json:"cronjobs"`
	Ingress               *networkv1.IngressList             `json:"ingress"`
	Services              *corev1.ServiceList                `json:"services"`
	Secrets               *corev1.SecretList                 `json:"secrets"`
	PVCs                  *corev1.PersistentVolumeClaimList  `json:"pvcs"`
	SchedulesV1           *k8upv1.ScheduleList               `json:"schedulesv1"`
	SchedulesV1Alpha1     *k8upv1alpha1.ScheduleList         `json:"schedulesv1alpha1"`
	PreBackupPodsV1       *k8upv1.PreBackupPodList           `json:"prebackuppodsv1"`
	PreBackupPodsV1Alpha1 *k8upv1alpha1.PreBackupPodList     `json:"prebackuppodsv1alpha1"`
	MariaDBConsumers      *mariadbv1.MariaDBConsumerList     `json:"mariadbconsumers"`
	MongoDBConsumers      *mongodbv1.MongoDBConsumerList     `json:"mongodbconsumers"`
	PostgreSQLConsumers   *postgresv1.PostgreSQLConsumerList `json:"postgresqlconsumers"`
	NetworkPolicies       *networkv1.NetworkPolicyList       `json:"networkpolicies"`
}

func NewCollector(client client.Client) *Collector {
	return &Collector{
		Client: client,
	}
}

func (c *Collector) Collect(ctx context.Context, namespace string) (*LagoonEnvState, error) {
	var state LagoonEnvState
	var err error
	state.Deployments, err = c.CollectDeployments(ctx, namespace)
	if err != nil {
		return nil, err
	}
	state.Cronjobs, err = c.CollectCronjobs(ctx, namespace)
	if err != nil {
		return nil, err
	}
	state.Ingress, err = c.CollectIngress(ctx, namespace)
	if err != nil {
		return nil, err
	}
	state.Services, err = c.CollectServices(ctx, namespace)
	if err != nil {
		return nil, err
	}
	state.Secrets, err = c.CollectSecrets(ctx, namespace)
	if err != nil {
		return nil, err
	}
	state.PVCs, err = c.CollectPVCs(ctx, namespace)
	if err != nil {
		return nil, err
	}
	state.MariaDBConsumers, err = c.CollectMariaDBConsumers(ctx, namespace)
	if err != nil {
		// handle if consumer crds not installed
		if !strings.Contains(err.Error(), "no matches for kind") {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}
	}
	state.MongoDBConsumers, err = c.CollectMongoDBConsumers(ctx, namespace)
	if err != nil {
		// handle if consumer crds not installed
		if !strings.Contains(err.Error(), "no matches for kind") {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}
	}
	state.PostgreSQLConsumers, err = c.CollectPostgreSQLConsumers(ctx, namespace)
	if err != nil {
		// handle if consumer crds not installed
		if !strings.Contains(err.Error(), "no matches for kind") {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}
	}
	state.SchedulesV1, err = c.CollectSchedulesV1(ctx, namespace)
	if err != nil {
		// handle if k8up v1 crds not installed
		if !strings.Contains(err.Error(), "no matches for kind") {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}
	}
	state.SchedulesV1Alpha1, err = c.CollectSchedulesV1Alpha1(ctx, namespace)
	if err != nil {
		// handle if k8up v1alpha1 crds not installed
		if !strings.Contains(err.Error(), "no matches for kind") {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}
	}
	state.PreBackupPodsV1, err = c.CollectPreBackupPodsV1(ctx, namespace)
	if err != nil {
		// handle if k8up v1 crds not installed
		if !strings.Contains(err.Error(), "no matches for kind") {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}
	}
	state.PreBackupPodsV1Alpha1, err = c.CollectPreBackupPodsV1Alpha1(ctx, namespace)
	if err != nil {
		// handle if k8up v1alpha1 crds not installed
		if !strings.Contains(err.Error(), "no matches for kind") {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}
	}
	state.NetworkPolicies, err = c.CollectNetworkPolicies(ctx, namespace)
	if err != nil {
		return nil, err
	}
	return &state, nil
}
