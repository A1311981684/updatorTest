package generate

import (
	"archive/tar"
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"updator/files"
)

//All the path need to be tarred
type UpdateContent struct {
	Name    string         `json:"name"`           // root name of the app
	Paths   []string       `json:"paths"`          //modified files paths
	Scripts []string       `json:"scripts"`        //executable scripts paths
	Version VersionControl `json:"VersionControl"` // version allowed from ... to a new version, and update log
}

//Mark version constrain
type VersionControl struct {
	From []string
	To   string
	UpdateLog []string
}

//Store Key = path, Value = MD5 string
type checksumMap struct {
	MD5s map[string]string
}

//Wrap two significant verification source into one
type verification struct {
	Versions VersionControl
	MD5      checksumMap
}

var targetAppName = "" //project name
var tempPath = ""      //temp path is used to store files need to be packaged(copy all the new files to this path)
var checksum checksumMap

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	absPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	if absPath[len(absPath)-1:] != string(filepath.Separator) {
		absPath += string(filepath.Separator)
	}
	tempPath = absPath + "Temp"
	checksum.MD5s = make(map[string]string)
	log.Println("Temp:", tempPath)
}

// Update file structure:
// *.update
//		|
//		|---appName "directory contains the new files and directories"
//		|
//		|---VERIFICATION "contains a map which KEY = [relative path, name start with appName], VALUE = [MD5]; And versions constrain"
//		|
//		|---...*.sh [many scripts if exist]
//
func CreateUpdate(content UpdateContent, outputWriter io.Writer) error {
	paths, err := convertPathByOS(content.Paths)
	if err != nil {
		return err
	}

	//Path must contain appName
	for _, v := range paths {
		if !strings.Contains(v, content.Name) {
			return errors.New("this appName does not match source path: " + content.Name)
		}
	}
	targetAppName = content.Name

	//append separator to temp path if needed
	if tempPath[len(tempPath)-1:] != string(filepath.Separator) {
		tempPath += string(filepath.Separator)
	}

	exits, err := files.PathExists(tempPath)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	if !exits {
		err = os.MkdirAll(tempPath, os.ModePerm)
		if err != nil {
			log.Println(err.Error())
			return err
		}
	}

	//Copy to temp path
	err = copyToANewDirectory(paths, tempPath)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	for _, v := range paths {
		err = walkPathFunc(v)
		if err != nil {
			log.Printf("filepath.Walk() error: %v\n", err)
			return err
		}
	}

	for k, v := range checksum.MD5s {
		log.Printf("Key: %v Value:%v \n", k, v)
	}

	//---------------生成gob文件-----------------------
	if  len(content.Version.From) == 0 || content.Version.To == "" {
		return errors.New("empty version information")
	}

	verification := verification{
		content.Version,
		checksum,
	}
	gobPath := tempPath + string(filepath.Separator) + "VERIFICATION"
	file, err := os.Create(gobPath)
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(verification)
	if err != nil {
		return err
	}
	//--------------gob生成完毕-------------------------

	//打包目录
	src := []string{tempPath + content.Name, tempPath + "VERIFICATION"}
	//If scripts exist, add them to the tar path
	if len(content.Scripts) > 0 {
		for _, v := range content.Scripts {
			src = append(src, v)
		}
	}
	err = tarFilesDirs(src, outputWriter)
	log.Println(src)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}

	return CleanTemp()
}

func copyToANewDirectory(from []string, to string) error {
	log.Println("Copy to:", to)
	if to[len(to)-1:] != string(filepath.Separator) {
		to += string(filepath.Separator)
	}
	fileInfos, err := ioutil.ReadDir(to)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	for _, v := range fileInfos {
		if v.IsDir() {
			err = os.RemoveAll(to + v.Name())
			if err != nil {
				log.Println(err.Error())
				return err
			}
		} else {
			err = os.Remove(to + v.Name())
			if err != nil {
				log.Println(err.Error())
				return err
			}
		}
	}

	// Copy TOI to a temp directory
	// Want to move D:\\targetAppName\\conf\\cfg.json to
	//              D:\\Temp\\targetAppName\\conf\\cfg.json
	// To achieve this, cut path at project name into two pieces, join the later one to the
	// target path
	for _, v := range from {
		//Not allowed
		if strings.Contains(v, ".idea") || strings.Contains(v, ".git") ||
			strings.Contains(v, "BACKUPS") || strings.Contains(v, "UPDATES"){
			continue
		}
		fileInfo, err := os.Stat(v)
		if err != nil {
			log.Println(err.Error(), v)
			return err
		}

		if fileInfo.IsDir() {
			tgt := to + targetAppName
			if filepath.Base(tgt) != targetAppName {
				tgt += string(filepath.Separator) + fileInfo.Name()
			}
			log.Println(v, tgt, targetAppName)
			files.CopyDir(v, tgt, targetAppName)
		} else {
			log.Println("aha")
			_, err = files.CopyFile(v, to+targetAppName+string(filepath.Separator)+fileInfo.Name())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func walkPathFunc(path string) error {
	if strings.Contains(path, ".idea") || strings.Contains(path, ".git") ||
		strings.Contains(path, "BACKUPS") || strings.Contains(path, "UPDATES"){
		return nil
	}

	info, err := os.Lstat(path) //返回描述文件的FileInfo信息。
	if err != nil {
		return err
	}

	if !info.IsDir() { // abbreviation for Mode().IsDir()　文件是否为目录
		err = addMD5(path, info) //如果是文件,调用walkFunc 求得对应的 Md5
		return err
	}

	names, err := readDirNames(path) //取得该路径下所有文件名称,存在names中
	if err != nil {
		return err
	}

	for _, name := range names { //遍历path下的所有文件

		//Here, if path ending with a separator,  remove it to make sure what next (adding a separator would be successful)
		if strings.LastIndex(path, string(filepath.Separator)) == len(path)-1 {
			path = path[:len(path)-1]
		}

		filename := path + string(filepath.Separator) + name //filename=path+name
		fileInfo, err := os.Lstat(filename)
		if err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			if err = addMD5(filename, fileInfo); err != nil {
				return err
			}
		} else {
			err = walkPathFunc(filename) //为文件夹，调用自己
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//Add md5
func addMD5(path string, info os.FileInfo) error {

	if info == nil {
		return nil
	}
	if !info.IsDir() {
		//输出如果是文件,则提示为文件
		//读取文件内容并生成对应的效验码   md5Hash  编码
		data, err := ioutil.ReadFile(path) //读取文件内容，并返回[]byte数据和错误信息。err == nil时，读取成功
		if err != nil {
			panic(err)
		}

		md5Hash := md5.New()
		md5Hash.Write(data)
		byteSlice := md5Hash.Sum(nil)
		md5Str := fmt.Sprintf("%x", byteSlice)
		index := strings.Index(path, targetAppName)
		//Key is a sub string of path which split from first target app name
		//For example, C:\\Temp\\app\\Conf\\cfg.json => app\\Conf\\cfg.json
		checksum.MD5s[path[index:]] = md5Str
		return nil
	} else {
		return nil
	}
}

//Return all names under the dir, including files and directories
func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	err = f.Close()
	if err != nil {
		return nil, err
	}
	return names, nil
}

func tarFilesDirs(paths []string, writer io.Writer) error {

	//simple tar writer without compress
	tw := tar.NewWriter(writer)
	defer func() {
		err := tw.Close()
		if err != nil {
			panic(err)
		}
	}()

	// add each file/dir as needed into the current tar archive
	for _, i := range paths {
		if err := tarIt(i, tw); err != nil {
			return err
		}
	}

	return nil
}

func tarIt(source string, tw *tar.Writer) error {

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}
	//For creating corresponding directory in the tar file for each file
	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}
	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Println(err.Error())
				return err
			}

			//Ignored path
			if strings.Contains(path, ".idea") || strings.Contains(path, ".git") ||
				strings.Contains(path, "BACKUPS") || strings.Contains(path, "UPDATES") {
				return nil
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				log.Println(err.Error())
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if err := tw.WriteHeader(header); err != nil {
				log.Println(err.Error())
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				log.Println(err.Error())
				return err
			}

			defer func() {
				err := file.Close()
				if err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(tw, file)
			if err != nil {
				log.Println(err.Error())
				return err
			}
			return err
		})
}

//deal with different OS path
//Convert linux path to windows path or convert linux path to windows path if needed
func convertPathByOS(paths []string) ([]string, error) {
	var res = make([]string, len(paths))

	switch runtime.GOOS {
	case "windows":
		var needConvert = false
		for _, v := range paths {
			//It has windows separator
			//Consider it has a format like windows : "F:\\Temp\\app\\conf"
			//								linux	: "/mnt/f/Temp/app/conf"
			//Whether the path is valid won't be checked. If invalid, it will return error afterward
			if strings.Contains(v, "\\") {
				continue
			}else if strings.Contains(v, "/") && strings.Contains(v, "mnt") { //If can't find any windows separator, at least we find unix separator instead
				needConvert = true
				break
			}else{ //Find no separator, not allowed
				return nil, errors.New("invalid path: "+ v)
			}
		}
		if needConvert {
			for k, v := range paths {
				// /mnt/f/Temp/app/conf
				// Replace all / to \\
				str := strings.Replace(v, "/", "\\", -1)
				// \\mnt\\f\\Temp\\app\\conf
				//Remove \\mnt\\
				str = strings.Replace(str, "\\mnt\\", "", 1)
				// f\\Temp\\app\\conf
				//Add ":", either f or F is OK
				disk := str[:1] // f
				path := str[1:]
				finalPath := disk + ":" + path
				// f:\\Temp\\app\\conf
				res[k] = finalPath
			}
			return res, nil
		}else{ //No need to convert
			return paths, nil
		}
	case "linux":
		var needConvert = false
		for _, v := range paths {
			if strings.Contains(v, "/") {
				continue
			}else if strings.Contains(v, "\\") && strings.Contains(v, ":") {
				needConvert = true
			} else {
				return nil, errors.New("invalid path: "+ v)
			}
		}
		if needConvert {
			for k, v := range paths {
				// f:\\Temp\\app\\conf or F:\\Temp\\app\\conf
				// replace \\ to /
				str := strings.Replace(v, "\\", "/", -1)
				// f:/Temp/app/conf or F:/Temp/app/conf
				//remove ":"
				str = strings.Replace(str, ":", "", 1)
				// f/Temp/app/conf or F/Temp/app/conf
				// disk path is case-sensitive, require lower-case
				disk := "/mnt/" + strings.ToLower(str[0:1])
				path := str[1:]
				finalResult := disk + path
				res[k] = finalResult
			}
			return res, nil
		}else {
			return paths, nil
		}
	default:
		return nil, errors.New("unsupported operating system: " + runtime.GOOS)
	}
}

func CleanTemp() error {
	//err := os.RemoveAll(tempPath)
	//if err != nil {
	//	if !os.IsNotExist(err) {
	//		return err
	//	}
	//}
	return nil
}