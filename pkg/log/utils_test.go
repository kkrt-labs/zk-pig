package log

import (
	"testing"

	"github.com/kkrt-labs/kakarot-controller/pkg/tag"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestTagsToFields(t *testing.T) {
	tests := []struct {
		name     string
		tags     []*tag.Tag
		expected []zapcore.Field
	}{
		{
			name: "converts all tag types correctly",
			tags: []*tag.Tag{
				{Key: "bool_key", Value: tag.BoolValue(true)},
				{Key: "int64_key", Value: tag.Int64Value(42)},
				{Key: "float64_key", Value: tag.Float64Value(3.14)},
				{Key: "string_key", Value: tag.StringValue("test")},
				{Key: "object_key", Value: tag.ObjectValue(struct {
					Test string
					Num  int
				}{Test: "test", Num: 10})},
			},
			expected: []zapcore.Field{
				zap.Bool("bool_key", true),
				zap.Int64("int64_key", 42),
				zap.Float64("float64_key", 3.14),
				zap.String("string_key", "test"),
				zap.Any("object_key", struct {
					Test string
					Num  int
				}{Test: "test", Num: 10}),
			},
		},
		{
			name:     "empty tags returns empty fields",
			tags:     []*tag.Tag{},
			expected: []zapcore.Field{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := TagsToFields(tt.tags)
			assert.ElementsMatch(t, tt.expected, fields)
		})
	}
}
