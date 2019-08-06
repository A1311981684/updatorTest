package update

import (
	"github.com/pkg/errors"
	"log"
)

type VersionControl struct {
	From []string
	To   string
}

var _versionController VersionControl

func checkVersion(currentVersion string) error {
	for _, v := range _versionController.From {
		if v == currentVersion {
			return nil
		}
	}
	return errors.New("this current version does not match the update package: " + currentVersion)
}

func SaveUpdatedVersion(currentVersion string) {
	log.Println("Successfully updated to this version" + currentVersion)
}