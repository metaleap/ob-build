package main

import (
	"flag"

	ugo "github.com/metaleap/go-util"
	uio "github.com/metaleap/go-util/io"
	ustr "github.com/metaleap/go-util/str"
)

func copyHive(dst string) {
	var skipDirs ustr.Matcher
	skipDirs.AddPatterns("cust")

	src := ugo.GopathSrcGithub("openbase", "ob-build", "default-hive")

	if len(dst) > 0 {
		uio.CopyAll(src, dst, &skipDirs)
	}

	dst = ugo.GopathSrcGithub("openbase", "ob-gae", "demo-app", "hive")
	uio.CopyAll(src, dst, &skipDirs)
}

func main() {
	dst := flag.String("hive_dst", "", "Destination hive dir path to copy default-hive to")
	flag.Parse()
	copyHive(*dst)
}
