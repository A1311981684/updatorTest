package update

import (
	"log"
	"os"
	"os/exec"
	"time"
)

//Send a message to restart manager via stdin pipe
func Restart() error {
	var err error
	cmd := exec.Command(updateConfig.RestartManagerPath, os.Args[0])

	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	err = cmd.Run()
	if err != nil {
		return err
	}
	//Write message to pipe
	_, err = cmdIn.Write([]byte("r"))
	if err != nil {
		return err
	}
	log.Println("r written")
	err = cmdIn.Close()
	if err != nil {
		return err
	}
	//Exit and let restart manager do its things
	Exit()
	return nil
}

//TODO Do something to save data before exit?
func Exit() {
	time.AfterFunc(1 * time.Second, func() {
		os.Exit(0)
	})
}
