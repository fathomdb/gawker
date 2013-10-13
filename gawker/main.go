package main

import (
    "flag"
    "github.com/fathomdb/gawker"
    "log"
    "os"
)

func main() {
    flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

    var confFile string
    flags.StringVar(&confFile,
        "conf",
        "/etc/gawker/gawker.conf",
        "File with more configuration values")

    var workdir string
    flags.StringVar(&workdir,
        "workdir",
        "/var/gawker",
        "Working directory (for log & pid files)")

    var confdir string
    flags.StringVar(&confdir,
        "confdir",
        "/etc/gawker",
        "Configuration directory")

    var pidfile string
    flags.StringVar(&pidfile,
        "pidfile",
        "/var/run/gawker.pid",
        "pidfile to create")

    var logfile string
    flags.StringVar(&logfile,
        "logfile",
        "/var/log/gawker.log",
        "Main log file to create")

    gawker.ParseWithConfigurationFile(flags, &confFile)

    logf, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
    if err != nil {
        log.Fatalln("Cannot open log file", err)
    }
    log.SetOutput(logf)

    runtime, err := gawker.NewRuntime(confdir, workdir)

    if err != nil {
        log.Panicf("Error initializing %v", err)
    }

    err = runtime.Run(pidfile)
    if err != nil {
        log.Panicf("Error while running: %v", err)
    }
}
