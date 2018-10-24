#!/bin/bash
go build -buildmode=plugin matcher.go
cd notifier_plugins/mailsender
./build.sh
