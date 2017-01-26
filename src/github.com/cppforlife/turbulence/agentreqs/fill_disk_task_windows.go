package agentreqs

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func (t FillDiskTask) Execute() error {
	if t.opts.Persistent {
		return t.fill("/var/vcap/store/.filler")
	}

	if t.opts.Ephemeral {
		return t.fill("/var/vcap/data/.filler")
	}

	if t.opts.Temporary {
		return t.fill(filepath.Join(os.TempDir(), ".filler"))
	}

	return t.fill("/.filler")
}

func (t FillDiskTask) fill(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return bosherr.WrapError(err, "Filling disk")
	}
	defer f.Close()

	zero := make([]byte, 1024*1024)
	var werr error
	for {
		if _, werr = f.Write(zero); werr != nil {
			break
		}
	}
	return bosherr.WrapError(werr, "Filling disk")
}
