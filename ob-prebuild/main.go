//	Processes files in `hive-prep` into `hive-default`, then copies `hive-default` to specified locations.
package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-utils/ufs"
	"github.com/go-utils/ugo"
)

var (
	wait        sync.WaitGroup
	hiveDirPath = ugo.GopathSrcGithub("openbase", "ob-build", "hive-default")
)

func compilerRun(cmd string, args ...string) {
	var (
		output []byte
		err    error
	)
	if output, err = exec.Command(cmd, args...).CombinedOutput(); err != nil {
		log.Printf("[%s]\tERROR: %v\n", cmd, err)
	} else if len(output) > 0 {
		log.Println(string(output))
	}
}

func compileWebFiles() (err error) {
	prepDirPath := ugo.GopathSrcGithub("openbase", "ob-build", "hive-prep")
	prepTmpPath := filepath.Join(prepDirPath, "_tmp")
	if err = ufs.EnsureDirExists(prepTmpPath); err != nil {
		return
	}

	//	convert hive-prep/path/file.old to hive-default/path/file.new
	//	return "" if base-name of path starts with "_" to skip processing such files
	getOutFilePath := func(srcFilePath, newExt string) (outFilePath string) {
		if srcDir, srcExt, srcBase := filepath.Dir(srcFilePath), filepath.Ext(srcFilePath), filepath.Base(srcFilePath); !strings.HasPrefix(srcBase, "_") {
			outFilePath = filepath.Join(hiveDirPath, srcDir[len(prepDirPath):], srcBase[:len(srcBase)-len(srcExt)]+newExt)
		}
		return
	}

	//	is file.src newer than file.dst?
	//	outFilePath may be "" as per getOutFilePath(), then returns false to skip processing
	shouldPrep := func(srcFilePath, outFilePath string) (newer bool) {
		if len(outFilePath) > 0 {
			newer, _ = ufs.IsNewerThan(srcFilePath, outFilePath)
		}
		return
	}

	//	preprocess file.src -> file.dst
	prepFile := func(filePath string) {
		defer wait.Done()
		var outFilePath string
		switch filepath.Ext(filePath) {
		case ".scss":
			if outFilePath = getOutFilePath(filePath, ".css"); shouldPrep(filePath, outFilePath) {
				compilerRun("sass", "--trace", "--scss", "--stop-on-error", "-f", "-g", "-l", "-t", "expanded", "--cache-location", prepTmpPath, filePath, outFilePath)
			}
			if outFilePath = getOutFilePath(filePath, ".min.css"); shouldPrep(filePath, outFilePath) {
				compilerRun("sass", "--trace", "--scss", "--stop-on-error", "-f", "-t", "compressed", "--cache-location", prepTmpPath, filePath, outFilePath)
			}
		}
	}

	if errs := ufs.WalkAllFiles(prepDirPath, func(filePath string) bool {
		if !strings.HasPrefix(filePath, prepTmpPath) {
			wait.Add(1)
			go prepFile(filePath)
		}
		return true
	}); len(errs) > 0 {
		err = errs[0]
	}
	wait.Wait()
	return
}

func copyHive(dst string) {
	defer wait.Done()
	log.Printf("Copy hive to: %s", dst)
	subDirs := []string{"dist", "cust"}
	for _, subDir := range subDirs {
		ufs.EnsureDirExists(filepath.Join(dst, subDir))
		if subDir == "dist" {
			if err := ufs.ClearDirectory(filepath.Join(dst, subDir)); err != nil {
				panic(err)
			}
		}
		if err := ufs.CopyAll(filepath.Join(hiveDirPath, subDir), filepath.Join(dst, subDir), nil); err != nil {
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
