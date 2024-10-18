package publish

import (
	"context"
	"fmt"
	"io"

	"github.com/registry-tools/rt-cli/internal/module"

	sdk "github.com/registry-tools/rt-sdk"
	"github.com/registry-tools/rt-sdk/generated/models"
)

// Publisher publishes modules to the registry.
type Publisher struct {
	SDK sdk.SDK
}

type ModuleVersion struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	System    string `json:"system"`
	Version   string `json:"version"`
	Namespace string `json:"namespace"`
}

func (v ModuleVersion) Module(namespace string) module.Module {
	return module.Module{
		Namespace: namespace,
		Name:      v.Name,
		System:    v.System,
		Version:   v.Version,
	}
}

// Publish publishes the specified module using the specified archive file.
func (p Publisher) Publish(ctx context.Context, info module.Module, reader io.ReadSeeker) (*ModuleVersion, error) {
	signedID, err := p.SDK.UploadFileArchive(ctx, fmt.Sprintf("%s-%s-%s", info.Name, info.System, info.Version), reader)
	if err != nil {
		return nil, err
	}

	moduleBody := models.NewTerraformModuleVersion()
	moduleBody.SetName(&info.Name)
	moduleBody.SetNamespace(&info.Namespace)
	moduleBody.SetSystem(&info.System)
	moduleBody.SetVersion(&info.Version)
	moduleBody.SetArchiveId(signedID)

	response, err := p.SDK.Api().TerraformModuleVersions().PostAsTerraformModuleVersionsPostResponse(ctx, moduleBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create terraform module: %w", err)
	}

	data := response.GetData()

	return &ModuleVersion{
		ID:        *data.GetId(),
		Name:      *data.GetName(),
		System:    *data.GetSystem(),
		Version:   *data.GetVersion(),
		Namespace: *data.GetNamespace(),
	}, nil
}
