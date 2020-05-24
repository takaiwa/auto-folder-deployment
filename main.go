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
const IntervalSec = 5 // 5sec

type StatusCode int

const (
	StatusInit   StatusCode = 1
	StatusCopied StatusCode = 2
	StatusError  StatusCode = 99
)

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

	if _, err = io.Copy(dstfd, srcfd); err == nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func copyDirSrcToDst(src string, dst string) error {
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
			if err = copyDirSrcToDst(srcfp, dstfp); err != nil {
				fmt.Println(err)
				return err
			}
		} else {
			if err = copyFile(srcfp, dstfp); err != nil {
				//fmt.Println("copyfile ERROR:",err)
				return err
			}
		}
	}
	return err
}
func copyDir(src string, dst string, folders map[string]StatusCode, folderName string) error {
	err := copyDirSrcToDst(src, dst)
	if err == nil {
		// リトライしなくていい
		folders[folderName] = StatusCopied
		fmt.Println("copied:", folderName, " to:", dst)
	} else {
		fmt.Println("copyDir ERROR:", err)
		// リトライが必要
		folders[folderName] = StatusError
	}
	return err
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func isCopied(src string, dst string) bool {
	var ret bool

	ret = Exists(dst)
	if ret {
		srcfiles, _ := ioutil.ReadDir(src)
		destfiles, _ := ioutil.ReadDir(dst)
		ret = len(srcfiles) <= len(destfiles)
	}
	return ret
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

	t := time.NewTicker(IntervalSec * time.Second)
	defer t.Stop()

	folderMap := map[string]StatusCode{}

	for {
		select {
		case <-t.C:
			var err error
			folders, err := ioutil.ReadDir(srcPath)
			if err != nil {
				panic(err)
			}

			tmpMap := map[string]StatusCode{}
			for _, f := range folders {
				code, exists := folderMap[f.Name()]
				if exists {
					tmpMap[f.Name()] = code
				} else {
					tmpMap[f.Name()] = StatusInit
				}
				fmt.Println("folderName:", f.Name(), ", code:", tmpMap[f.Name()])
			}
			folderMap = tmpMap

			for folderName := range folderMap {

				srcFolderPath := srcPath + "/" + folderName
				destFolderPath := dstPath + "/" + folderName

				if isCopied(srcFolderPath, destFolderPath) {
					fmt.Println("Exist:", folderName)
					code, exists := folderMap[folderName]
					if exists {
						switch code {
						case StatusInit:
							folderMap[folderName] = StatusCopied
						case StatusError:
							//retry
							t.Stop()
							copyDir(srcFolderPath, destFolderPath, folderMap, folderName)
							t = time.NewTicker(IntervalSec * time.Second)
						default:
						}
					}
				} else {
					t.Stop()
					copyDir(srcFolderPath, destFolderPath, folderMap, folderName)
					t = time.NewTicker(IntervalSec * time.Second)
				}
			}
		}
	}
}
