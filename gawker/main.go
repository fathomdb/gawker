package main

import (
    "flag"
    "github.com/fathomdb/gawker"
    "log"
    "os"
    "strings"
)

//import _ "net/http/pprof"

func main() {
    //go http.ListenAndServe(":6060", nil)
    //	var err error

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
        "Working directory")

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
        "logfile to create to create")

    var listenAddresses string
    flags.StringVar(&listenAddresses,
        "listen",
        "tcp://127.0.0.1:777",
        "Address(es) on which to listen")

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

    err = runtime.Listen(pidfile, strings.Split(listenAddresses, ","))
    if err != nil {
        log.Panicf("Error from listeners: %v", err)
    }
}
