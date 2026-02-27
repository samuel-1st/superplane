package cloudsmith

import (
	"fmt"

	"github.com/superplanehq/superplane/pkg/core"
)

const (
	ResourceTypeRepository = "cloudsmith.repository"
)

func listCloudsmithResources(resourceType string, ctx core.ListResourcesContext) ([]core.IntegrationResource, error) {
	switch resourceType {
	case ResourceTypeRepository:
		return listCloudsmithRepositories(ctx)

	default:
		return []core.IntegrationResource{}, nil
	}
}

func listCloudsmithRepositories(ctx core.ListResourcesContext) ([]core.IntegrationResource, error) {
	namespace, err := ctx.Integration.GetConfig("namespace")
	if err != nil {
		return nil, fmt.Errorf("integration namespace is required: %w", err)
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return nil, err
	}

	repositories, err := client.ListRepositories(string(namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	resources := make([]core.IntegrationResource, 0, len(repositories))
	for _, repository := range repositories {
		value := string(namespace) + "/" + repository.Slug
		resources = append(resources, core.IntegrationResource{
			Type: ResourceTypeRepository,
			Name: value,
			ID:   value,
		})
	}

	return resources, nil
}
