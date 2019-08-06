package update

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"updator/go-update"
)

type UpdaterConfig struct {
	CurrentProjectName string //the name of the app
	RestartManagerPath string //sub process executable path, used to restart/restore app
}

var updateConfig *UpdaterConfig
var Separator = string(filepath.Separator)

var backupPath string
var newFilePath string

func init() {
	abs := filepath.Dir(os.Args[0])
	backupPath = abs + Separator + "BACKUPS" + Separator
	_, err := os.Stat(backupPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(backupPath, os.ModePerm)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	newFilePath = abs + Separator + "UPDATES" + Separator + "FILES" + Separator
	if _, err := os.Stat(newFilePath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(newFilePath, os.ModePerm)
			if err != nil {
				panic(err)
			}
		} else {
			if err != nil {
				panic(err)
			}
		}
	}
	newFilePath, _ = filepath.Abs(newFilePath)
	newFilePath += string(filepath.Separator)
}

func SetUpdateConfig(config *UpdaterConfig) {
	updateConfig = config
}
func GetUpdateConfig() *UpdaterConfig {
	return updateConfig
}

/*
Sequences:
1 Un-tar the update package from DOWNLOADS to the FILES.
2 Verify if pre-version match current project(For example, must be updated from version 1.1, if current
version is not version 1.1, update will not be started).
3 Read the CHECKSUM gob file.
4 For each file name in the gob, calculate and verify its MD5 and compare with the MD5 that included in the gob.
5 If all the checksum are checked and matched, continue. Else, Return error.
6 Copy all the files that are going to be replaced by new files to the BACKUPS directory. If it is
found in new files but not found in current project, just ignore it and add it to the project.
7 Apply new files to cover old files.If error occurred, Recover copied backup files back to their original location.
*/

func doUpdateFiles(currentVersion string) error {
	if updateConfig == nil {
		return errors.New("update config not set")
	}
	//Verify Versions
	err := checkVersion(currentVersion)
	if err != nil {
		return err
	}
	//Apply the doUpdateFiles to all the files in the UPDATES/FILES path
	err = rangeFiles(newFilePath + updateConfig.CurrentProjectName)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	//doUpdateFiles successfully, save it to local file
	//SaveUpdatedVersion(currentVersion)
	return nil
}

func rangeFiles(path string) error {

	fileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println(err.Error(), fileInfos[0].Name(), path)
		return err
	}

	for _, v := range fileInfos {
		//Make sure path ends with separator
		if strings.LastIndex(path, Separator) != len(path)-1 {
			path += string(filepath.Separator)
		}

		if v.IsDir() {

			//Create backup directories according to the update package
			err = os.MkdirAll(strings.Replace(path+v.Name(), "UPDATES"+Separator+"FILES", "BACKUPS", 1), os.ModePerm)
			if err != nil {
				log.Println(err.Error())
				return err
			}
			//Continue to find next file to be updated
			err = rangeFiles(path + v.Name())
			if err != nil {
				log.Println(err.Error())
				return err
			}
		} else {

			//Open the new file
			f, err := os.Open(path + v.Name())

			if err != nil {
				log.Println(err.Error(), path, v.Name())
				return err
			}

			tgp, err := filepath.Abs(strings.Replace(path, Separator+updateConfig.CurrentProjectName+Separator+"UPDATES"+Separator+"FILES", "", 1) + v.Name())
			if err != nil {
				return err
			}
			osp := strings.Replace(path, "UPDATES"+Separator+"FILES", "BACKUPS", 1) + v.Name()
			options := update.Options{
				//TargetPath refers to the file needed to be replaced by a new file
				TargetPath: tgp,
				TargetMode: os.ModePerm,
				//OldSavePath refers to the backup file path
				OldSavePath: osp,
			}
			// get the directory of the target file exists in
			updateDir := filepath.Dir(tgp)
			//filename := filepath.Base(tgp)
			//Make sure updateDir does exist
			_, err = os.Stat(updateDir)
			if err != nil {
				log.Println(err.Error())
				if os.IsNotExist(err) {
					err = os.MkdirAll(updateDir, os.ModePerm)
					if err != nil {
						log.Println(err.Error())
						return err
					}
				} else {
					log.Println(err.Error())
					return err
				}
			}
			//Make sure OldSavePath does exist
			backupDir := filepath.Dir(osp)
			_, err = os.Stat(backupDir)
			if err != nil {
				log.Println(err.Error())
				if os.IsNotExist(err) {
					err = os.MkdirAll(backupDir, os.ModePerm)
					if err != nil {
						log.Println(err.Error())
						return err
					}
				} else {
					log.Println(err.Error())
					return err
				}
			}

			err = update.Apply(f, options)
			if err != nil {
				log.Println(err.Error())
				return err
			}

			err = f.Close()
			if err != nil && strings.Contains(err.Error(), os.ErrClosed.Error()) {
				//log.Println("file closed.")
				continue
			}else if err != nil {
				log.Println(err.Error())
				return err
			}
		}
	}
	return nil
}
