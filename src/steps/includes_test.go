package steps

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInclusionString(t *testing.T) {
	tests := []struct {
		incl Include
		want string
	}{
		{IncludeNone, "none"},
		{IncludeAll, "all"},
		{IncludeAccessList, "accessList"},
		{IncludePreState, "preState"},
		{IncludeStateDiffs, "stateDiffs"},
		{IncludeCommitted, "committed"},
		{IncludeAccessList | IncludePreState, "accessList,preState"},
		{IncludeAccessList | IncludePreState | IncludeStateDiffs, "accessList,preState,stateDiffs"},
		{IncludeAccessList | IncludePreState | IncludeStateDiffs | IncludeCommitted, "all"},
	}
	for _, test := range tests {
		if got := test.incl.String(); got != test.want {
			assert.Equal(t, test.want, got, "Inclusion(%v).String() = %q; want %q", test.incl, got, test.want)
		}
	}
}

func TestParseInclusion(t *testing.T) {
	tests := []struct {
		strs []string

		expectedIncl Include
		expectedErr  bool
	}{
		{[]string{"none"}, IncludeNone, false},
		{[]string{"all"}, IncludeAll, false},
		{[]string{"accessList"}, IncludeAccessList, false},
		{[]string{"preState"}, IncludePreState, false},
		{[]string{"stateDiffs"}, IncludeStateDiffs, false},
		{[]string{"committed"}, IncludeCommitted, false},
		{[]string{"accessList", "preState"}, IncludeAccessList | IncludePreState, false},
		{[]string{"accessList", "preState", "stateDiffs"}, IncludeAccessList | IncludePreState | IncludeStateDiffs, false},
		{[]string{"all", "none"}, IncludeAll, false},
		{[]string{"all", "none", "invalid"}, 0, true},
	}
	for _, test := range tests {
		got, err := ParseInclude(test.strs...)
		if test.expectedErr {
			require.Error(t, err, "ParseInclusion(%v) = %v; want error", test.strs, got)
		} else {
			require.NoError(t, err)
			assert.Equalf(t, test.expectedIncl, got, "ParseInclusion(%v) = %v; want %v", test.strs, got, test.expectedIncl)
		}
	}
}

func TestValidInclusions(t *testing.T) {
	assert.Equal(t, "[\"accessList\" \"preState\" \"stateDiffs\" \"committed\" \"all\"]", fmt.Sprintf("%q", ValidInclusions))
}
