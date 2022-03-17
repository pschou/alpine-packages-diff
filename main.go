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
)

var version = "test"
var indexFileName = "APKINDEX.tar.gz"

func UNUSED(unused ...interface{}) {}

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

	oldSuffix := flag.String("suffix", ".old", "Suffix used for the old Index")
	indexRootPath := flag.String("irp", ".", "Root Path of indexes repo will be appended")
	inRepoPath := flag.String("repo", "/latest-stable/main/x86_64", "Repo path to use in file list")
	outputFile := flag.String("output", "-", "Output for comparison result")

	flag.Parse()

	apk_index := make(map[string]apk_info)

	if fileExists(*indexRootPath + "/" + *inRepoPath + "/" + indexFileName + *oldSuffix) {
		loadIndex(*indexRootPath+"/"+*inRepoPath+"/"+indexFileName+*oldSuffix, apk_index)
	}

	loadIndex(*indexRootPath+"/"+*inRepoPath+"/"+indexFileName, apk_index)

	out := os.Stdout

	if *outputFile != "-" {
		f, err := os.Create(*outputFile)
		check(err)

		defer f.Close()

		out = f
	}

	for _, v := range apk_index {
		fmt.Fprintf(out, "{apline}%s %d %s\n", v.hash, v.size, path.Join(*inRepoPath, v.name+"-"+v.version+".apk"))
	}

	UNUSED(oldSuffix, indexRootPath, inRepoPath, outputFile, apk_index, out)
}

func loadIndex(indexPath string, apk_index map[string]apk_info) {
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
			line = indexData.Text()
		}

		if val, ok := apk_index[pkgInfo.name]; ok {
			if val.hash == pkgInfo.hash {
				delete(apk_index, pkgInfo.name)

				continue
			}
		}

		apk_index[pkgInfo.name] = pkgInfo
	}

	fmt.Printf("Len %d\n", len(apk_index))
}

func split_on_gzip_header(data []byte) ([]byte, []byte) {
	gzip_header := []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00}
	arr := []byte(gzip_header)

	pos := int64(len(gzip_header))

	has_sig := false
	for !has_sig {
		if !(bytes.Equal(data[pos:pos+int64(len(gzip_header))], gzip_header)) {
			arr = append(arr, data[pos])

			pos += 1
		} else {
			has_sig = true
		}
	}

	sig := data[:pos]
	content := data[pos:]

	return sig, content
}

// isDirectory determines if a file represented
// by `path` is a directory or not
func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
