package main

import (
	"flag"
	"path/filepath"
	"runtime"
	"sync"

	ugo "github.com/metaleap/go-util"
	uio "github.com/metaleap/go-util/io"
)

var wait sync.WaitGroup

func copyHive(dst string, cust bool) {
	defer wait.Done()
	subDirs := []string{"dist"}
	if cust {
		subDirs = append(subDirs, "cust")
	}
	src := ugo.GopathSrcGithub("openbase", "ob-build", "default-hive")
	for _, subDir := range subDirs {
		if err := uio.ClearDirectory(filepath.Join(dst, subDir)); err != nil {
			panic(err)
		}
		if err := uio.CopyAll(filepath.Join(src, subDir), filepath.Join(dst, subDir), nil); err != nil {
			panic(err)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	wait.Add(1)
	go copyHive(ugo.GopathSrcGithub("openbase", "ob-gae", "demo-app", "hive"), true)
	dst := flag.String("hive_dst", "", "Destination hive dir path to copy default-hive to")
	if flag.Parse(); len(*dst) > 0 {
		wait.Add(1)
		go copyHive(*dst, false)
	}
	wait.Wait()
}
