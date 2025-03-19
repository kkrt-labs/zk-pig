package steps

import (
	"fmt"
	"strings"
)

// Include is a bitmask that represents the data to include in the generated Prover Input.
type Include int

const (
	expAccessList = 0
	expPreState   = 1
	expStateDiffs = 2
	expCommitted  = 3
)

const (
	IncludeNone       Include = 0
	IncludeAccessList Include = 1 << expAccessList
	IncludePreState   Include = 1 << expPreState
	IncludeStateDiffs Include = 1 << expStateDiffs
	IncludeCommitted  Include = 1 << expCommitted
	IncludeAll        Include = IncludeAccessList | IncludePreState | IncludeStateDiffs | IncludeCommitted
)

var ValidIncludes = []Include{
	IncludeNone,
	IncludeAccessList,
	IncludePreState,
	IncludeStateDiffs,
	IncludeCommitted,
	IncludeAll,
}

var (
	includeNoneStr = "none"
	includeAllStr  = "all"
	// includesStr MUST respect the order of the exp* constants.
	includesStr = []string{
		"accessList",
		"preState",
		"stateDiffs",
		"committed",
		includeAllStr,
		includeNoneStr,
	}
)

var includesStrReverse = map[string]Include{
	includesStr[expAccessList]: IncludeAccessList,
	includesStr[expPreState]:   IncludePreState,
	includesStr[expStateDiffs]: IncludeStateDiffs,
	includesStr[expCommitted]:  IncludeCommitted,
	includeAllStr:              IncludeAll,
	includeNoneStr:             IncludeNone,
}

func (opt Include) String() string {
	if opt == IncludeNone {
		return includeNoneStr
	}

	if opt.Include(IncludeAll) {
		return includeAllStr
	}

	inclusions := make([]string, 0)
	for i, inclStr := range includesStr[:len(includesStr)-2] {
		if opt.Include(Include(1 << i)) {
			inclusions = append(inclusions, inclStr)
		}
	}

	if len(inclusions) == 0 {
		return includeNoneStr
	}

	return strings.Join(inclusions, ",")
}

func (opt Include) Include(i Include) bool {
	return opt&i == i
}

// ParseIncludes parses a string and returns the corresponding Inclusion value.
// It returns an error if the string is not a valid inclusion option.
func ParseIncludes(strs ...string) (Include, error) {
	incl := IncludeNone
	for _, str := range strs {
		if in, ok := includesStrReverse[str]; ok {
			incl |= in
		} else {
			return IncludeNone, fmt.Errorf("invalid inclusion option: %s", str)
		}
	}

	return incl, nil
}

// WithDataInclude sets the inclusion option for the Preparer.
func WithDataInclude(include Include) PrepareOption {
	return func(p *preparer) error {
		p.includeOpt = include
		return nil
	}
}
