package selector_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/turbulence/incident/selector"
)

type SimpleInstance struct {
	id, group, deployment, az string
	missingVM                 bool
}

func (i SimpleInstance) ID() string         { return i.id }
func (i SimpleInstance) Group() string      { return i.group }
func (i SimpleInstance) Deployment() string { return i.deployment }
func (i SimpleInstance) AZ() string         { return i.az }
func (i SimpleInstance) HasVM() bool        { return !i.missingVM }

var _ = Describe("Limit", func() {
	Describe("Limit", func() {
		It("does basic selection", func() {
			str := `
{
	"Deployment": {
		"Name": "dep1"
	},
	"Group": {
		"Name": "group1"
	},
	"ID": {
		"Limit": "1"
	}
}`

			var req Request

			err := json.Unmarshal([]byte(str), &req)
			Expect(err).ToNot(HaveOccurred())

			in := []Instance{
				SimpleInstance{id: "id1-missing-vm", group: "group1", deployment: "dep1", az: "az1", missingVM: true},
				SimpleInstance{id: "id1", group: "group1", deployment: "dep1", az: "az1"},
				SimpleInstance{id: "id2", group: "group1", deployment: "dep1", az: "az1"},
				SimpleInstance{id: "id3", group: "group1", deployment: "dep1", az: "az1"},
				SimpleInstance{id: "id1", group: "group2", deployment: "dep1", az: "az1"},
				SimpleInstance{id: "id2", group: "group2", deployment: "dep1", az: "az2"},
				SimpleInstance{id: "id1", group: "group1", deployment: "dep2", az: "az1"},
			}

			out, err := req.AsSelector().Select(in)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(out)).To(Equal(1))
			Expect(out[0].ID()).ToNot(Equal("id1-missing-vm"))
		})

		It("supports wildcards", func() {
			str := `{ "Deployment": { "Name": "*-dep1-*" } }`

			var req Request

			err := json.Unmarshal([]byte(str), &req)
			Expect(err).ToNot(HaveOccurred())

			in := []Instance{
				SimpleInstance{deployment: "dep1"},
				SimpleInstance{deployment: "1-dep1-2"},
				SimpleInstance{deployment: "1-dep1-3-other"},
				SimpleInstance{deployment: "1-dep1"},
				SimpleInstance{deployment: "dep1-2"},
			}

			out, err := req.AsSelector().Select(in)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(out)).To(Equal(2))
			Expect([]string{
				out[0].Deployment(),
				out[1].Deployment(),
			}).To(ConsistOfLen(2, []string{"1-dep1-2", "1-dep1-3-other"}))
		})
	})
})
