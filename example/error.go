package main

import "github.com/SpirentOrion/luddite"

const (
	// Service's error codes
	EcodeUserExists = luddite.EcodeServiceBase
)

var errorDefs = map[int]luddite.ErrorDefinition{
	EcodeUserExists: {"USER_ALREADY_EXISTS", "User already exists: %s"},
}
