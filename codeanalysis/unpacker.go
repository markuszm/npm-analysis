package codeanalysis

import (
	"bytes"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path"
)

type Unpacker interface {
	UnpackPackages(packages map[string]string) (map[string]string, error)
	UnpackPackage(packageFilePath string) (string, error)
}

type DiskUnpacker struct {
	TempFolder string
}

func NewDiskUnpacker(tempFolder string) *DiskUnpacker {
	return &DiskUnpacker{TempFolder: tempFolder}
}

func (d *DiskUnpacker) UnpackPackages(packages map[string]string) (map[string]string, error) {
	result := make(map[string]string, len(packages))
	for pkg, pkgPath := range packages {
		extractPath, err := d.UnpackPackage(pkgPath)
		if err != nil {
			return result, err
		}
		result[pkg] = extractPath
	}
	return result, nil
}

func (d *DiskUnpacker) UnpackPackage(packageFilePath string) (string, error) {
	extractPath := path.Join(d.TempFolder, path.Base(packageFilePath))
	err := unpackWithArchiver(packageFilePath, extractPath)
	return extractPath, err
}

func unpackWithArchiver(packageFilePath, extractPath string) error {
	a := archiver.MatchingFormat(extractPath)
	err := a.Open(packageFilePath, extractPath)
	return err
}

func unpackWithTar(packageFilePath, extractPath string) error {
	err := os.Mkdir(extractPath, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "could not create folder to extract to")
	}
	cmd := exec.Command("tar", "-xf", packageFilePath, "-C", extractPath)
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	err = cmd.Run()
	if err != nil {
		return errors.Wrapf(err, errOut.String())
	}
	return nil
}
