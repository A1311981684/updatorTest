package update

import (
	"encoding/gob"
	"os"
)

type Verification struct {
	MD5      ChecksumMap
	Versions VersionControl
}
var verification Verification

func loadVerification() error {
	f, err := os.Open(newFilePath + "VERIFICATION")
	if err != nil {
		return err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}()

	decoder := gob.NewDecoder(f)
	err = decoder.Decode(&verification)
	if err != nil {
		return err
	}

	_checksumMap = verification.MD5
	_versionController = verification.Versions
	return nil
}
