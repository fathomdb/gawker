package gawker

import (
    "fmt"
    "os/exec"
)

// Borrowed from docker
// Untar reads a stream of bytes from `archive`, parses it as a tar archive,
// and unpacks it into the directory at `path`.
// The archive may be compressed with one of the following algorithgms:
//  identity (uncompressed), gzip, bzip2, xz.
// FIXME: specify behavior when target path exists vs. doesn't exist.
func Untar(archive, dest string) error {
    compression := "x"

    cmd := exec.Command("tar", "--numeric-owner", "-f", archive, "-C", dest, "-x"+compression)
    //cmd.Stdin = bufferedArchive
    // Hardcode locale environment for predictable outcome regardless of host configuration.
    //   (see https://github.com/dotcloud/docker/issues/355)
    cmd.Env = []string{"LANG=en_US.utf-8", "LC_ALL=en_US.utf-8"}
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("%s: %s", err, output)
    }
    return nil
}
