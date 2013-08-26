package gawker

import (
    "flag"
    "github.com/fathomdb/gommons"
    "io/ioutil"
    "log"
    "os"
    "strings"
)

func ReadAll(filename string) (contents string, err error) {
    fileBytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return
    }

    return string(fileBytes), nil
}

func ParseWithConfigurationFile(f *flag.FlagSet, confFileVar *string) {
    f.Parse(os.Args[1:])

    if *confFileVar == "" {
        defaultConfFile := os.Args[0] + ".conf"
        log.Printf("Checking for %s\n", defaultConfFile)
        exists, e := gommons.FileExists(defaultConfFile)
        if e != nil {
            log.Panic("Error checking for config file\n", e)
        }
        if exists {
            *confFileVar = defaultConfFile
        }
    }

    if *confFileVar != "" {
        exists, e := gommons.FileExists(*confFileVar)
        if e != nil {
            log.Panic("Error checking for config file\n", e)
        }

        if exists {
            log.Printf("Parsing config file: %s\n", *confFileVar)

            confFileContents, e := ReadAll(*confFileVar)
            if e != nil {
                log.Panic("Error reading config file\n", e)
            }

            args := os.Args[1:]

            for _, line := range strings.Split(confFileContents, "\n") {
                line = strings.TrimSpace(line)
                if line == "" {
                    continue
                }

                if strings.HasPrefix(line, "#") {
                    continue
                }

                if !strings.HasPrefix(line, "--") {
                    line = "--" + line
                }
                args = append(args, line)
            }

            //log.Printf("Lines %v\n", args)

            f.Parse(args)
        } else {
            log.Printf("Configuration file does not exist: %s", *confFileVar)
        }
    }
}
