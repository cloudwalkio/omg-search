#!/bin/bash

mkdir -p $LOG_DIR_SEARCH

exec ./omg-search -token $1
