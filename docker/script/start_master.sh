#!/bin/sh
rm -rf /cfs/data* /cfs/log/*
mkdir -p /cfs/log /cfs/data/wal /cfs/data/store /cfs/bin
echo "start master"
/home/mrsun/go/src/chubaofs/chubaofs/chubaofs/build/bin/cfs-server -f -c /home/mrsun/go/src/chubaofs/chubaofs/chubaofs/cmd/cfg/master.json > output


#/cfs/bin/cfs-server -f -c /cfs/conf/master.json > output
