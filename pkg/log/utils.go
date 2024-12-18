package log

import (
	"github.com/kkrt-labs/kakarot-controller/pkg/tag"
	"go.uber.org/zap"
)

// TagsToFields converts a slice of tags to zap fields
func TagsToFields(tags []*tag.Tag) []zap.Field {
	fields := make([]zap.Field, 0, len(tags))
	for _, t := range tags {
		switch t.Value.Type {
		case tag.BOOL:
			fields = append(fields, zap.Bool(string(t.Key), t.Value.Interface.(bool)))
		case tag.INT64:
			fields = append(fields, zap.Int64(string(t.Key), t.Value.Interface.(int64)))
		case tag.FLOAT64:
			fields = append(fields, zap.Float64(string(t.Key), t.Value.Interface.(float64)))
		case tag.STRING:
			fields = append(fields, zap.String(string(t.Key), t.Value.Interface.(string)))
		case tag.OBJECT:
			fields = append(fields, zap.Any(string(t.Key), t.Value.Interface))
		}
	}
	return fields
}
