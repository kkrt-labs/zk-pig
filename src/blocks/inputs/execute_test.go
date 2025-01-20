package blockinputs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutor(t *testing.T) {
	for _, name := range testcases {
		t.Run(name, func(t *testing.T) {
			testDataInputs := loadTestDataInputs(t, testDataInputsPath(name))
			proverInputs := &testDataInputs.ProverInputs
			e := NewExecutor().(*executor)
			_, err := e.Execute(context.Background(), proverInputs)
			assert.NoError(t, err)
		})
	}
}
