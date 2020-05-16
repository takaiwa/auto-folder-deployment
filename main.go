package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"
)

//const INTERVAL_SEC = 600 // 10min
const INTERVAL_SEC = 5 // 5sec

func copyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func copyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = copyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = copyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func main() {
	flag.Parse()
	var srcPath, dstPath string

	if len(flag.Args()) == 2 {
		srcPath = flag.Arg(0)
		dstPath = flag.Arg(1)
	} else {
		fmt.Println("Invalid parameters")
		return
	}

	t := time.NewTicker(INTERVAL_SEC * time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			files, err := ioutil.ReadDir(srcPath)
			if err != nil {
				panic(err)
			}
			for _, file := range files {
				if file.IsDir() {
					path := dstPath + "/" + file.Name()
					if Exists(path) {
						fmt.Println("Exist:", file.Name())
					} else {
						fmt.Println("copy:", file.Name(), " to:", dstPath)
						copyDir(srcPath, dstPath)
					}
				}
			}
		}
	}
}
