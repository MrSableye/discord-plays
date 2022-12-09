package main

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

type BackupOption struct {
	Directory string
	HowOften  time.Duration
}

type BackupOptions struct {
	Options []BackupOption
}

var backupOptions BackupOptions

func startBackuper() {
	js := RSF("backup_options.json")
	json.Unmarshal([]byte(js), &backupOptions)
	for _, option := range backupOptions.Options {
		go backup(option.HowOften, option.Directory)
	}
}

func backup(dur time.Duration, directory string) {
	for {
		filepath := directory + "/autobackup_" + time.Now().Format("2006-01-02_15-04-05") + ".sav"
		ret := get("save")
		bs, err := ioutil.ReadAll(ret.Body)
		check(err)
		hexstring := string(bs)
		bytes, err := hex.DecodeString(hexstring)
		check(err)
		err = os.WriteFile(filepath, bytes, 0644)
		if err != nil {
			break
		}
		time.Sleep(dur)
	}
}
