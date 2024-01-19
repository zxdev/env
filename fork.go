package env

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// Fork is an wrapper around NewEnv that enables a program to run normally
// or like a daemon with start|stop signals and control referencse are written
// to /var/fork/{name.pid} and should be left alone for proper Fork processing
func Fork(cfg ...interface{}) {

	env := NewEnv()

	if len(os.Args) > 1 {

		name := filepath.Base(os.Args[0])
		pidFile := Dir(env.Var, "fork", name+".pid")

		switch os.Args[1] {
		case "start":

			if _, err := os.Stat(pidFile); errors.Is(err, fs.ErrExist) {
				fmt.Fprintln(os.Stderr, "Already running!")
				os.Exit(0)
			}

			f, err := os.Create(pidFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to create %s!\n", pidFile)
				os.Exit(0)
			}

			// start as external process; remove fork start command
			cmd := exec.Command(os.Args[0], os.Args[2:]...)
			if err = cmd.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to start %s\n", name)
				f.Close()
				os.Remove(pidFile)
				os.Exit(0)
			}

			f.WriteString(strconv.Itoa(cmd.Process.Pid))
			f.Close()

			fmt.Fprint(os.Stderr, cmd.Process.Pid)
			os.Exit(0)

		case "stop":

			if _, err := os.Stat(pidFile); err != nil {
				fmt.Fprintln(os.Stderr, "Not running!")
				os.Exit(0)
			}
			data, _ := os.ReadFile(pidFile)
			pid, err := strconv.Atoi(string(data))
			if pid == 0 || err != nil {
				fmt.Fprintf(os.Stderr, "Unable to parse %s\n", pidFile)
				os.Exit(0)
			}

			if process, err := os.FindProcess(pid); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to locate %s +%d\n", name, pid)
			} else {
				if process.Signal(os.Interrupt) != nil {
					fmt.Fprintf(os.Stderr, "Unable to stop %s +%d\n", name, pid)
				}
			}

			os.Remove(pidFile)
			os.Exit(0)

		}

	}

	var p Parser
	p.Do(cfg...)

}
