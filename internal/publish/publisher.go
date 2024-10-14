package publish

import (
	"context"
	"fmt"
	"io"

	"github.com/registry-tools/publish/internal/module"

	sdk "github.com/registry-tools/rt-sdk"
	"github.com/registry-tools/rt-sdk/generated/api"
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

	moduleBody := api.NewTerraformModuleVersionsPostRequestBody_module()
	moduleBody.SetName(&info.Name)
	moduleBody.SetNamespace(&info.Namespace)
	moduleBody.SetSystem(&info.System)
	moduleBody.SetVersion(&info.Version)

	moduleMeta := api.NewTerraformModuleVersionsPostRequestBody_meta()
	moduleMeta.SetArchiveId(signedID)

	moduleData := api.NewTerraformModuleVersionsPostRequestBody()
	moduleData.SetModule(moduleBody)
	moduleData.SetMeta(moduleMeta)

	response, err := p.SDK.Api().TerraformModuleVersions().PostAsTerraformModuleVersionsPostResponse(ctx, moduleData, nil)
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
