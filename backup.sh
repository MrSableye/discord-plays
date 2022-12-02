#!/bin/bash
curl http://localhost:1234/req?action=save
mkdir backup
date=$(date '+%Y-%m-%d')
cp ./build/save.sav ./backup/${date}.sav