#!/bin/bash



if [[ -f gin ]]; then
    rm ./gin
fi

go build -tags=jsoniter -o gin ./main.go