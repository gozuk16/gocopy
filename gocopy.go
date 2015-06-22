package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"runtime"
)

const File = 1
const Dir = 2
const Err = 9

type CopyFileList struct {
	fileType int
	srcFile  string
	dstFile  string
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println("Error: ", err)
		return false
	}
	if fileInfo.IsDir() {
		return true
	}
	return false
}

func logCopyFile(i int, list []CopyFileList) {
	fmt.Printf("%3d %s\n", i, list[i].srcFile)
	fmt.Printf("%d   %s\n", list[i].fileType, list[i].dstFile)
}

func getFileList(srcPath, dstPath string, list []CopyFileList) []CopyFileList {
	if !isDirectory(srcPath) {
		fmt.Println(srcPath, "is File.")
		// エラーはskip(Access is denied. になる場合がある)
		list[len(list)-1].fileType = Err
		logCopyFile(len(list)-1, list)
		//os.Exit(1)
		return list
	}

	fileList, err := ioutil.ReadDir(srcPath)
	if err != nil {
		fmt.Println("Error: ", err)
		// エラーはskip(Access is denied. になる場合がある)
		list[len(list)-1].fileType = Err
		logCopyFile(len(list)-1, list)
		//os.Exit(1)
		return list
	}
	for i := range fileList {
		fullSrcPath := filepath.Join(srcPath, fileList[i].Name())
		fullDstPath := filepath.Join(dstPath, fileList[i].Name())
		if fileList[i].IsDir() {
			list = append(list, CopyFileList{Dir, fullSrcPath, fullDstPath})
			list = getFileList(fullSrcPath, fullDstPath, list)
		} else {
			list = append(list, CopyFileList{File, fullSrcPath, fullDstPath})
		}
	}
	return list
}

func copyFile(srcFile, dstFile string) {
	/*
		content, err := ioutil.ReadFile(srcFile)
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(dstFile, content, 644)
		if err != nil {
			panic(err)
		}
	*/
	src, err := os.Open(srcFile)
	if err != nil {
		panic(err)
	}
	defer src.Close()

	dst, err := os.Create(dstFile)
	if err != nil {
		panic(err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	srcPath := flag.Arg(0)
	dstPath := flag.Arg(1)
	fmt.Println("arg1: ", srcPath)
	fmt.Println("arg2: ", dstPath)

	var wg sync.WaitGroup

	if !isExist(srcPath) {
		fmt.Println("Source file or directory not found.")
		os.Exit(1)
	}

	if !isDirectory(srcPath) {
		// source is File
		if isExist(dstPath) {
			// destination is exist
			if isDirectory(dstPath) {
				// copy destination directory
				fmt.Println("dst is directory.")
				_, file := filepath.Split(srcPath)
				fmt.Println("copy ", srcPath, " to ", filepath.Join(dstPath, file))
				copyFile(srcPath, filepath.Join(dstPath, file))
			} else {
				// overwrite destination file
				fmt.Println("dst is file.")
			}
		} else {
			// destination is not exist
			dir, file := filepath.Split(dstPath)
			fmt.Println(dir)
			fmt.Println(file)
			if isDirectory(dir) {
				// copy destination file in directory
				fmt.Println("dst is newfile in directory")
			} else {
				// error
				fmt.Println("dst is error")
				os.Exit(1)
			}
		}
	} else {
		list := []CopyFileList{}
		// source is Directory
		if isExist(dstPath) {
			// destination is exist
			if isDirectory(dstPath) {
				// copy destination directory
				fmt.Println("dst is directory.")
				list = getFileList(srcPath, dstPath, list)
			} else {
				// overwrite destination file
				fmt.Println("dst is file.")
				os.Exit(1)
			}
		} else {
			// destnation is not exist
			list = getFileList(srcPath, dstPath, list)
		}

		cpus := runtime.NumCPU()
		runtime.GOMAXPROCS(cpus)
		fmt.Println("cpus:", cpus)
		for _, target := range list {
			//logCopyFile(i, list)
			//fmt.Printf("%d ", i)
			if target.fileType == Dir {
				os.MkdirAll(target.dstFile, 0777)
			} else if target.fileType == File {
				wg.Add(1)
				go func(copyTarget CopyFileList) {
					defer wg.Done()
					copyFile(copyTarget.srcFile, target.dstFile)
				}(target)
				wg.Wait()
			}
		}
	}
}
