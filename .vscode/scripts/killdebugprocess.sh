set +e
pkill -P $(cat /tmp/porter-debug.pid) 
kill $(cat /tmp/porter-debug.pid) 
exit 0