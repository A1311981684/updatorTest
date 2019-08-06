package update

import (
	"os"
	"path/filepath"
	"log"
)

//Cleanup removes the extracted files in FILES
//BACKUPS and DOWNLOADS will be remained
func cleanUp() error {
	err := os.RemoveAll(filepath.Dir(filepath.Dir(newFilePath)))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	log.Println("Update done.")
	return nil
}
