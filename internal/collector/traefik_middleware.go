package collector

import (
	"context"

	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Collector) CollectTraefikMiddleware(ctx context.Context, namespace string) (*traefik.MiddlewareList, error) {
	// grab all middleware with lagoon.sh/project label, since middleware are not specific to a service
	labelRequirements1, _ := labels.NewRequirement("lagoon.sh/project", selection.Exists, nil)
	listOption := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{
			Selector: labels.NewSelector().Add(*labelRequirements1),
		},
	})
	list := &traefik.MiddlewareList{}
	err := c.Client.List(ctx, list, listOption)
	if apierrors.IsForbidden(err) {
		return nil, err
	}
	return list, nil
}
