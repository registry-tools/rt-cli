# `rt`, the Registry Tools CLI

Publish modules to a Registry Tools registry, using the CLI or GitHub Action.

## Environment Configuration

`REGISTRY_TOOLS_TOKEN` - required
`REGISTRY_TOOLS_HOSTNAME` - defaults to `registrytools.cloud`
`LOG_LEVEL` - defaults to `WARN`

## GitHub Action Usage

```
uses: registry-tools/publish-action
  env:
    REGISTRY_TOOLS_TOKEN: ${{ secrets.REGISTRY_TOOLS_TOKEN }}
  with:
    version: "1.0.0"
    module: "my-compute-module"
    system: "aws"
    directory: "."
```

## CLI Usage

`rt publish --namespace=platform --version=2.5.0 --name=test --system=null --directory .`

Defaults can also be extracted from the directory name if it is structured like "terraform-<system>-<name>"

Example output

```
$ pwd
/root/modules/terraform-rt-private-registry

$ rt publish --namespace=platform --version=2.5.0
Version:   1.4.0
Name:      private-registry
System:    rt
Directory: .
Size:      9 kB (3 kB compressed)
Publish to registrytools.cloud? You must type 'yes' to confirm:
```
