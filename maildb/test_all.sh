#! /usr/bin/bash

go test -run=TestDecode
go test -run=TestTarget
go test -run=TestTransport
go test -run=TestDBdefaults
go test -run=TestAccess
go test -run=Test_Transport
go test -run=TestDomain
go test -run=TestAddress
go test -run=TestAliasOps
go test -run=TestMailbox
