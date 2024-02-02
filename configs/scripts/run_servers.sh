#!/bin/bash

PROJECT_ROOT_DIR=absolute-path-to-project-root-directory

cd $PROJECT_ROOT_DIR

if [ ! -d "$PROJECT_ROOT_DIR/logs" ]; then
    mkdir logs
    echo "logs directory created."
else
    echo "logs directory already exists."
fi

# Run content server
cd content/cmd
nohup go run main.go > $PROJECT_ROOT_DIR/logs/content_server.log 2>&1 &

# Run search server
cd ../../search/cmd
nohup go run main.go > $PROJECT_ROOT_DIR/logs/search_server.log 2>&1 &

# Run space server
cd ../../space/cmd
nohup go run main.go > $PROJECT_ROOT_DIR/logs/space_server.log 2>&1 &

# Run timeline server
cd ../../timeline/cmd
nohup go run main.go > $PROJECT_ROOT_DIR/logs/timeline_server.log 2>&1 &

# Run usrmgmt server
cd ../../usrmgmt/cmd
nohup go run main.go > $PROJECT_ROOT_DIR/logs/usrmgmt_server.log 2>&1 &

cd ../../boarding_finder/cmd
nohup go run main.go > $PROJECT_ROOT_DIR/logs/bdFinder_server.log 2>&1 &

echo "--done--"
