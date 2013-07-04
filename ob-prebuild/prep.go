package main

import (
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-utils/ufs"
	"github.com/go-utils/ugo"
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

//	is file.src newer than file.dst?
//	outFilePath may be "" as per getOutFilePath(), then returns false to skip processing
func shouldPrep(srcFilePath, outFilePath string) (newer bool) {
	if len(outFilePath) > 0 {
		newer, _ = ufs.IsNewerThan(srcFilePath, outFilePath)
	}
	return
}
