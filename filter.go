package yapscan

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/dustin/go-humanize"
	"github.com/fkie-cad/yapscan/procio"
)

// FilterMatch contains information about the matching of a MemorySegmentFilter.
type FilterMatch struct {
	Result bool
	MSI    *procio.MemorySegmentInfo
	Reason string // Reason for filter mismatch, if Result is false
}

// MemorySegmentFilterFunc is a callback, used to filter *procio.MemorySegmentInfo
// instances.
type MemorySegmentFilterFunc func(info *procio.MemorySegmentInfo) bool

// MemorySegmentFilter describes an interface, capable of filtering
// *procio.MemorySegmentInfo instances.
type MemorySegmentFilter interface {
	Filter(info *procio.MemorySegmentInfo) *FilterMatch
	Description() string
}

type baseFilter struct {
	filter         MemorySegmentFilterFunc
	Parameter      interface{}
	reasonTemplate string
	description    string
}

func (f *baseFilter) Description() string {
	return f.description
}

func (f *baseFilter) renderReason(info *procio.MemorySegmentInfo) string {
	t := template.New("filterReason")

	t.Funcs(template.FuncMap{
		"bytes": func(val interface{}) string {
			n := reflect.ValueOf(val)
			num := n.Uint()
			return humanize.Bytes(uint64(num))
		},
		"join": func(glue string, slice interface{}) string {
			s := reflect.ValueOf(slice)
			if s.Kind() != reflect.Slice {
				panic("argument is not a slice")
			}
			parts := make([]string, s.Len(), s.Len())
			for i := 0; i < s.Len(); i++ {
				str, ok := s.Index(i).Interface().(fmt.Stringer)
				if !ok {
					panic("slice does not contain implementations of the fmt.Stringer interface")
				}
				parts[i] = str.String()
			}
			return strings.Join(parts, glue)
		},
	})

	_, err := t.Parse(f.reasonTemplate)
	if err != nil {
		panic("could not parse filter reason template: " + err.Error())
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, &struct {
		Filter MemorySegmentFilter
		MSI    *procio.MemorySegmentInfo
	}{
		Filter: f,
		MSI:    info,
	})

	if err != nil {
		panic(err)
	}

	return buf.String()
}

func (f *baseFilter) Filter(info *procio.MemorySegmentInfo) *FilterMatch {
	var reasonForMismatch string

	matches := f.filter(info)
	if !matches {
		reasonForMismatch = f.renderReason(info)
	}

	return &FilterMatch{
		Result: matches,
		MSI:    info,
		Reason: reasonForMismatch,
	}
}

// NewFilterFromFunc creates a new filter from a given MemorySegmentFilterFunc.
func NewFilterFromFunc(filter MemorySegmentFilterFunc, parameter interface{}, reasonTemplate string, description string) MemorySegmentFilter {
	return &baseFilter{
		filter:         filter,
		Parameter:      parameter,
		reasonTemplate: reasonTemplate,
		description:    description,
	}
}

// NewMaxSizeFilter creates a new filter, matching *procio.MemorySegmentInfo
// with the given maximum size.
func NewMaxSizeFilter(size uintptr) MemorySegmentFilter {
	return NewFilterFromFunc(
		func(info *procio.MemorySegmentInfo) bool {
			return info.Size <= size
		},
		size,
		"segment too large, size: {{.MSI.Size|bytes}}, max-size: {{.Filter.Parameter|bytes}}",
		fmt.Sprintf("segment must be smaller than %s", humanize.Bytes(uint64(size))),
	)
}

// NewMinSizeFilter creates a new filter, matching *procio.MemorySegmentInfo
// with the given minimum size.
func NewMinSizeFilter(size uintptr) MemorySegmentFilter {
	return NewFilterFromFunc(
		func(info *procio.MemorySegmentInfo) bool {
			return info.Size >= size
		},
		size,
		"segment too small, size: {{.MSI.Size|bytes}}, min-size: {{.Filter.Parameter|bytes}}",
		fmt.Sprintf("segment must be larger than %s", humanize.Bytes(uint64(size))),
	)
}

// NewStateFilter creates a new filter, matching *procio.MemorySegmentInfo
// with a procio.State equal to one of the given states.
func NewStateFilter(states []procio.State) MemorySegmentFilter {
	return NewFilterFromFunc(
		func(info *procio.MemorySegmentInfo) bool {
			for _, s := range states {
				if info.State == s {
					return true
				}
			}
			return false
		},
		states,
		"segment has wrong state, state: {{.MSI.State}}, allowed states: {{.Filter.Parameter|join \", \"}}",
		fmt.Sprintf("segment state must be one of %q", states),
	)
}

// NewTypeFilter creates a new filter, matching *procio.MemorySegmentInfo
// with a procio.SegmentType equal to one of the given types.
func NewTypeFilter(types []procio.SegmentType) MemorySegmentFilter {
	return NewFilterFromFunc(
		func(info *procio.MemorySegmentInfo) bool {
			for _, t := range types {
				if info.Type == t {
					return true
				}
			}
			return false
		},
		types,
		"segment has wrong type, type: {{.MSI.SegmentType}}, allowed types: {{.Filter.Parameter|join \", \"}}",
		fmt.Sprintf("segment type must be one of %q", types),
	)
}

// NewPermissionsFilterExact creates a new filter, matching *procio.MemorySegmentInfo
// with procio.Permissions exactly equal to one of the given perms.
func NewPermissionsFilterExact(perms []procio.Permissions) MemorySegmentFilter {
	return NewFilterFromFunc(
		func(info *procio.MemorySegmentInfo) bool {
			for _, p := range perms {
				if info.CurrentPermissions.EqualTo(p) {
					return true
				}
			}
			return false
		},
		perms,
		"segment has wrong permissions, permissions: {{.MSI.CurrentPermissions}}, allowed permissions: {{.Filter.Parameter|join \", \"}}",
		fmt.Sprintf("permissions must be equal to one of %q", perms),
	)
}

// NewPermissionsFilter creates a new filter, matching *procio.MemorySegmentInfo
// with procio.Permissions equal to or more permissive than the given perm.
func NewPermissionsFilter(perm procio.Permissions) MemorySegmentFilter {
	return NewFilterFromFunc(
		func(info *procio.MemorySegmentInfo) bool {
			return info.CurrentPermissions.IsMoreOrEquallyPermissiveThan(perm)
		},
		perm,
		"segment has wrong permissions, permissions: {{.MSI.CurrentPermissions}}, min-permissions: {{.Filter.Parameter}}",
		fmt.Sprintf("permissions must be equally or more permissive than %v", perm),
	)
}

// NewRSSRatioFilter creates a new filter, matching *procio.MemorySegmentInfo
// with RSS/Size ratio equal or greater than the given value.
func NewRSSRatioFilter(ratio float64) MemorySegmentFilter {
	return NewFilterFromFunc(
		func(info *procio.MemorySegmentInfo) bool {
			return float64(info.RSS)/float64(info.Size) >= ratio
		},
		ratio,
		"segment has too low RSS/Size ratio, actual ratio: {{.MSI.RSS}}/{{.MSI.Size}}, min ratio: {{.Filter.Parameter}}",
		fmt.Sprintf("RSS/Size ratio must be at least %v", ratio),
	)
}

type andFilter struct {
	filters []MemorySegmentFilter
}

// NewAndFilter creates a new filter, which is the logical AND-combination
// of all given MemorySegmentFilter instances.
func NewAndFilter(filters ...MemorySegmentFilter) MemorySegmentFilter {
	fil := make([]MemorySegmentFilter, 0, len(filters))
	for _, filter := range filters {
		if filter != nil {
			fil = append(fil, filter)
		}
	}
	return &andFilter{
		filters: fil,
	}
}

func (f *andFilter) Description() string {
	if len(f.filters) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(f.filters[0].Description())
	if len(f.filters) > 1 {
		for _, filter := range f.filters[1:] {
			if filter == nil {
				continue
			}

			sb.WriteString(" AND ")
			sb.WriteString(filter.Description())
		}
	}
	return sb.String()
}

func (f *andFilter) Filter(info *procio.MemorySegmentInfo) *FilterMatch {
	result := &FilterMatch{
		Result: true,
		MSI:    info,
	}
	reasons := make([]string, 0)
	for _, filter := range f.filters {
		r := filter.Filter(info)
		if !r.Result {
			result.Result = false
			reasons = append(reasons, result.Reason)
		}
	}
	if !result.Result {
		result.Reason = strings.Join(reasons, " AND ")
	}
	return result
}
