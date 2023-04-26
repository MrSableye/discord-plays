package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func RSF(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func GetAbsolutePath() string {
	var ret string
	for {
		fmt.Scan(&ret)
		abs := filepath.IsAbs(ret)
		exists := FileExists(ret)
		if !exists {
			fmt.Println("File does not exist. Please enter a valid path.")
		}
		if abs && exists {
			break
		}
		fmt.Println("Invalid path. Please enter an absolute path, not a relative path.")
	}
	return ret
}

func GetNumber(def int) int {
	var num string
	for {
		fmt.Scan(&num)
		if num == "" {
			fmt.Println("Using default value: " + strconv.Itoa(def) + ".")
			return def
		}
		ret, err := strconv.Atoi(num)
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
		} else {
			return ret
		}
	}
}
