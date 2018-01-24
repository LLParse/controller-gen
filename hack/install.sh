#!/bin/bash

gofmt -w *.go args/*.go generators/*.go
go install

