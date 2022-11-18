#!/bin/bash

CGO_ENABLED=0 GOOS=darwin go build -a -installsuffix cgo -o main .