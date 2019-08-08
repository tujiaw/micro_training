#!/bin/bash
pid=`pgrep -f "micro api --address 0.0.0.0:3100"`
kill -9 $pid