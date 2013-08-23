package gawker

import (
    "fmt"
    "github.com/fathomdb/processes"
    "io/ioutil"
    "log"
    "os"
    "os/signal"
    "strconv"
    "strings"
    "syscall"
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

func (r *Runtime) Listen(pidfile string, listenAddresses []string) error {
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
    server, err := NewServer(r)
    if err != nil {
        return err
    }
    chErrors := make(chan error, len(listenAddresses))
    for _, listenAddress := range listenAddresses {
        listenAddressParts := strings.SplitN(listenAddress, "://", 2)
        if listenAddressParts[0] == "unix" {
            syscall.Unlink(listenAddressParts[1])
        } else if listenAddressParts[0] == "tcp" {
            if !strings.HasPrefix(listenAddressParts[1], "127.0.0.1") {
                log.Panic("Binding on anything other than 127.0.0.1 is blocked for security reasons")
            }
        } else {
            log.Fatalf("Invalid protocol format: %s", listenAddress)
            os.Exit(-1)
        }
        go func() {
            chErrors <- ListenAndServe(listenAddressParts[0], listenAddressParts[1], server, true)
        }()
    }
    for i := 0; i < len(listenAddresses); i += 1 {
        err := <-chErrors
        if err != nil {
            return err
        }
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
