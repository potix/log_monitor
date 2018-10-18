#!/bin/bash
protoc  logpb/log.proto --go_out=plugins=grpc:.
go build log_reciever.go
go build log_monitor.go
cd actor_plugins/matcher
./build.sh
cd ../../
cd actor_plugins/sender
./build.sh
