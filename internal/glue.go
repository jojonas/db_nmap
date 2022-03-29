package internal

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

var TestedVersions = []string{"7.40", "7.70", "7.80", "7.92"}

type HandleHostFunc func(host NmapHost) error

func ParseNmapXML(reader io.Reader, handle HandleHostFunc) error {
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
				for _, attr := range t.Attr {
					if attr.Name.Local == "version" {
						checkVersion(attr.Value)
						break
					}
				}
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
				log.Debug("XML document complete.")
				break outer
			}
		default:
		}
	}

	return nil
}

func checkVersion(version string) {
	isTested := false
	for _, testedVersions := range TestedVersions {
		if version == testedVersions {
			isTested = true
			break
		}
	}

	if !isTested {
		log.Warnf("This program was not tested against Nmap version %s.", version)
	}
}
