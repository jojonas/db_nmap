package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type HandleHostFunc func(host NmapHost) error

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

func parseNmapXML(reader io.Reader, handle HandleHostFunc) error {
	decoder := xml.NewDecoder(reader)

outer:
	for {
		token, err := decoder.Token()

		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Debug("Unexpected EOF")
				break outer
			}

			return fmt.Errorf("reading token: %w", err)
		}
		if token == nil {
			log.Debug("Unexpected end")
			break outer
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "nmaprun":
				header := NmaprunHeader{}
				err = decoder.DecodeElement(&header, &t)
				if err != nil {
					return fmt.Errorf("reading <nmaprun>: %w", err)
				}

				checkVersion(header)
			case "host":
				host := NmapHost{}
				err = decoder.DecodeElement(&host, &t)
				if err != nil {
					return fmt.Errorf("reading <host>: %w", err)
				}

				err := handle(host)
				if err != nil {
					return fmt.Errorf("handling <host>: %w", err)
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "nmaprun":
				log.Debug("XML document complete")
				break outer
			}
		default:
		}
	}

	return nil
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

func checkVersion(header NmaprunHeader) {
	testedVersions := []string{"7.92"}
	isTested := false
	for _, version := range testedVersions {
		if header.Version == version {
			isTested = true
			break
		}
	}

	if !isTested {
		log.Warnf("db_nmap was not tested against Nmap version %s!", header.Version)
	}
}

func runNmap(cmd *exec.Cmd, handle HandleHostFunc) error {
	outputFilename := findOutputFile(cmd.Args)
	cmd.Args = ensureArgument(cmd.Args, "-oX", "/dev/fd/3")

	var readerPipe io.Reader
	readerPipe, writerPipe, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("creating reader/writer pipe: %w", err)
	}
	defer writerPipe.Close()

	if outputFilename != "" {
		outputFile, err := os.Create(outputFilename)
		if err != nil {
			return fmt.Errorf("opening %v for writing: %w", outputFilename, err)
		}
		defer outputFile.Close()

		readerPipe = io.TeeReader(readerPipe, outputFile)
	}

	cmd.ExtraFiles = []*os.File{writerPipe}

	go func() {
		err := parseNmapXML(readerPipe, handle)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error watching XML: %v\n", err)
		}
	}()

	log.Debugf("Running %q ...", cmd)
	err = cmd.Run()
	writerPipe.Close()

	if err != nil {
		return fmt.Errorf("running command %q: %w", cmd, err)
	}

	return nil
}
