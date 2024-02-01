#!/bin/bash

PROJECT_ROOT_DIR=path-to-nexster-backend-repo
LOGS_DIR=path-to-log-directory

cd $PROJECT_ROOT_DIR

if [ ! -d "$PROJECT_ROOT_DIR/logs" ]; then
    mkdir logs
    echo "logs directory created."
else
    echo "logs directory already exists."
fi

# Run content server
cd content/cmd
nohup go run main.go > $LOGS_DIR/content_server.log 2>&1 &

# Run search server
cd ../../search/cmd
nohup go run main.go > $LOGS_DIR/search_server.log 2>&1 &

# Run space server
cd ../../space/cmd
nohup go run main.go > $LOGS_DIR/space_server.log 2>&1 &

# Run timeline server
cd ../../timeline/cmd
nohup go run main.go > $LOGS_DIR/timeline_server.log 2>&1 &

# Run usrmgmt server
cd ../../usrmgmt/cmd
nohup go run main.go > $LOGS_DIR/usrmgmt_server.log 2>&1 &

cd ../../boarding_finder/cmd
nohup go run main.go > $LOGS_DIR/bdFinder_server.log 2>&1 &

echo "--done--"
