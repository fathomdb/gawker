package gawker

import (
    "github.com/fathomdb/processes"
)

type Runtime struct {
    Processes  *processes.WatchedProcessManager
    Containers *ContainerManager
    Images     *ImageManager

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

    runtime.Containers = NewContainerManager(runtime)
    runtime.Images = NewImageManager(runtime)

    return runtime, nil
}
