#!/bin/sh

trap 'echo exiting' TERM INT

echo 'sleeping infinity'
sleep infinity & wait
