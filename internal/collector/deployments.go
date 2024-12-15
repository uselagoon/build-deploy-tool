package collector

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Collector) CollectDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	labelRequirements1, _ := labels.NewRequirement("lagoon.sh/service", selection.Exists, nil)
	listOption := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{
			Selector: labels.NewSelector().Add(*labelRequirements1),
		},
	})
	list := &appsv1.DeploymentList{}
	err := c.Client.List(ctx, list, listOption)
	if err != nil {
		return nil, err
	}
	return list, nil
}
