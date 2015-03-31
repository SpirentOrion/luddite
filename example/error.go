package main

import "github.com/SpirentOrion/luddite"

const (
	// Service's error codes
	EcodeUserExists = luddite.EcodeServiceBase
)

var errorMessages = map[int]string{
	EcodeUserExists: "User already exists: %s",
}
