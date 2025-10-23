package hooks

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func RunHooks(hookName, hookDir string) error {
	if _, err := os.Stat(hookDir); os.IsNotExist(err) {
		// there are no hooks to run, so just bail
		return nil
	}
	files, err := os.ReadDir(hookDir)
	if err != nil {
		// there was an error getting the list of files in the hookdirectory
		// just bail to prevent issues
		fmt.Printf("failed to read hook directory: %s", err)
		return nil
	}

	hookCount := 0
	for _, file := range files {
		if file.IsDir() || !isExecutable(fmt.Sprintf("%s/%s", hookDir, file.Name())) {
			// skip if directory or not executable
			continue
		}
		hookCount++

		// print the step header for build log parser
		fmt.Printf("##############################################\nBEGIN %s - %d\n##############################################\n", hookName, hookCount)
		st := time.Now()
		binaryPath := filepath.Join(hookDir, file.Name())

		// execute the binary and capture its error
		// stream stdout and stderr
		var stdBuffer bytes.Buffer
		mw := io.MultiWriter(os.Stdout, &stdBuffer)
		cmd := exec.Command(binaryPath)
		cmd.Stdout = mw
		cmd.Stderr = mw
		err := cmd.Run()
		exitCode := cmd.ProcessState.ExitCode()

		var warning string
		if exitCode == 1 {
			// for 1 return the error and fail the build
			return fmt.Errorf("hook %s %d failed with exit code %d: %v", hookName, hookCount, exitCode, err)
		} else if exitCode > 1 {
			// if the exit code is greater than 1, we will consider as a warning
			// set the warning flag for the step footer
			warning = " WithWarnings"
			// add to the warnings counter to add to the end of the build to flag build as warning
			f, err := os.OpenFile("/tmp/warnings", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			if _, err = f.WriteString(fmt.Sprintf("%s:%d\n", hookName, hookCount)); err != nil {
				panic(err)
			}
		}

		et := time.Now()
		diff := time.Time{}.Add(et.Sub(st))
		tz, _ := et.Zone()
		// print the step footer for build log parser
		fmt.Printf("##############################################\nSTEP %s - %d: Completed at %s (%s) Duration %s Elapsed %s%s\n##############################################\n", hookName, hookCount, et.Format("2006-01-02 15:04:05"), tz, diff.Format("15:04:05"), diff.Format("15:04:05"), warning)
	}

	return nil
}

func isExecutable(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return (info.Mode() & 0111) != 0
}
