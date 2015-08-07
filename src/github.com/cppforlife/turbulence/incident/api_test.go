package incident_test

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	incident "github.com/cppforlife/turbulence/incident"
)

var _ = Describe("Job", func() {
	Describe("SelectedIndices", func() {
		Context("when Indices are specified", func() {
			It("returns specified indices", func() {
				job := incident.Job{Indices: []int{1, 2, 3}}
				Expect(job.SelectedIndices([]int{1, 2, 3})).To(ConsistOf([]int{1, 2, 3}))
			})

			It("does not include specified indices that are not available", func() {
				job := incident.Job{Indices: []int{1, 2, 3, 6}}
				Expect(job.SelectedIndices([]int{2, 3, 5})).To(ConsistOf([]int{2, 3}))
			})
		})

		limitTests := func() {
			Context("when specified limit is 0", func() {
				Context("as number", func() {
					It("returns no indices", func() {
						job := incident.Job{Limit: "0"}
						Expect(job.SelectedIndices([]int{})).To(HaveLen(0))
						Expect(job.SelectedIndices([]int{1, 2, 3})).To(HaveLen(0))
					})
				})

				Context("as percentage", func() {
					It("returns no indices", func() {
						job := incident.Job{Limit: "0%"}
						Expect(job.SelectedIndices([]int{})).To(HaveLen(0))
						Expect(job.SelectedIndices([]int{1, 2, 3})).To(HaveLen(0))
					})
				})
			})

			Context("when specified limit is smaller than number of available indices", func() {
				Context("as number", func() {
					It("returns all available indices", func() {
						job := incident.Job{Limit: "2"}
						Expect(job.SelectedIndices([]int{1, 2, 3})).To(ConsistOfLen(2, []int{1, 2, 3}))
					})
				})

				Context("as percentage", func() {
					It("returns all available indices", func() {
						job := incident.Job{Limit: "33%"}
						Expect(job.SelectedIndices([]int{1, 2, 3})).To(ConsistOfLen(1, []int{1, 2, 3}))

						job = incident.Job{Limit: "34%"}
						Expect(job.SelectedIndices([]int{1, 2, 3})).To(ConsistOfLen(2, []int{1, 2, 3}))

						job = incident.Job{Limit: "66%"}
						Expect(job.SelectedIndices([]int{1, 2, 3})).To(ConsistOfLen(2, []int{1, 2, 3}))

						job = incident.Job{Limit: "67%"}
						Expect(job.SelectedIndices([]int{1, 2, 3})).To(ConsistOfLen(3, []int{1, 2, 3}))
					})
				})
			})

			Context("when specified limit is same as number of available indices", func() {
				Context("as number", func() {
					It("returns all available indices", func() {
						job := incident.Job{Limit: "3"}
						Expect(job.SelectedIndices([]int{1, 2, 3})).To(ConsistOf([]int{1, 2, 3}))
					})
				})

				Context("as percentage", func() {
					It("returns all available indices", func() {
						job := incident.Job{Limit: "100%"}
						Expect(job.SelectedIndices([]int{1, 2, 3})).To(ConsistOf([]int{1, 2, 3}))
					})
				})
			})

			Context("when specified limit is larger than number of available indices", func() {
				Context("as number", func() {
					It("returns all available indices", func() {
						job := incident.Job{Limit: "5"}
						Expect(job.SelectedIndices([]int{1, 2, 3})).To(ConsistOf([]int{1, 2, 3}))
					})
				})

				Context("as percentage", func() {
					It("returns no indices", func() {
						job := incident.Job{Limit: "140%"}
						Expect(job.SelectedIndices([]int{1, 2, 3})).To(ConsistOf([]int{1, 2, 3}))
					})
				})
			})

			It("is fair enough in choosing indices", func() {
				job := incident.Job{Limit: "1"}
				counts := map[int]int{}

				for i := 0; i < 10000; i++ {
					indices, err := job.SelectedIndices([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
					Expect(err).ToNot(HaveOccurred())

					for _, j := range indices {
						counts[j]++
					}
				}

				Expect(counts).To(HaveLen(10))

				for _, count := range counts {
					Expect(count).To(BeNumerically(">=", 800)) // at least 8% for each index
				}
			})
		}

		Context("when Limit is specified", func() {
			BeforeEach(func() {
				rand.Seed(time.Now().UTC().UnixNano())
			})

			for i := 0; i < 100; i++ {
				Describe(fmt.Sprintf("iteration %d", i), limitTests)
			}
		})
	})
})
