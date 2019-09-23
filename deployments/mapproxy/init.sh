#!/bin/bash

test -f mapproxy.yaml || mapproxy-util create -t base-config $PWD
mapproxy-util serve-develop mapproxy.yaml -b 0.0.0.0:8080