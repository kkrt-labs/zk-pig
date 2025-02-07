package generator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutor(t *testing.T) {
	for _, name := range testcases {
		t.Run(name, func(t *testing.T) {
			testDataInputs := loadTestDataInputs(t, testDataInputsPath(name))
			proverInput := &testDataInputs.ProverInput
			e := NewExecutor().(*executor)
			_, err := e.Execute(context.Background(), proverInput)
			assert.Equal(t, false, err != nil)
		})
	}
}
