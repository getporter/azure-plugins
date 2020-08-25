#!/bin/bash
set -m
OUTPUT=$(tempfile)
set +e
~/.porter/porter install "$@" > ${OUTPUT} 2>&1 &
echo $! >/tmp/porter-debug.pid
until (tail -f ${OUTPUT} & TAILPID=$! ; sleep 3; kill -9 ${TAILPID}) | grep "API server listening at: 127.0.0.1"; do echo dlv server not running after 3 seconds; cat ${OUTPUT}; exit 1; done
exit 0