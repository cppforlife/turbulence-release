package agentreqs

import (
	"encoding/json"
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type TaskReq struct {
	Error string
}

// TaskOptionsSlice is used for unmarshalling different source types
type TaskOptionsSlice []TaskOptions

type TaskOptions interface {
	_private()
}

func TaskOptsType(taskOpts TaskOptions) string {
	t := fmt.Sprintf("%T", taskOpts)
	t = strings.TrimPrefix(t, "agentreqs.")
	return strings.TrimSuffix(t, "Options")
}

func (s *TaskOptionsSlice) UnmarshalJSON(data []byte) error {
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

			var opts TaskOptions

			switch {
			case optType == TaskOptsType(KillOptions{}):
				var o KillOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == TaskOptsType(KillProcessOptions{}):
				var o KillProcessOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == TaskOptsType(StressOptions{}):
				var o StressOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == TaskOptsType(ControlNetOptions{}):
				var o ControlNetOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == TaskOptsType(FirewallOptions{}):
				var o FirewallOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == TaskOptsType(FillDiskOptions{}):
				var o FillDiskOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == TaskOptsType(ShutdownOptions{}):
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

func (s TaskOptionsSlice) MarshalJSON() ([]byte, error) {
	for i, o := range s {
		switch typedO := o.(type) {
		case KillOptions:
			typedO.Type = TaskOptsType(typedO)
			s[i] = typedO

		case KillProcessOptions:
			typedO.Type = TaskOptsType(typedO)
			s[i] = typedO

		case StressOptions:
			typedO.Type = TaskOptsType(typedO)
			s[i] = typedO

		case ControlNetOptions:
			typedO.Type = TaskOptsType(typedO)
			s[i] = typedO

		case FirewallOptions:
			typedO.Type = TaskOptsType(typedO)
			s[i] = typedO

		case FillDiskOptions:
			typedO.Type = TaskOptsType(typedO)
			s[i] = typedO

		case ShutdownOptions:
			typedO.Type = TaskOptsType(typedO)
			s[i] = typedO

		default:
			return nil, bosherr.Errorf("Unknown task type '%T'", o)
		}
	}

	return json.Marshal([]TaskOptions(s))
}
