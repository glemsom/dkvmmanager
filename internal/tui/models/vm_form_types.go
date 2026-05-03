package models

import (
	"regexp"
)

// FormMode determines whether the form creates or edits a VM
type FormMode int

const (
	FormCreate FormMode = iota
	FormEdit
)

// nameRegex is used for VM name validation
var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9 _-]+$`)

// macRegex is used for MAC address validation
var macRegex = regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`)