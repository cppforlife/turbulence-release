package selector

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

var (
	limitRegexp = regexp.MustCompile(`\A([0-9]+)(%?)(?:\s*\-\s*([0-9]+)(%?))?\z`)
)

type Limit struct { // e.g. 1, 0-1, 0%-20%, 20%
	start, end int
	percent    bool
	applied    bool // by default limit is equivalent to 100%
}

func NewLimitorFromString(s string) (Limit, error) {
	matches := limitRegexp.FindStringSubmatch(s)
	if len(matches) == 0 {
		return Limit{}, fmt.Errorf("Limit must match '%s'", limitRegexp)
	}

	if len(matches) != 5 {
		panic("Internal inconsistency: port range regexp mismatch")
	}

	// matches[0] is full
	// matches[1] is start
	// matches[2] is %
	// matches[3] is end
	// matches[4] is %

	start, err := strconv.Atoi(matches[1])
	if err != nil {
		panic(fmt.Sprintf("Expected '%s' to be an int", matches[1]))
	}

	if matches[3] != "" && matches[2] != matches[4] {
		return Limit{}, fmt.Errorf("Limit range start and end must both be percent or integer")
	}

	isPercent := matches[2] == "%"

	if matches[3] == "" {
		return NewLimit(start, start, isPercent)
	}

	end, err := strconv.Atoi(matches[3])
	if err != nil {
		panic(fmt.Sprintf("Expected '%s' to be an int", matches[3]))
	}

	return NewLimit(start, end, isPercent)
}

func NewLimit(start, end int, percent bool) (Limit, error) {
	if start < 0 {
		return Limit{}, errors.New("Limit start cannot be negative")
	}
	if percent {
		if start > 100 {
			return Limit{}, errors.New("Limit start cannot be over 100%")
		}
		if end > 100 {
			return Limit{}, errors.New("Limit end cannot be over 100%")
		}
	}
	if start > end {
		return Limit{}, errors.New("Limit start must be <= end")
	}
	return Limit{start: start, end: end, percent: percent, applied: true}, nil
}

func (l Limit) Limit(in []string) ([]string, error) {
	if !l.applied {
		return in, nil
	}

	// todo should we error over here?
	// if !l.percent && l.start > len(in) {
	// 	return nil, errors.New("Limit range start is larger than size of input")
	// }

	picked := []string{}

	for _, idx := range l.selectIdxs(len(in)) {
		picked = append(picked, in[idx])
	}

	if l.start != 0 && len(picked) == 0 {
		return nil, errors.New("Expected limit to keep at least one item")
	}

	return picked, nil
}

func (l Limit) String() string {
	if !l.applied {
		return ""
	}
	suffix := ""
	if l.percent {
		suffix = "%"
	}
	if l.start == l.end {
		return strconv.Itoa(l.start) + suffix
	}
	return fmt.Sprintf("%d%s-%d%s", l.start, suffix, l.end, suffix)
}

func (l Limit) selectIdxs(max int) []int {
	n := l.start
	if l.end > l.start {
		n += rand.Intn(l.end - l.start)
	}
	return rand.Perm(max)[0:l.numOrPercent(n, max)]
}

func (l Limit) numOrPercent(n, max int) int {
	if l.percent {
		n = int(math.Ceil(float64(n) / 100.0 * float64(max)))
	}
	if n > max {
		return max
	}
	return n
}

func (l *Limit) UnmarshalJSON(s []byte) error {
	if string(s) == `""` {
		*l = Limit{} // not applied
		return nil
	}

	limit, err := NewLimitorFromString(strings.Replace(string(s), `"`, "", -1))
	if err != nil {
		return err
	}

	*l = limit

	return nil
}

func (l *Limit) MarshalJSON() ([]byte, error) {
	return []byte(`"` + l.String() + `"`), nil
}
