package field

import (
	_ "github.com/leebrouse/eino/internal/config"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/spf13/viper"
)

// FieldConfig 描述一个字段的所有可配置项
type FieldConfig struct {
	Name       string            `json:"name"`
	DataType   entity.FieldType  `json:"dataType"`
	PrimaryKey bool              `json:"primaryKey,omitempty"`
	AutoID     bool              `json:"autoID,omitempty"`
	TypeParams map[string]string `json:"typeParams,omitempty"`
}

// DefaultConfig 给出默认的字段列表（和你原来写死的一致）
func DefaultConfig() []FieldConfig {
	return []FieldConfig{
		{
			Name:       "id",
			DataType:   entity.FieldTypeVarChar,
			PrimaryKey: true,
			AutoID:     true,
			TypeParams: map[string]string{"max_length": "128"},
		},
		{
			Name:       "content",
			DataType:   entity.FieldTypeVarChar,
			TypeParams: map[string]string{"max_length": "4096"},
		},
		{
			Name:     "metadata",
			DataType: entity.FieldTypeJSON,
		},
		{
			Name:       "vector",
			DataType:   entity.FieldTypeFloatVector,
			TypeParams: map[string]string{"dim": viper.GetString("gemini.dim")},
		},
	}
}

// NewFields 根据传入的字段配置生成 []*entity.Field
// 如果 cfg 为 nil，则使用 DefaultConfig()
func NewFields(cfg []FieldConfig) []*entity.Field {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	fields := make([]*entity.Field, 0, len(cfg))
	for _, c := range cfg {
		f := &entity.Field{
			Name:       c.Name,
			DataType:   c.DataType,
			PrimaryKey: c.PrimaryKey,
			AutoID:     c.AutoID,
			TypeParams: c.TypeParams,
		}
		fields = append(fields, f)
	}
	return fields
}
