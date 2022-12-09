package main

import (
	"io/ioutil"
	"log"
)

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func RSF(path string) string {
	b, err := ioutil.ReadFile(path)
	check(err)
	return string(b)
}
