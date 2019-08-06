package rollback

import "os/exec"
import "os"
import "log"

//Ask restart manager to perform the rollback
//In case windows system forbid us to change running exe
func PerformRollback(manager , backupPath, appName string) error {
	var err error
	cmd := exec.Command(manager, backupPath, os.Args[0], appName)

	cmdIn, _ := cmd.StdinPipe()

	err = cmd.Start()
	if err != nil {
		return err
	}

	_, err = cmdIn.Write([]byte("b"))
	if err != nil {
		return err
	}

	err = cmdIn.Close()
	if err != nil {
		return err
	}
	log.Println("It should be done after a few seconds...")

	os.Exit(0)
	return nil
}
