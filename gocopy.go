package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

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

func getFileList(path string) {
	if !isDirectory(path) {
		fmt.Println(path, "is File.")
		os.Exit(1)
	}

	fileList, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	for i := range fileList {
		fullpath := filepath.Join(path, fileList[i].Name())
		if fileList[i].IsDir() {
			getFileList(fullpath)
		}
		fmt.Println(fullpath)
	}
}

func copyFile() {
	/*
		content, err := ioutil.ReadFile(flag.Arg(0))
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(flag.Arg(1), content, 644)
		if err != nil {
			panic(err)
		}
	*/
}

func main() {
	flag.Parse()

	srcPath := flag.Arg(0)
	dstPath := flag.Arg(1)
	fmt.Println("arg1: ", srcPath)
	fmt.Println("arg2: ", dstPath)

	if !isExist(srcPath) {
		fmt.Println("Source file or directory not found.")
		os.Exit(1)
	}

	if !isDirectory(srcPath) {
		// source is File
		if isExist(dstPath) {
			// destnation is exist
			if isDirectory(dstPath) {
				// copy destination directory
				fmt.Println("dst is directory.")
			} else {
				// overwrite destination file
				fmt.Println("dst is file.")
			}
		} else {
			// destnation is not exist
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
		// source is Directory
		getFileList(srcPath)
	}
}
