package downloader

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
)

func VerifyIntegrity(shasum, filePath string) error {
	hasher := sha1.New()

	file, openErr := os.Open(filePath)
	defer file.Close()

	if openErr != nil {
		return errors.Wrapf(openErr, "Error opening package")
	}

	if _, err := io.Copy(hasher, file); err != nil {
		return errors.Wrapf(err, "Error opening calculating shasum for package: %s", file.Name())
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))
	if checksum != shasum {
		// delete file if it fails integrity check
		deleteErr := os.Remove(filePath)
		if deleteErr != nil {
			log.Fatal("Error deleting file with failed integrity check - must stop")
			return deleteErr
		}
		return errors.New("File integrity check failed")
	}

	return nil
}
