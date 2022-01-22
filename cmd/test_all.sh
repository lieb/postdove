#! /usr/bin/bash

go test -run=Test_Import
go test -run=Test_Cmds
go test -run=Test_Domain
go test -run=Test_Address
go test -run=TestAliasCmds
go test -run=TestVMailboxCmd
go test -run=Test_Create
go test -run=TestCreateNoAliases
