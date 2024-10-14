package publish

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hashicorp/go-slug"
)

func slugDirectoryToFile(dir string, writer io.Writer) (int64, error) {
	packer, err := slug.NewPacker(slug.ApplyTerraformIgnore(), slug.DereferenceSymlinks())
	if err != nil {
		return 0, fmt.Errorf("failed to init slug packer. %w", err)
	}

	meta, err := packer.Pack(dir, writer)
	if err != nil {
		return 0, fmt.Errorf("failed to pack specified directory: %w", err)
	}

	return meta.Size, nil
}

// PackAsFile slugs the specified directory as a temp file. It is the caller's
// responsibility to close and remove the file after it is used. The file is
// returned ready to be read, seeked to offset 0.
func PackAsFile(dir string) (string, int64, error) {
	file, err := os.CreateTemp("", "slug")
	if err != nil {
		return "", 0, fmt.Errorf("failed to create temp file. %w", err)
	}
	defer file.Close()

	size, err := slugDirectoryToFile(dir, file)
	if err != nil {
		return file.Name(), size, err
	}

	log.Printf("[DEBUG] Packed %q into %q (%d bytes)", dir, file.Name(), size)

	_, err = file.Seek(0, 0)
	if err != nil {
		return file.Name(), size, fmt.Errorf("failed to seek to beginning of temp file: %w", err)
	}

	return file.Name(), size, nil
}
