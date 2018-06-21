package codeanalysis

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
	return result, nil
}

func (d *DiskUnpacker) UnpackPackage(packageFilePath string) (string, error) {
	return "", nil
}
