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

type TaskOptions interface{}

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

			var opts interface{}

			switch {
			case optType == "kill":
				var o KillOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == "stress":
				var o StressOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == "control-net":
				var o ControlNetOptions
				err, opts = json.Unmarshal(bytes, &o), o

			case optType == "firewall":
				var o FirewallOptions
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
