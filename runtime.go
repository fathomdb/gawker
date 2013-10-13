package gawker

import (
    "github.com/fathomdb/processes"
)

type Runtime struct {
    Processes *processes.WatchedProcessManager

    ConfDir string
    WorkDir string
}

func NewRuntime(confdir, workdir string) (*Runtime, error) {
    runtime := &Runtime{}

    runtime.WorkDir = workdir
    runtime.ConfDir = confdir

    var err error
    runtime.Processes, err = processes.NewWatchedProcessManager(workdir, confdir)
    if err != nil {
        return nil, err
    }

    return runtime, nil
}
