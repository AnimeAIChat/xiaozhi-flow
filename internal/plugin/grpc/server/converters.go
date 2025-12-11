package server

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
	"xiaozhi-server-go/internal/plugin/capability"
)

// ConvertSchemaToPB 转换Schema到protobuf格式
func ConvertSchemaToPB(schema capability.Schema) *structpb.Struct {
	if len(schema.Properties) == 0 && len(schema.Required) == 0 {
		return nil
	}

	result := &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}

	if schema.Type != "" {
		result.Fields["type"] = structpb.NewStringValue(schema.Type)
	}

	if len(schema.Properties) > 0 {
		properties := &structpb.Struct{Fields: make(map[string]*structpb.Value)}
		for key, prop := range schema.Properties {
			propStruct := &structpb.Struct{Fields: make(map[string]*structpb.Value)}
			propStruct.Fields["type"] = structpb.NewStringValue(prop.Type)
			if prop.Description != "" {
				propStruct.Fields["description"] = structpb.NewStringValue(prop.Description)
			}
			if prop.Default != nil {
				propStruct.Fields["default"] = ConvertInterfaceToPB(prop.Default)
			}
			if prop.Secret {
				propStruct.Fields["secret"] = structpb.NewBoolValue(prop.Secret)
			}
			properties.Fields[key] = structpb.NewStructValue(propStruct)
		}
		result.Fields["properties"] = structpb.NewStructValue(properties)
	}

	if len(schema.Required) > 0 {
		required := &structpb.ListValue{Values: make([]*structpb.Value, len(schema.Required))}
		for i, req := range schema.Required {
			required.Values[i] = structpb.NewStringValue(req)
		}
		result.Fields["required"] = structpb.NewListValue(required)
	}

	return result
}

// ConvertPBToMap 转换protobuf结构到map
func ConvertPBToMap(pb *structpb.Struct) map[string]interface{} {
	if pb == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})
	for key, value := range pb.Fields {
		result[key] = ConvertPBToInterface(value)
	}

	return result
}

// ConvertMapToPB 转换map到protobuf结构
func ConvertMapToPB(m map[string]interface{}) *structpb.Struct {
	if m == nil {
		return nil
	}

	fields := make(map[string]*structpb.Value)
	for key, value := range m {
		fields[key] = ConvertInterfaceToPB(value)
	}

	return &structpb.Struct{Fields: fields}
}

// ConvertPBToInterface 转换protobuf值到Go接口
func ConvertPBToInterface(v *structpb.Value) interface{} {
	switch v.Kind.(type) {
	case *structpb.Value_NullValue:
		return nil
	case *structpb.Value_NumberValue:
		return v.GetNumberValue()
	case *structpb.Value_StringValue:
		return v.GetStringValue()
	case *structpb.Value_BoolValue:
		return v.GetBoolValue()
	case *structpb.Value_StructValue:
		return ConvertPBToMap(v.GetStructValue())
	case *structpb.Value_ListValue:
		list := v.GetListValue()
		result := make([]interface{}, len(list.Values))
		for i, item := range list.Values {
			result[i] = ConvertPBToInterface(item)
		}
		return result
	default:
		return nil
	}
}

// ConvertInterfaceToPB 转换Go接口到protobuf值
func ConvertInterfaceToPB(v interface{}) *structpb.Value {
	switch val := v.(type) {
	case nil:
		return structpb.NewNullValue()
	case bool:
		return structpb.NewBoolValue(val)
	case int32:
		return structpb.NewNumberValue(float64(val))
	case int64:
		return structpb.NewNumberValue(float64(val))
	case float32:
		return structpb.NewNumberValue(float64(val))
	case float64:
		return structpb.NewNumberValue(val)
	case string:
		return structpb.NewStringValue(val)
	case map[string]interface{}:
		return structpb.NewStructValue(ConvertMapToPB(val))
	case []interface{}:
		list := &structpb.ListValue{}
		for _, item := range val {
			list.Values = append(list.Values, ConvertInterfaceToPB(item))
		}
		return structpb.NewListValue(list)
	default:
		return structpb.NewStringValue(fmt.Sprintf("%v", val))
	}
}
