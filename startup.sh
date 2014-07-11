#!/bin/bash

mkdir -p $LOG_DIR

./omg-search -token $1 > $LOG_DIR/server.log