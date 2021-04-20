#! /usr/bin/bash

# Test each package in order of dependencies
(cd maildb; ./test_all.sh)
(cd cmd; ./test_all.sh)
