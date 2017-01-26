// +build !windows

package agentreqs

import bosherr "github.com/cloudfoundry/bosh-utils/errors"

func (t FillDiskTask) Execute() error {
	if t.opts.Persistent {
		return t.fill("/var/vcap/store/.filler")
	}

	if t.opts.Ephemeral {
		return t.fill("/var/vcap/data/.filler")
	}

	if t.opts.Temporary {
		return t.fill("/tmp/.filler")
	}

	return t.fill("/.filler")
}

func (t FillDiskTask) fill(path string) error {
	_, _, _, err := t.cmdRunner.RunCommand("dd", "if=/dev/zero", "of="+path, "bs=1M")
	if err != nil {
		return bosherr.WrapError(err, "Filling disk")
	}

	return nil
}
