package selector_test

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/turbulence/incident/selector"
)

var _ = Describe("Limit", func() {
	Describe("Limit", func() {
		limitTests := func() {
			Context("when specified limit is 0", func() {
				Context("as number", func() {
					It("returns no indices", func() {
						limit := MustNewLimitFromString("0")
						Expect(limit.Limit([]string{})).To(HaveLen(0))
						Expect(limit.Limit([]string{"1", "2", "3"})).To(HaveLen(0))
					})
				})

				Context("as percentage", func() {
					It("returns no indices", func() {
						limit := MustNewLimitFromString("0%")
						Expect(limit.Limit([]string{})).To(HaveLen(0))
						Expect(limit.Limit([]string{"1", "2", "3"})).To(HaveLen(0))
					})
				})
			})

			Context("when specified limit is smaller than number of available indices", func() {
				Context("as number", func() {
					It("returns all available indices", func() {
						limit := MustNewLimitFromString("2")
						Expect(limit.Limit([]string{"1", "2", "3"})).To(ConsistOfLen(2, []string{"1", "2", "3"}))
					})
				})

				Context("as percentage", func() {
					It("returns all available indices", func() {
						limit := MustNewLimitFromString("33%")
						Expect(limit.Limit([]string{"1", "2", "3"})).To(ConsistOfLen(1, []string{"1", "2", "3"}))

						limit = MustNewLimitFromString("34%")
						Expect(limit.Limit([]string{"1", "2", "3"})).To(ConsistOfLen(2, []string{"1", "2", "3"}))

						limit = MustNewLimitFromString("66%")
						Expect(limit.Limit([]string{"1", "2", "3"})).To(ConsistOfLen(2, []string{"1", "2", "3"}))

						limit = MustNewLimitFromString("67%")
						Expect(limit.Limit([]string{"1", "2", "3"})).To(ConsistOfLen(3, []string{"1", "2", "3"}))
					})
				})
			})

			Context("when specified limit is exactly half of available indices", func() {
				Context("as number", func() {
					It("returns half of available indices", func() {
						limit := MustNewLimitFromString("1")
						Expect(limit.Limit([]string{"1", "2"})).To(ConsistOfLen(1, []string{"1", "2"}))
					})
				})

				Context("as percentage", func() {
					It("returns half of available indices", func() {
						limit := MustNewLimitFromString("50%")
						Expect(limit.Limit([]string{"1", "2"})).To(ConsistOfLen(1, []string{"1", "2"}))
					})
				})
			})

			Context("when specified limit is same as number of available indices", func() {
				Context("as number", func() {
					It("returns all available indices", func() {
						limit := MustNewLimitFromString("3")
						Expect(limit.Limit([]string{"1", "2", "3"})).To(ConsistOf([]string{"1", "2", "3"}))
					})
				})

				Context("as percentage", func() {
					It("returns all available indices", func() {
						limit := MustNewLimitFromString("100%")
						Expect(limit.Limit([]string{"1", "2", "3"})).To(ConsistOf([]string{"1", "2", "3"}))
					})
				})
			})

			Context("when specified limit is larger than number of available indices", func() {
				Context("as number", func() {
					It("returns all available indices", func() {
						limit := MustNewLimitFromString("5")
						Expect(limit.Limit([]string{"1", "2", "3"})).To(ConsistOf([]string{"1", "2", "3"}))
					})
				})

				Context("as percentage", func() {
					It("returns all available indices", func() {
						limit := MustNewLimitFromString("100%") // cannot be over 100%
						Expect(limit.Limit([]string{"1", "2", "3"})).To(ConsistOf([]string{"1", "2", "3"}))
					})
				})
			})

			It("is fair enough in choosing indices", func() {
				limit := MustNewLimitFromString("1")
				counts := map[string]int{}

				for i := 0; i < 10000; i++ {
					vals, err := limit.Limit([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"})
					Expect(err).ToNot(HaveOccurred())

					for _, j := range vals {
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

			i := 0

			for i = 0; i < 100; i++ {
				Describe(fmt.Sprintf("iteration %d", i), limitTests)
			}
		})
	})
})
