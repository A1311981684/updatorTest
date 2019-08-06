package update

import (
	"crypto/md5"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ChecksumMap struct {
	MD5s map[string]string
}

var _checksumMap ChecksumMap

//var _checksumMap map[string]string

// First, load Checksum GOB from extracted update package files
// Second, go through all file paths in the GOB's map key, calculate their MD5 and compare with GOB's map value respectively
func checkMD5s() error {

	err := CheckChecksum()
	if err != nil {
		return err
	}
	return nil
}

// Checksum verification Traversal
func CheckChecksum() error {
	if _checksumMap.MD5s == nil || len(_checksumMap.MD5s) == 0 {
		return errors.New("verification map is empty or not loaded")
	}
	//Try to get the extracted new files, the only directory extracted should be the root directory: projectName
	files, err := ioutil.ReadDir(newFilePath)
	if err != nil {
		return err
	}
	var dirs = make([]string, 0)
	for _, v := range files {
		if v.IsDir() {
			dirs = append(dirs, v.Name())
		}
	}
	if len(dirs) == 0 {
		return errors.New("no extracted update directory found" + newFilePath)
	} else if len(dirs) > 1 {
		return fmt.Errorf("too many directories found in the extracted content: %d, %v", len(dirs), dirs)
	}
	//Check if update package project name match current project name
	if dirs[0] != updateConfig.CurrentProjectName {
		return errors.New("update package does not match")
	}

	//Recursion function to traverse checksums
	var rangeFunc func(path string) error
	rangeFunc = func(path string) error {
		//Read all subjects under the path
		files, err = ioutil.ReadDir(path)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		//For each subject found
		for _, v := range files {
			if v.IsDir() {
				//If this subject is a directory, re-do rangeFunc to it
				err = rangeFunc(path + string(filepath.Separator) + v.Name())
				if err != nil {
					return err
				}
			} else {
				//If this subject is a file:
				//Get extracted verification from _checksumMap, the method to the key used to find the verification value is：
				//1 Get the new file path :[Path][Separator][FileName].
				//2 Get the relative path of the old file :Replace [New file directory(extracted directory)] with Empty
				// string "" in the new file path.
				//3 The key is the relative path of the old file.
				mapsSeparator := ""
				for k := range _checksumMap.MD5s {
					if strings.Contains(k, "\\") {
						mapsSeparator = "\\"
						break
					} else {
						mapsSeparator = "/"
						break
					}
				}

				fileName := v.Name()
				if v.Name()[:1] == "." {
					fileName = fileName[1:]
				}
				newFile := path + Separator + fileName
				keyPath := strings.Replace(newFile, newFilePath, "", 1)
				keyPath = strings.Replace(keyPath, Separator, mapsSeparator, -1)
				if md5Ext, ok := _checksumMap.MD5s[keyPath]; ok {
					err = Check(newFile, md5Ext)
					if err != nil {
						return err
					}
				} else {
					log.Println("find verification in map failed:", keyPath)
					log.Println(_checksumMap)
					return errors.New("find verification in map failed: " + keyPath)
				}

			}
		}
		return nil
	}

	return rangeFunc(newFilePath + dirs[0])
}

//Check MD5 of a single file
func Check(filePath, checksum string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	//calculate the file MD5
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Println("Copy:", err)
		return err
	}
	bytesMD5 := h.Sum(nil)
	strMD5 := fmt.Sprintf("%x", bytesMD5)

	if strMD5 != checksum {
		reverseCap := strings.ToUpper(strMD5)
		if reverseCap != checksum {
			log.Println(filePath, reverseCap, checksum)
			return errors.New("MD5 verification not match")
		}
	}
	return nil
}
