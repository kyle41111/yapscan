// Code generated by go-enum DO NOT EDIT.
// Version: 0.4.1
// Revision: 2485d8f8373fccc23f5de42b4b3215ae3c9aef03
// Build Date: 2022-05-16T21:15:57Z
// Built By: goreleaser

package procio

import (
	"fmt"
	"strings"
)

const (
	// CrashMethodCreateThreadOnNull is a CrashMethod of type CreateThreadOnNull.
	CrashMethodCreateThreadOnNull CrashMethod = iota
)

const _CrashMethodName = "createThreadOnNull"

var _CrashMethodNames = []string{
	_CrashMethodName[0:18],
}

// CrashMethodNames returns a list of possible string values of CrashMethod.
func CrashMethodNames() []string {
	tmp := make([]string, len(_CrashMethodNames))
	copy(tmp, _CrashMethodNames)
	return tmp
}

var _CrashMethodMap = map[CrashMethod]string{
	CrashMethodCreateThreadOnNull: _CrashMethodName[0:18],
}

// String implements the Stringer interface.
func (x CrashMethod) String() string {
	if str, ok := _CrashMethodMap[x]; ok {
		return str
	}
	return fmt.Sprintf("CrashMethod(%d)", x)
}

var _CrashMethodValue = map[string]CrashMethod{
	_CrashMethodName[0:18]:                  CrashMethodCreateThreadOnNull,
	strings.ToLower(_CrashMethodName[0:18]): CrashMethodCreateThreadOnNull,
}

// ParseCrashMethod attempts to convert a string to a CrashMethod.
func ParseCrashMethod(name string) (CrashMethod, error) {
	if x, ok := _CrashMethodValue[name]; ok {
		return x, nil
	}
	return CrashMethod(0), fmt.Errorf("%s is not a valid CrashMethod, try [%s]", name, strings.Join(_CrashMethodNames, ", "))
}

// MarshalText implements the text marshaller method.
func (x CrashMethod) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *CrashMethod) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseCrashMethod(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}
