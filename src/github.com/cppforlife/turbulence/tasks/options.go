package tasks

import (
	"encoding/json"
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func OptionsType(taskOpts Options) string {
	t := fmt.Sprintf("%T", taskOpts)
	t = strings.TrimPrefix(t, "tasks.")
	return strings.TrimSuffix(t, "Options")
}

func (s *OptionsSlice) UnmarshalJSON(data []byte) error {
	var maps []map[string]interface{}

	err := json.Unmarshal(data, &maps)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshalling task options")
	}

	for _, m := range maps {
		if optType, ok := m["Type"]; ok {
			bytes, err := json.Marshal(m)
			if err != nil {
				return bosherr.WrapErrorf(err, "Marshalling task options")
			}

			var opts Options

			switch {
			case optType == OptionsType(KillOptions{}):
				var o KillOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == OptionsType(KillProcessOptions{}):
				var o KillProcessOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == OptionsType(StressOptions{}):
				var o StressOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == OptionsType(ControlNetOptions{}):
				var o ControlNetOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == OptionsType(FirewallOptions{}):
				var o FirewallOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == OptionsType(FillDiskOptions{}):
				var o FillDiskOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == OptionsType(ShutdownOptions{}):
				var o ShutdownOptions
				err, opts = json.Unmarshal(bytes, &o), o

			default:
				err = bosherr.Errorf("Unknown task type '%s'", optType)
			}

			if err != nil {
				return bosherr.WrapErrorf(err, "Unmarshalling task type '%s'", optType)
			}

			*s = append(*s, opts)
		} else {
			return bosherr.Error("Missing task type")
		}
	}

	return nil
}

func (s OptionsSlice) MarshalJSON() ([]byte, error) {
	for i, o := range s {
		switch typedO := o.(type) {
		case KillOptions:
			typedO.Type = OptionsType(typedO)
			s[i] = typedO

		case KillProcessOptions:
			typedO.Type = OptionsType(typedO)
			s[i] = typedO

		case StressOptions:
			typedO.Type = OptionsType(typedO)
			s[i] = typedO

		case ControlNetOptions:
			typedO.Type = OptionsType(typedO)
			s[i] = typedO

		case FirewallOptions:
			typedO.Type = OptionsType(typedO)
			s[i] = typedO

		case FillDiskOptions:
			typedO.Type = OptionsType(typedO)
			s[i] = typedO

		case ShutdownOptions:
			typedO.Type = OptionsType(typedO)
			s[i] = typedO

		default:
			return nil, bosherr.Errorf("Unknown task type '%T'", o)
		}
	}

	return json.Marshal([]Options(s))
}
