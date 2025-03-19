package steps

import (
	"fmt"
	"strings"
)

func init() {
	for incl, str := range inclusionsStr {
		inclusionsStrReverse[str] = incl
	}
}

// Include is a bitmask that represents the data to include in the generated Prover Input.
type Include int

const (
	IncludeNone       Include = 0
	IncludeAccessList Include = 2 << 0
	IncludePreState   Include = 2 << 1
	IncludeStateDiffs Include = 2 << 2
	IncludeCommitted  Include = 2 << 3
	IncludeAll        Include = IncludeAccessList | IncludePreState | IncludeStateDiffs | IncludeCommitted
)

var inclusionsStr = map[Include]string{
	IncludeAll:        "all",
	IncludeNone:       "none",
	IncludeAccessList: "accessList",
	IncludePreState:   "preState",
	IncludeStateDiffs: "stateDiffs",
	IncludeCommitted:  "committed",
}

var (
	inclusionsStrReverse = map[string]Include{}
	ValidInclusions      = []Include{IncludeAccessList, IncludePreState, IncludeStateDiffs, IncludeCommitted, IncludeAll}
)

func (opt Include) String() string {
	if opt.Include(IncludeAll) {
		return inclusionsStr[IncludeAll]
	}
	inclusions := make([]string, 0)
	for incl, str := range inclusionsStr {
		if incl != IncludeNone && incl != IncludeAll && opt.Include(incl) {
			inclusions = append(inclusions, str)
		}
	}

	if len(inclusions) == 0 {
		return inclusionsStr[IncludeNone]
	}

	return strings.Join(inclusions, ",")
}

func (opt Include) Include(i Include) bool {
	return opt&i == i
}

// ParseInclude parses a string and returns the corresponding Inclusion value.
// It returns an error if the string is not a valid inclusion option.
func ParseInclude(strs ...string) (Include, error) {
	i := IncludeNone
	for _, str := range strs {
		if incl, ok := inclusionsStrReverse[str]; ok {
			i |= incl
		} else {
			return IncludeNone, fmt.Errorf("invalid inclusion option: %s", str)
		}
	}

	return i, nil
}

// WithInclusion sets the inclusion option for the Preparer.
func WithInclusion(inclusion Include) PrepareOption {
	return func(p *preparer) error {
		p.includeOpt = inclusion
		return nil
	}
}
