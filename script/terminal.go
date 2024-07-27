package script

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)


func RunCommand(cmds string , logger *log.Logger ) (string, error) {

	// Print the shell prompt
	path, err := os.Getwd()
	if err != nil {
		logger.Printf("ERR : %s\n",err.Error())
		return "" , err
	} else {
		path += "> "
	}

	// Read the input from the user
	input := cmds

	// Trim the input string
	input = strings.TrimSpace(input)

	// Split the input into command and arguments
	args := strings.Split(input, " ")
	cmd := args[0]
	cmdArgs := args[1:]


	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	switch cmd {
	case "cd":
		if len(cmdArgs) < 1 {
	
			logger.Printf("ERR : %s\n",errors.New("cd: missing argument"))
			return "", errors.New("cd: missing argument")
		}

		dir := strings.TrimPrefix(cmds, "cd ")
   		dir = expandPath(dir)
   		err := os.Chdir(dir)
		if err != nil {
			logger.Printf("ERR : %s\n",err.Error())
			return "", err

		}
	
		logger.Printf("LOG : %s\n",dir + "> ")
		return fmt.Sprintf("LOG : %s>\n",dir) ,nil

	case "pwd":
		dir, err := os.Getwd()
		if err != nil {
			logger.Printf("ERR : %s\n",err.Error())
			return "", err
		} else {
			logger.Printf("LOG : %s\n",dir + "> ")
			return fmt.Sprintf("LOG : %s>\n",dir), nil
		}
	// case "":
	// 	return "", nil
	default:

		// Execute the bash command with '-c' flag
		execCmd := exec.CommandContext(ctx, "bash", "-c", cmds)
		out, err := execCmd.CombinedOutput()
		if err != nil {
			logger.Printf("ERR : %s\n",out)
			logger.Printf("ERR : %s\n",err.Error())
			return "", err
		}
		logger.Printf("LOG : %s\n",path + string(out))
		return string(out), nil

	}

}



// expandPath handles '~' and relative paths.
func expandPath(path string) string {
	if path[:2] == "~/" {
	 usr, _ := user.Current()
	 dir := usr.HomeDir
	 return filepath.Join(dir, path[2:])
	}
	return path
   }
