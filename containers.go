package gawker

import (
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "github.com/fathomdb/gommons"
    "github.com/fathomdb/processes"
    "io"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "strings"
)

type ContainerManager struct {
    runtime *Runtime
}

type LxcConfig struct {
    ManualConfig string
}

type InjectFile struct {
    Path     string
    Contents []byte
    Mode     int
}

type ContainerInfo struct {
    Key         string
    Image       string
    LxcConfig   LxcConfig
    InjectFiles []InjectFile
}

func NewContainerManager(runtime *Runtime) (s *ContainerManager) {
    s = &ContainerManager{}
    s.runtime = runtime
    return s
}

func (s *ContainerManager) GetContainersDir() string {
    return s.runtime.WorkDir + "/vms"
}

func (s *ContainerManager) List() (ret []*ContainerInfo, err error) {
    ret = []*ContainerInfo{}

    dir := s.GetContainersDir()

    files, err := gommons.ListDirectory(dir)
    if err != nil {
        log.Printf("Error listing files in vms dir", err)
        return nil, err
    }

    for _, file := range files {
        if !file.IsDir() {
            continue
        }

        key := file.Name()
        if key == "" {
            continue
        }

        if strings.HasPrefix(key, ".") {
            continue
        }

        info := &ContainerInfo{}
        info.Key = key

        ret = append(ret, info)
    }

    return ret, nil
}

func (s *ContainerManager) GetContainerInfo(key string) (*ContainerInfo, error) {
    err := gommons.CheckSafeName(key)
    if err != nil {
        return nil, err
    }

    dir := s.GetContainersDir()

    path := dir + "/" + key

    infopath := path + "/config.json"
    info, err := readContainerInfo(infopath)
    if err != nil {
        log.Printf("Error reading container info %s %v", infopath, err)
        return nil, err
    }

    return info, nil
}

func readContainerInfo(path string) (*ContainerInfo, error) {
    c := &ContainerInfo{}
    err := gommons.ReadJson(path, c)
    if err != nil {
        return nil, err
    }
    return c, nil
}

func injectFiles(rootfs string, config *ContainerInfo) error {
    for _, inject := range config.InjectFiles {
        relpath := strings.TrimPrefix(inject.Path, "/")
        abspath := filepath.Join(rootfs, relpath)

        // Check for relative path attacks (e.g. ../../etc/passwd)
        rel, err := filepath.Rel(rootfs, abspath)
        if err != nil {
            return err
        }
        if rel != relpath {
            return fmt.Errorf("Invalid file injection path")
        }

        if err := os.MkdirAll(filepath.Dir(abspath), 0700); err != nil {
            return err
        }

        mode := inject.Mode
        if mode == 0 {
            mode = 0700
        }

        f, err := os.OpenFile(abspath, os.O_RDWR|os.O_CREATE, os.FileMode(mode))
        if err != nil {
            return err
        }
        defer f.Close()
        if _, err := f.Write(inject.Contents); err != nil {
            return err
        }
    }

    return nil
}

func buildRandomName() (string, error) {
    r := make([]byte, 16)
    n, err := io.ReadFull(rand.Reader, r)
    if err != nil {
        return "", err
    }
    if n != len(r) {
        return "", fmt.Errorf("Failed to generate random name")
    }

    return hex.EncodeToString(r), nil
}

func (s *ContainerManager) CreateContainer(config *ContainerInfo) (err error) {
    var jsonData []byte

    name := config.Key
    if name == "" {
        name, err = buildRandomName()
        if err != nil {
            return err
        }
    }

    err = gommons.CheckSafeName(name)
    if err != nil {
        return err
    }

    jsonData, err = json.Marshal(config)
    if err != nil {
        return err
    }

    image, err := s.runtime.Images.GetImageInfo(config.Image)
    if err != nil {
        return err
    }

    if image == nil {
        return fmt.Errorf("Image not found")
    }

    dir := s.GetContainersDir() + "/" + name
    rootfs := dir + "/rootfs"

    exists, err := gommons.FileExists(rootfs)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("Container with same name already exists")
    }

    err = os.MkdirAll(rootfs, 0700)
    if err != nil {
        return err
    }

    confPath := dir + "/config.json"
    err = ioutil.WriteFile(confPath, []byte(jsonData), 0600)
    if err != nil {
        return err
    }

    archivepath, err := s.runtime.Images.getArchivePath(config.Image)
    if err != nil {
        return err
    }

    err = Untar(archivepath, rootfs)
    if err != nil {
        return err
    }

    err = injectFiles(rootfs, config)
    if err != nil {
        return err
    }

    lxcConfig := config.LxcConfig.ManualConfig
    lxcConfig = strings.Replace(lxcConfig, "{{ROOTFS}}", rootfs, -1)

    lxcConfigPath := dir + "/config.lxc"
    err = ioutil.WriteFile(lxcConfigPath, []byte(lxcConfig), 0600)
    if err != nil {
        return err
    }

    return nil
}

func (s *ContainerManager) DeleteContainer(name string) (err error) {
    err = gommons.CheckSafeName(name)
    if err != nil {
        return err
    }

    return fmt.Errorf("Delete container not yet supported")

    //	confPath := s.GetProcessesDir() + "/" + name + ".conf"
    //	err = SafeDelete(confPath)
    //	if err != nil {
    //		return
    //	}
    //	return nil
}

func (s *ContainerManager) StartContainer(name string) (err error) {
    container, err := s.GetContainerInfo(name)
    if err != nil {
        return err
    }

    proc := &processes.WatchedProcessConfig{}
    proc.Name = "/usr/bin/lxc-start"

    dir := s.GetContainersDir() + "/" + name

    args := []string{}
    args = append(args, "-n", container.Key)
    args = append(args, "-f", dir+"/config.lxc")

    proc.Args = args
    proc.Dir = dir

    err = s.runtime.Processes.WriteProcess("vm-"+container.Key, proc)
    if err != nil {
        return err
    }

    return nil
}

func (s *ContainerManager) StopContainer(name string) (err error) {
    container, err := s.GetContainerInfo(name)
    if err != nil {
        return err
    }

    err = s.runtime.Processes.DeleteProcess("vm-" + container.Key)
    if err != nil {
        return err
    }

    return nil
}
