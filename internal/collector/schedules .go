package collector

import (
	"context"

	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	k8upv1alpha1 "github.com/vshn/k8up/api/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Collector) CollectSchedulesV1(ctx context.Context, namespace string) (*k8upv1.ScheduleList, error) {
	labelRequirements1, _ := labels.NewRequirement("lagoon.sh/service", selection.Exists, nil)
	listOption := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{
			Selector: labels.NewSelector().Add(*labelRequirements1),
		},
	})
	list := &k8upv1.ScheduleList{}
	err := c.Client.List(ctx, list, listOption)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (c *Collector) CollectSchedulesV1Alpha1(ctx context.Context, namespace string) (*k8upv1alpha1.ScheduleList, error) {
	labelRequirements1, _ := labels.NewRequirement("lagoon.sh/service", selection.Exists, nil)
	listOption := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{
			Selector: labels.NewSelector().Add(*labelRequirements1),
		},
	})
	list := &k8upv1alpha1.ScheduleList{}
	err := c.Client.List(ctx, list, listOption)
	if err != nil {
		return nil, err
	}
	return list, nil
}
