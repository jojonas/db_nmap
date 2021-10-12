package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jojonas/db_nmap/internal"
)

func ensureArgument(args []string, arguments ...string) []string {
	searchArg := arguments[0]

	for i, a := range args {
		// find the first argument in args
		if a == searchArg {
			// if found, copy all arguments into the args array
			for j, s2 := range arguments {
				args[i+j] = s2
			}
			return args
		}
	}

	return append(args, arguments...)
}

func findArgument(args []string, argument string) string {
	for i, a := range args {
		if a == argument {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
	}
	return ""
}

func hasArgument(args []string, argument string) bool {
	for _, a := range args {
		if a == argument {
			return true
		}
	}
	return false
}

func findOutputFile(args []string) string {
	output := findArgument(args, "-oX")
	if output != "" {
		return output
	}

	output = findArgument(args, "-oA")
	if output != "" {
		return output + ".xml"
	}

	return ""
}

func runNmap(cmd *exec.Cmd, handle internal.HandleHostFunc) error {
	resumeFilename := findArgument(cmd.Args, "--resume")
	isResume := resumeFilename != ""

	outputFilename := ""

	if isResume {
		log.Debugf("Resuming from file %q.", resumeFilename)
		outputFilename = strings.TrimSuffix(resumeFilename, filepath.Ext(resumeFilename)) + ".xml"
	} else {
		outputFilename = findOutputFile(cmd.Args)
		cmd.Args = ensureArgument(cmd.Args, "-oX", "/dev/fd/3")
	}

	var readerPipe io.Reader
	readerPipe, writerPipe, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("creating reader/writer pipe: %w", err)
	}
	defer writerPipe.Close()

	if outputFilename != "" {
		log.Debugf("Teeing to file %q.", outputFilename)

		var outputFile *os.File
		if isResume {
			outputFile, err = os.OpenFile(outputFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				return fmt.Errorf("opening %q for appending: %w", outputFilename, err)
			}
		} else {
			outputFile, err = os.Create(outputFilename)
			if err != nil {
				return fmt.Errorf("opening %v for writing: %w", outputFilename, err)
			}
		}
		defer outputFile.Close()

		readerPipe = io.TeeReader(readerPipe, outputFile)
	}

	cmd.ExtraFiles = []*os.File{writerPipe}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		err := internal.ParseNmapXML(readerPipe, handle)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error watching XML: %v\n", err)
		}
		wg.Done()
	}()

	log.Debugf("Running %q ...", cmd)
	err = cmd.Run()
	writerPipe.Close()

	if err != nil {
		return fmt.Errorf("running command %q: %w", cmd, err)
	}

	wg.Wait()

	return nil
}
