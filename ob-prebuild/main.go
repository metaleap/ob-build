//	Processes files in `hive-prep` into `hive-default`, then copies `hive-default` to specified locations.
package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/go-utils/ufs"
	"github.com/go-utils/ugo"
)

var (
	wait        sync.WaitGroup
	hiveDirPath = ugo.GopathSrcGithub("openbase", "ob-build", "hive-default")
)

func copyHive(dst string) {
	defer wait.Done()
	log.Printf("Copy hive to: %s", dst)
	var err error
	var subDst string
	for _, subDir := range []string{"dist", "cust"} {
		subDst = filepath.Join(dst, subDir)
		ufs.EnsureDirExists(subDst)
		switch subDir {
		case "dist":
			err = ufs.ClearDirectory(subDst)
		case "cust":
			_, err = ufs.ClearEmptyDirectories(subDst)
		default:
			err = fmt.Errorf("TODO: update copyHive() in ob-prebuild/main.go")
		}
		if err == nil {
			err = ufs.CopyAll(filepath.Join(hiveDirPath, subDir), subDst, nil)
		}
		if err != nil {
			panic(err)
		}
	}
}

func resetCust() {
	distDirPath, custDirPath := filepath.Join(hiveDirPath, "dist"), filepath.Join(hiveDirPath, "cust")
	//	clear everything in cust
	ufs.ClearDirectory(custDirPath)
	//	recreate entire dist directory hierarchy in cust, but empty (no files, only directories)
	ufs.WalkAllDirs(distDirPath, func(dirPath string) bool {
		ufs.EnsureDirExists(filepath.Join(custDirPath, dirPath[len(distDirPath):]))
		return true
	})
}

func main() {
	ugo.MaxProcs()
	resetCust()
	err := compileWebFiles()
	if err != nil {
		panic(err)
	}

	//	copy to GAE demo-app/hive
	wait.Add(1)
	go copyHive(ugo.GopathSrcGithub("openbase", "ob-gae", "demo-app", "hive"))

	//	copy to other user-specified hive directory such as "/user/foo/my-ob-dev/test2"
	dst := flag.String("hive_dst", "", fmt.Sprintf("Destination hive dir path to copy %s to", hiveDirPath))
	if flag.Parse(); len(*dst) > 0 {
		wait.Add(1)
		go copyHive(*dst)
	}

	wait.Wait()
}
