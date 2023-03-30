package fishi

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// file spec2go implements conversion of a fishi spec to a series of go files.
// This is the only way that we can do full validation with a hooks package.

// asynchronounsly copy the package to the target directory. returns non-nil
// error if scanning the package and directory creation was successful. later,
// pushes the first error that occurs while copying the contents of a file to
// the channel, or nil to the channel if the copy was successful.
func copyPackageToTargetAsync(goPackage string, targetDir string) (copyResult chan error, err error) {
	pkgs, err := packages.Load(nil, goPackage)
	if err != nil {
		return nil, fmt.Errorf("scanning package: %w", err)
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected one package, got %d", len(pkgs))
	}

	pkg := pkgs[0]

	// Permissions:
	// rwxr-xr-x = 755

	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("creating target dir: %w", err)
	}

	ch := make(chan error)
	go func() {
		for _, file := range pkg.GoFiles {
			baseFilename := filepath.Base(file)

			fileData, err := os.ReadFile(file)
			if err != nil {
				ch <- fmt.Errorf("reading source file %s: %w", baseFilename, err)
				return
			}

			dest := filepath.Join(targetDir, baseFilename)

			// write the file to the dest path
			err = os.WriteFile(dest, fileData, 0644)
			if err != nil {
				ch <- fmt.Errorf("writing source file %s: %w", baseFilename, err)
				return
			}
		}

		ch <- nil
	}()

	return ch, nil
}
