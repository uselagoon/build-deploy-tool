package collector

import (
	"context"

	mariadbv1 "github.com/amazeeio/dbaas-operator/apis/mariadb/v1"
	mongodbv1 "github.com/amazeeio/dbaas-operator/apis/mongodb/v1"
	postgresv1 "github.com/amazeeio/dbaas-operator/apis/postgres/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Collector) CollectMariaDBConsumers(ctx context.Context, namespace string) (*mariadbv1.MariaDBConsumerList, error) {
	labelRequirements1, _ := labels.NewRequirement("lagoon.sh/service", selection.Exists, nil)
	listOption := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{
			Selector: labels.NewSelector().Add(*labelRequirements1),
		},
	})
	list := &mariadbv1.MariaDBConsumerList{}
	err := c.Client.List(ctx, list, listOption)
	if apierrors.IsForbidden(err) {
		return nil, err
	}
	return list, nil
}

func (c *Collector) CollectMongoDBConsumers(ctx context.Context, namespace string) (*mongodbv1.MongoDBConsumerList, error) {
	labelRequirements1, _ := labels.NewRequirement("lagoon.sh/service", selection.Exists, nil)
	listOption := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{
			Selector: labels.NewSelector().Add(*labelRequirements1),
		},
	})
	list := &mongodbv1.MongoDBConsumerList{}
	err := c.Client.List(ctx, list, listOption)
	if apierrors.IsForbidden(err) {
		return nil, err
	}
	return list, nil
}

func (c *Collector) CollectPostgreSQLConsumers(ctx context.Context, namespace string) (*postgresv1.PostgreSQLConsumerList, error) {
	labelRequirements1, _ := labels.NewRequirement("lagoon.sh/service", selection.Exists, nil)
	listOption := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{
			Selector: labels.NewSelector().Add(*labelRequirements1),
		},
	})
	list := &postgresv1.PostgreSQLConsumerList{}
	err := c.Client.List(ctx, list, listOption)
	if apierrors.IsForbidden(err) {
		return nil, err
	}
	return list, nil
}
