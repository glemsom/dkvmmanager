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

// focusKind describes what a focus position represents
type focusKind int

const (
	focusText focusKind = iota
	focusListItem
	focusAddBtn
	focusToggle
	focusSaveBtn
)

// focusPos is one navigable position in the form's flat list
type focusPos struct {
	kind      focusKind
	fieldName string
	listIndex int // only for focusListItem
}

// nameRegex is used for VM name validation
var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9 _-]+$`)

// macRegex is used for MAC address validation
var macRegex = regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`)
