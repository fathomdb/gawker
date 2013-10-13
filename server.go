package gawker

import (
    "fmt"
    "github.com/fathomdb/processes"
    "io/ioutil"
    "log"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "time"
)

func (srv *Server) GetProcesses() []*processes.WatchedProcessInfo {
    return srv.Runtime.Processes.List()
}

func createPidFile(pidfile string) error {
    if pidString, err := ioutil.ReadFile(pidfile); err == nil {
        pid, err := strconv.Atoi(string(pidString))
        if err == nil {
            if _, err := os.Stat(fmt.Sprintf("/proc/%d/", pid)); err == nil {
                return fmt.Errorf("pid file found, ensure gawker is not running or delete %s", pidfile)
            }
        }
    }

    file, err := os.Create(pidfile)
    if err != nil {
        return err
    }

    defer file.Close()

    _, err = fmt.Fprintf(file, "%d", os.Getpid())

    if err == nil {
        log.Printf("Created pidfile at %s", pidfile)
    }

    return err
}

func removePidFile(pidfile string) {
    if err := os.Remove(pidfile); err != nil {
        log.Printf("Error removing %s: %s", pidfile, err)
    }
}

func (r *Runtime) Run(pidfile string) error {
    if err := createPidFile(pidfile); err != nil {
        log.Fatal(err)
    }
    defer removePidFile(pidfile)

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, os.Kill, os.Signal(syscall.SIGTERM))
    go func() {
        sig := <-c
        log.Printf("Received signal '%v', exiting\n", sig)
        removePidFile(pidfile)
        os.Exit(0)
    }()

    for {
        time.Sleep(10 * time.Second)
    }

    return nil
}

func NewServer(runtime *Runtime) (*Server, error) {
    server := &Server{
        Runtime: runtime,
    }
    return server, nil
}

type Server struct {
    Runtime *Runtime
}
