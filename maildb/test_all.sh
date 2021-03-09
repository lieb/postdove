#! /usr/bin/bash

go test -run=TestDecode
go test -run=TestTarget
go test -run=TestTransport
go test -run=TestDomain
go test -run=TestDBLoad
go test -run=TestAliasOps
