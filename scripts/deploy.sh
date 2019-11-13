#!/bin/bash
echo "compiling"
go build -ldflags="-s -w"

echo "start ssh agent"
eval `ssh-agent`
ssh-add

echo "stopping service"
ssh root@pipeto.me 'systemctl stop pipe-to-me'
echo "copying files"
scp ./pipe-to-me root@pipeto.me:/root/data/pipe-to-me/pipe-to-me
echo "starting service"
ssh root@pipeto.me 'systemctl start pipe-to-me'

echo "stop ssh agent"
killall ssh-agent
