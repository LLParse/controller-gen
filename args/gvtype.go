package args

import (
	"flag"
	"fmt"
	"strings"

	"k8s.io/code-generator/cmd/client-gen/types"
)

type gvTypeValue struct {
	builder   *groupVersionsBuilder
	gvToTypes *map[types.GroupVersion][]string
	groups    []string
	changed   bool
}

func NewGVTypesValue(builder *groupVersionsBuilder, gvToTypes *map[types.GroupVersion][]string, def []string) *gvTypeValue {
	gvt := new(gvTypeValue)
	gvt.builder = builder
	gvt.gvToTypes = gvToTypes
	if def != nil {
		if err := gvt.set(def); err != nil {
			panic(err)
		}
	}
	return gvt
}

var _ flag.Value = &gvTypeValue{}

func (s *gvTypeValue) set(vs []string) error {
	if !s.changed {
		*s.gvToTypes = map[types.GroupVersion][]string{}
	}

	for _, input := range vs {
		gvString, typeStr, err := parseGroupVersionType(input)
		if err != nil {
			return err
		}
		s.groups = append(s.groups, gvString)
		gv, err := types.ToGroupVersion(gvString)
		if err != nil {
			return err
		}
		types, ok := (*s.gvToTypes)[gv]
		if !ok {
			types = []string{}
		}
		types = append(types, typeStr)
		(*s.gvToTypes)[gv] = types
	}

	s.builder.groups = s.groups
	return s.builder.update()
}

func (s *gvTypeValue) Set(val string) error {
	vs, err := readAsCSV(val)
	if err != nil {
		return err
	}
	if err := s.set(vs); err != nil {
		return err
	}
	s.changed = true
	return nil
}

func (s *gvTypeValue) Type() string {
	return "stringSlice"
}

func (s *gvTypeValue) String() string {
	strs := make([]string, 0, len(*s.gvToTypes))
	for gv, ts := range *s.gvToTypes {
		for _, t := range ts {
			strs = append(strs, gv.Group.String()+"/"+gv.Version.String()+"/"+t)
		}
	}
	str, _ := writeAsCSV(strs)
	return "[" + str + "]"
}

func parseGroupVersionType(gvtString string) (gvString string, typeStr string, err error) {
	invalidFormatErr := fmt.Errorf("invalid value: %s, should be of the form group/version/type", gvtString)
	subs := strings.Split(gvtString, "/")
	length := len(subs)
	switch length {
	case 2:
		// gvtString of the form group/type, e.g. api/Service,extensions/ReplicaSet
		return subs[0] + "/", subs[1], nil
	case 3:
		return strings.Join(subs[:length-1], "/"), subs[length-1], nil
	default:
		return "", "", invalidFormatErr
	}
}
