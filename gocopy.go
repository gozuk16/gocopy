package main

import (
	"flag"
	"fmt"
	"log"
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
	log.Printf("%3d %s\n", i, list[i].srcFile)
	log.Printf("%d   %s\n", list[i].fileType, list[i].dstFile)
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

	var wg sync.WaitGroup
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)
	for _, target := range fileList {
		fullSrcPath := filepath.Join(srcPath, target.Name())
		fullDstPath := filepath.Join(dstPath, target.Name())
		if target.IsDir() {
			os.MkdirAll(fullDstPath, 0777)
			//list = append(list, CopyFileList{Dir, fullSrcPath, fullDstPath})
			list = getFileList(fullSrcPath, fullDstPath, list)
		} else {
			content := readFile(fullSrcPath)
			wg.Add(1)
			go func(dst string, b []byte) {
				defer wg.Done()
				writeFile(b, dst)
			}(fullDstPath, content)
			wg.Wait()
			//copyFile(fullSrcPath, fullDstPath)
			//list = append(list, CopyFileList{File, fullSrcPath, fullDstPath})
		}
	}
	return list
}

func copyFile(srcFile, dstFile string) {
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

func readFile(src string) []byte {
	content, err := ioutil.ReadFile(src)
	if err != nil {
		panic(err)
	}
	return content
}

func writeFile(content []byte, dst string) {
	err := ioutil.WriteFile(dst, content, 644)
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	srcPath := flag.Arg(0)
	dstPath := flag.Arg(1)
	log.Println("arg1: ", srcPath)
	log.Println("arg2: ", dstPath)

	//var wg sync.WaitGroup

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
				log.Println("dst is directory.")
				list = getFileList(srcPath, dstPath, list)
				//log.Println("finish getFileList: size =", len(list))
			} else {
				// overwrite destination file
				fmt.Println("dst is file.")
				os.Exit(1)
			}
		} else {
			// destnation is not exist
			list = getFileList(srcPath, dstPath, list)
		}
	}
	log.Println("finished")
}
