package update

import (
	"io"
	"updator/untar"
	"log"
	"os"
)

func Start(inputUpdate io.Reader, currentVersion string) error {
	//Before untar, remove old UPDATES file
	var err error
	_, err = os.Stat(newFilePath)
	if err != nil {
		if os.IsExist(err) {
			err = os.RemoveAll(newFilePath)
			if err != nil {
				return err
			}
			err = os.MkdirAll(newFilePath, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	err = untar.Untar(inputUpdate, newFilePath)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = loadScripts()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	err = executePre()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	err = loadVerification()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	err = checkMD5s()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	err = doUpdateFiles(currentVersion)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	err = executePost()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return cleanUp()
}
