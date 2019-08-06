package files

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CopyDir(src string, dst string, appName string) {

	err := filepath.Walk(src, func(src string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			//let Walk handle it
		} else {
			if strings.Contains(src, ".idea") || strings.Contains(src, ".git") ||
				strings.Contains(src, "BACKUPS") || strings.Contains(src, "UPDATES") ||
				strings.Contains(src, "Temp") || strings.Contains(filepath.Base(src), ".update") ||
				filepath.Ext(src) == "new" || filepath.Ext(src) == ".log" || strings.Contains(filepath.Base(src), "RestartManager") {
				return nil
			}
			log.Println("My SRC:", src)
			replacePartIndex := strings.Index(src, appName)
			appendPart := src[replacePartIndex:]
			appendPart2 := filepath.Dir(dst) + string(filepath.Separator)
			newPath := appendPart2 + appendPart

			_, err = CopyFile(src, newPath)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//copy file
func CopyFile(src, dst string) (w int64, err error) {

	state, err := os.Stat(src)
	if err != nil {
		return 0, err
	}
	if state.IsDir() {
		return 0, errors.New(src + " is not a file")
	}

	srcFile, err := os.Open(src)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer func() {
		err := srcFile.Close()
		if err != nil {
			panic(err)
		}
	}()

	destDir := filepath.Dir(dst)

	b, err := PathExists(destDir)
	if b == false {
		err := os.MkdirAll(destDir, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	dstFile, err := os.Create(dst)

	if err != nil {
		return
	}

	defer func() {
		err := dstFile.Close()
		if err != nil {
			panic(err)
		}
	}()

	return io.Copy(dstFile, srcFile)
}

