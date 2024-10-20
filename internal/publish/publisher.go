package publish

import (
	"context"
	"fmt"
	"io"
	"strings"

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
		if strings.Contains(err.Error(), "Not found") {
			return nil, fmt.Errorf("authentication failed, check your credentials or re-run 'rt login'")
		}
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
		if sdk.IsNotFoundError(err) {
			return nil, fmt.Errorf("namespace does not exist or you do not have permission to publish to it")
		}
		return nil, sdk.FormatAPIError(err)
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
