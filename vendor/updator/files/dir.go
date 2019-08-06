package files

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

//Only care about files under the root path, skip every directory if exists
func GetAllFileNamesUnder(path string) ([]string, error) {
	//Read all content under the path
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var fileNames []string
	//Classify type of each content
	for _, v := range files {
		if !v.IsDir() {
			//If it is a file type, add to the result
			fileNames = append(fileNames, v.Name())
		} else {
			log.Println("[GetAllFileNames() :warning: path contains directory]")
		}
	}
	return fileNames, nil
}

func GetCurrentProjectName() (string, error) {
	//First get the full path of current working directory
	dirPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return "", err
	}
	// "/home/app/app.exe", "/app/app.exe", "D:\\project\\app\\app.exe
	projectName := filepath.Base(filepath.Dir(dirPath))
	if projectName == "." || projectName == string(filepath.Separator) {
		return "", errors.New("get name failed, got a '.' or separator")
	}
	return projectName, nil
}