#!/bin/bash
curl http://localhost:1234/req?action=save
mkdir -p backup
date=$(date '+%Y-%m-%d_%H-%M-%S')
# TODO: ./build/save.sav is hardcoded
cp ./build/save.sav ./backup/${date}.sav