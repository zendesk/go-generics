#!/bin/bash

set -e

fuzzTime=${1:-10}

files=$(grep -r --include='**_test.go' --files-with-matches 'func Fuzz' .)

for file in ${files}
do
    funcs=$(grep -o 'Fuzz\w\+' $file)
    echo $funcs
    for func in ${funcs}
    do
        echo "Fuzzing $func in $file"
        parentDir=$(dirname $file)
        echo "go test $parentDir -run=^${func}$ -fuzz=^${func}$ -fuzztime=${fuzzTime}s -parallel=3"
        go test -v $parentDir -run="^${func}\$" -fuzz="^${func}$" -fuzztime=${fuzzTime}s -parallel=3
    done
done