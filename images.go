package gawker

import (
    "fmt"
    "github.com/fathomdb/gommons"
    "log"
)

type ImageManager struct {
    runtime *Runtime
}

type ImageInfo struct {
}

func NewImageManager(runtime *Runtime) (s *ImageManager) {
    s = &ImageManager{}
    s.runtime = runtime
    return s
}

func (s *ImageManager) GetImagesDir() string {
    return s.runtime.WorkDir + "/images"
}

func (s *ImageManager) GetImageInfo(key string) (*ImageInfo, error) {
    err := gommons.CheckSafeName(key)
    if err != nil {
        return nil, err
    }

    infopath := s.GetImagesDir() + "/" + key + ".json"

    info, err := readImageInfo(infopath)
    if err != nil {
        log.Printf("Error reading image info %s %v", infopath, err)
        return nil, err
    }

    return info, nil
}

func (s *ImageManager) getArchivePath(key string) (string, error) {
    err := gommons.CheckSafeName(key)
    if err != nil {
        return "", err
    }

    path := s.GetImagesDir() + "/" + key + ".tar.gz"

    exists, err := gommons.FileExists(path)
    if err != nil {
        return "", err
    }

    if exists {
        return path, nil
    }

    return "", fmt.Errorf("Image archive not found")
}

func readImageInfo(path string) (*ImageInfo, error) {
    c := &ImageInfo{}
    err := gommons.ReadJson(path, c)
    if err != nil {
        return nil, err
    }
    return c, nil
}
