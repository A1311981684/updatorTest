package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	files2 "updator/files"
	"updator/generate"
	"updator/rollback"
	"updator/update"
)

func main() {

	files, err := files2.GetAllFileNamesUnder("./")
	if err!=nil {
		panic(err)
	}
	for _, v :=range files {
		if strings.Contains(v, "date_") {
			err = os.Remove("./" + v)
			if err != nil {
				panic(err)
			}
			break
		}
	}

	scanner := bufio.NewScanner(os.Stdin)

	update.SetUpdateConfig(&update.UpdaterConfig{
		CurrentProjectName:"updatorTest",
		RestartManagerPath:`F:\VirtualBox\Ubuntu\GoPath\src\updatorTest\RestartManager.exe`,
	})

	var input string
	for scanner.Scan() {
		input = scanner.Text()
		switch input {
		case "pack":
			var content generate.UpdateContent

			content.Name = "updatorTest"
			content.Version = generate.VersionControl{From:[]string{"1.1", "1.2"}, To:"1.5", UpdateLog:[]string{"Nothing1", "Nothing2"}}
			content.Paths = []string{"F:\\VirtualBox\\Ubuntu\\GoPath\\src\\updatorTest\\AAAAAAA",
				"F:\\VirtualBox\\Ubuntu\\GoPath\\src\\updatorTest\\date"}
			content.Scripts = []string{}
			//Create
			outputFile, err := os.Create(".\\app.update")
			if err != nil {
				log.Println(err.Error())
				continue//contiune
			}
			err = generate.CreateUpdate(content, outputFile)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			log.Println("update file is created.")
			err = outputFile.Close()
			if err != nil {
				panic(err)
			}
		case "update":
			updateFile, err := os.Open("app.update")
			if err != nil {
				log.Println(err.Error())
				continue
			}
			err = update.Start(updateFile, "1.1")
			if err != nil {
				log.Println(err.Error())
			}
			err = updateFile.Close()
			if err != nil {
				panic(err)
			}

		case "restart":
			err := update.Restart()
			if err != nil {
				panic(err)
			}

		case "rollback":
			err = rollback.PerformRollback(update.GetUpdateConfig().RestartManagerPath, "./BACKUPS", "updatorTest")
			if err != nil {
				panic(err)
			}
			log.Println("DONE.")
		default:

		}
	}
}
