package main

import (
	"bufio"
	"log"
	"os"
	"updator/generate"
	"updator/update"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var input string
	for scanner.Scan() {
		input = scanner.Text()
		switch input {
		case "pack":
			var content generate.UpdateContent

			content.Name = "."
			content.Version = generate.VersionControl{From:[]string{"1.1", "1.2"}, To:"1.5", UpdateLog:[]string{"Nothing1", "Nothing2"}}
			content.Paths = []string{".\\AAAAAAA"}
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
		default:

		}
	}
}
