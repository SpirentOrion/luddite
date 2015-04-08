package main

const (
	EcodeUserExists = "USER_ALREADY_EXISTS"
)

var errorDefs = map[string]string{
	EcodeUserExists: "User already exists: %s",
}
