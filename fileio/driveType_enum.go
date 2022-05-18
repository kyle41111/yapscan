// Code generated by go-enum DO NOT EDIT.
// Version: 0.4.1
// Revision: 2485d8f8373fccc23f5de42b4b3215ae3c9aef03
// Build Date: 2022-05-16T21:15:57Z
// Built By: goreleaser

package fileio

import (
	"fmt"
	"strings"
)

const (
	// DriveTypeUnknown is a DriveType of type Unknown.
	DriveTypeUnknown DriveType = iota
	// DriveTypeRemovable is a DriveType of type Removable.
	DriveTypeRemovable
	// DriveTypeFixed is a DriveType of type Fixed.
	DriveTypeFixed
	// DriveTypeRemote is a DriveType of type Remote.
	DriveTypeRemote DriveType = iota + 1
	// DriveTypeCDRom is a DriveType of type CDRom.
	DriveTypeCDRom DriveType = iota + 4
	// DriveTypeRAM is a DriveType of type RAM.
	DriveTypeRAM DriveType = iota + 11
)

const _DriveTypeName = "UnknownRemovableFixedRemoteCDRomRAM"

var _DriveTypeNames = []string{
	_DriveTypeName[0:7],
	_DriveTypeName[7:16],
	_DriveTypeName[16:21],
	_DriveTypeName[21:27],
	_DriveTypeName[27:32],
	_DriveTypeName[32:35],
}

// DriveTypeNames returns a list of possible string values of DriveType.
func DriveTypeNames() []string {
	tmp := make([]string, len(_DriveTypeNames))
	copy(tmp, _DriveTypeNames)
	return tmp
}

var _DriveTypeMap = map[DriveType]string{
	DriveTypeUnknown:   _DriveTypeName[0:7],
	DriveTypeRemovable: _DriveTypeName[7:16],
	DriveTypeFixed:     _DriveTypeName[16:21],
	DriveTypeRemote:    _DriveTypeName[21:27],
	DriveTypeCDRom:     _DriveTypeName[27:32],
	DriveTypeRAM:       _DriveTypeName[32:35],
}

// String implements the Stringer interface.
func (x DriveType) String() string {
	if str, ok := _DriveTypeMap[x]; ok {
		return str
	}
	return fmt.Sprintf("DriveType(%d)", x)
}

var _DriveTypeValue = map[string]DriveType{
	_DriveTypeName[0:7]:                    DriveTypeUnknown,
	strings.ToLower(_DriveTypeName[0:7]):   DriveTypeUnknown,
	_DriveTypeName[7:16]:                   DriveTypeRemovable,
	strings.ToLower(_DriveTypeName[7:16]):  DriveTypeRemovable,
	_DriveTypeName[16:21]:                  DriveTypeFixed,
	strings.ToLower(_DriveTypeName[16:21]): DriveTypeFixed,
	_DriveTypeName[21:27]:                  DriveTypeRemote,
	strings.ToLower(_DriveTypeName[21:27]): DriveTypeRemote,
	_DriveTypeName[27:32]:                  DriveTypeCDRom,
	strings.ToLower(_DriveTypeName[27:32]): DriveTypeCDRom,
	_DriveTypeName[32:35]:                  DriveTypeRAM,
	strings.ToLower(_DriveTypeName[32:35]): DriveTypeRAM,
}

// ParseDriveType attempts to convert a string to a DriveType.
func ParseDriveType(name string) (DriveType, error) {
	if x, ok := _DriveTypeValue[name]; ok {
		return x, nil
	}
	return DriveType(0), fmt.Errorf("%s is not a valid DriveType, try [%s]", name, strings.Join(_DriveTypeNames, ", "))
}

// MarshalText implements the text marshaller method.
func (x DriveType) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *DriveType) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseDriveType(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}
