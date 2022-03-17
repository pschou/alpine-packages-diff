package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

var version = "test"
var indexFileName = "APKINDEX.tar.gz"

type apk_info struct {
	hash    string
	name    string
	size    int
	version string
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Alpine Index Diff, Version: %s\n\nUsage: %s [options...]\n\n", version, os.Args[0])
		flag.PrintDefaults()
	}

	var newFile = flag.String("new", "NEW_APKINDEX.tar.gz", "The newer APKINDEX.tar.gz file or repodata/ dir for comparison")
	var oldFile = flag.String("old", "OLD_APKINDEX.tar.gz", "The older APKINDEX.tar.gz file or repodata/ dir for comparison")
	var inRepoPath = flag.String("repo", "/latest-stable/main/x86_64", "Repo path to use in file list")
	var outputFile = flag.String("output", "-", "Output file for comparison result")
	var showNew = flag.Bool("showAdded", false, "Display packages only in the new list")
	var showOld = flag.Bool("showRemoved", false, "Display packages only in the old list")
	var showCommon = flag.Bool("showCommon", false, "Display packages in both the new and old lists")

	flag.Parse()

	var new_apk_index = []apk_info{}
	var old_apk_index = []apk_info{}

	if exist, isdir := isDirectory(*newFile); exist {
		if isdir {
			*newFile = path.Join(*newFile, indexFileName)
		}
		new_apk_index = loadIndex(*newFile)
	}
	if exist, isdir := isDirectory(*oldFile); exist {
		if isdir {
			*oldFile = path.Join(*oldFile, indexFileName)
		}
		old_apk_index = loadIndex(*oldFile)
	}

	out := os.Stdout

	if *outputFile != "-" {
		f, err := os.Create(*outputFile)
		check(err)

		defer f.Close()
		out = f
	}

	// initialized with zeros
	newMatched := make([]int8, len(new_apk_index))
	oldMatched := make([]int8, len(old_apk_index))

	log.Println("doing matchups")
matchups:
	for iNew, pNew := range new_apk_index {
		for iOld, pOld := range old_apk_index {
			//if reflect.DeepEqual(pNew, pOld) {
			if pNew.hash == pOld.hash &&
				pNew.name == pOld.name &&
				pNew.size == pOld.size &&
				pNew.version == pOld.version {
				newMatched[iNew] = 1
				oldMatched[iOld] = 1
				continue matchups
			}
		}
	}

	fmt.Fprintln(out, "# Alpine-diff matchup, version:", version)
	fmt.Fprintln(out, "# new:", *newFile, "old:", *oldFile)

	if *showNew {
		for iNew, v := range new_apk_index {
			if newMatched[iNew] == 0 {
				// This package was not seen in OLD
				fmt.Fprintf(out, "{apline}%s %d %s\n", v.hash, v.size, path.Join(*inRepoPath, v.name+"-"+v.version+".apk"))
			}
		}
	}

	if *showCommon {
		for iNew, v := range new_apk_index {
			if newMatched[iNew] == 1 {
				// This package was seen in BOTH
				fmt.Fprintf(out, "{apline}%s %d %s\n", v.hash, v.size, path.Join(*inRepoPath, v.name+"-"+v.version+".apk"))
			}
		}
	}

	if *showOld {
		for iOld, v := range old_apk_index {
			if oldMatched[iOld] == 0 {
				// This package was not seen in NEW
				fmt.Fprintf(out, "{apline}%s %d %s\n", v.hash, v.size, path.Join(*inRepoPath, v.name+"-"+v.version+".apk"))
			}
		}
	}
}

func loadIndex(indexPath string) (apk_index []apk_info) {
	apk_index = []apk_info{}
	fd, err := os.Open(indexPath)
	check(err)

	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	check(err)

	_, zindex := split_on_gzip_header(data)

	zbuf := bytes.NewBuffer(zindex)
	gzr, err := gzip.NewReader(zbuf)
	check(err)

	tindex, err := ioutil.ReadAll(gzr)
	check(err)

	tbuf := bytes.NewBuffer(tindex)
	tr := tar.NewReader(tbuf)

	h, err := tr.Next()
	check(err)

	if h.Name != "APKINDEX" {
		_, err = tr.Next()
		check(err)
	}

	indexData := bufio.NewScanner(tr)

	for indexData.Scan() {
		var pkgInfo apk_info

		line := indexData.Text()

		// C:Q1oHg4kAnVFve7dHe30IKgyaCykSg=
		// P:postfix-openrc
		// V:3.6.4-r0
		// S:2518

		for line != "" {
			field := line[0:1]
			value := line[2:]

			switch field {
			case "C":
				pkgInfo.hash = value
			case "P":
				pkgInfo.name = value
			case "S":
				pkgInfo.size, err = strconv.Atoi(value)
				check(err)
			case "V":
				pkgInfo.version = value
			}

			indexData.Scan()
			line = strings.TrimSpace(indexData.Text())
		}

		apk_index = append(apk_index, pkgInfo)

	}

	return
}

func split_on_gzip_header(data []byte) ([]byte, []byte) {
	gzip_header := []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00}
	//arr := []byte(gzip_header)

	pos := bytes.Index(data[1:], gzip_header) + 1
	if pos < 1 {
		return []byte{}, data
	}
	return data[:pos], data[pos:]
}

// isDirectory determines if a file represented
// by `path` is a directory or not
func isDirectory(path string) (exist bool, isdir bool) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, false
	}
	return true, fileInfo.IsDir()
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
