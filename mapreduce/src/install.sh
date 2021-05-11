#!/bin/bash
echo "Begin Installing:\n"

#clear golang installation
sudo rm -rf /usr/local/go
sudo rm /usr/bin/go /usr/bin/gofmt

#install golang
wget https://studygolang.com/dl/golang/go1.16.4.linux-arm64.tar.gz
tar -C /usr/local -zxvf go1.16.4.linux-arm64.tar.gz
echo "export GOROOT=/usr/local/go" >> /etc/profile
echo "export PATH=\$PATH:\$GOROOT/bin" >> /etc/profile
source /etc/profile

ln -s /usr/local/go/bin/go /usr/bin/go

#test golang install
go version

#change $GOPROXY
export GO111MODULE=on
export GOPROXY=https://goproxy.io,direct

#install cypto/ssh
go get golang.org/x/crypto/ssh@v0.0.0-20201221181555-eec23a3978ad

#install python-dev
sudo apt-get install python3-dev
sudo apt-get install python2.7-dev

scp -r root@192.168.0.111:/root/mapreduce /root/mapreduce

echo "Finish Install."