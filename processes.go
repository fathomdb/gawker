package gawker

import (
    "sync"
)

type ProcessInfo struct {
    Tags map[string]interface{}
}

var pidTags map[int]*ProcessInfo = make(map[int]*ProcessInfo)
var pidTagsLock sync.Mutex

func GetPidTags(pid int) (tags map[string]interface{}) {
    pidTagsLock.Lock()
    item := pidTags[pid]
    if item != nil {
        tags = item.Tags
    }
    pidTagsLock.Unlock()
    return tags
}
