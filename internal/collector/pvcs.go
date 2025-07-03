package collector

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Collector) CollectPVCs(ctx context.Context, namespace string) (*corev1.PersistentVolumeClaimList, error) {
	labelRequirements1, _ := labels.NewRequirement("lagoon.sh/template", selection.Exists, nil)
	listOption := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{
			Selector: labels.NewSelector().Add(*labelRequirements1),
		},
	})
	list := &corev1.PersistentVolumeClaimList{}
	err := c.Client.List(ctx, list, listOption)
	if err != nil {
		return nil, err
	}
	return list, nil
}
