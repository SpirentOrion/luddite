package datastore

import "fmt"

type LockError struct {
	Item string
}

func (e *LockError) Error() string {
	return fmt.Sprintf("%s is locked", e.Item)
}

type UpdatePreemptionError struct {
	OldSerial int64
	NewSerial int64
}

func (e *UpdatePreemptionError) Error() string {
	return fmt.Sprintf("%d subsequent update(s) have occurred", e.NewSerial-e.OldSerial)
}

type ValidationError struct {
	ErrorString string
}

func (e *ValidationError) Error() string {
	return e.ErrorString
}
