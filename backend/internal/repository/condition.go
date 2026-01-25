package repository

import (
	"fmt"
	"strings"
)

func BuildCondition(where map[string]any) (whereSql string,
	values []any, err error,
) {

	for key, value := range where {
		conditionKey := strings.Split(key, " ")
		if whereSql != "" {
			whereSql += " AND "
		}
		// '=', '<', '>', '<=', '>=', '<>', '!=', '<=>',
		// 'like', 'like binary', 'not like', 'ilike',
		// '&', '|', '^', '<<', '>>',
		// 'rlike', 'regexp', 'not regexp',
		// '~', '~*', '!~', '!~*', 'similar to',
		// 'not similar to', 'not ilike', '~~*', '!~~*',
		switch len(conditionKey) {
		case 1:
			// 处理等于
			// 检查是否包含换行符，如果包含则转换为 IN 查询
			if strVal, ok := value.(string); ok && strings.Contains(strVal, "\n") {
				lines := strings.Split(strVal, "\n")
				var inValues []string
				for _, line := range lines {
					trimLine := strings.TrimSpace(line)
					if trimLine != "" {
						inValues = append(inValues, trimLine)
					}
				}

				if len(inValues) > 0 {
					whereSql += fmt.Sprint(conditionKey[0], " in ?")
					values = append(values, inValues)
				} else {
					// 如果分割后没有有效值，回退到原逻辑
					whereSql += fmt.Sprint(conditionKey[0], " = ?")
					values = append(values, value)
				}
			} else {
				// 检查是否为切片类型，如果是则使用 IN 查询
				// 如果合并case需要使用reflect有性能消耗
				switch v := value.(type) {
				case []string:
					if len(v) > 0 {
						whereSql += fmt.Sprint(conditionKey[0], " IN ?")
						values = append(values, v)
					} else {
						whereSql += "1 = 0"
					}
				case []uint:
					if len(v) > 0 {
						whereSql += fmt.Sprint(conditionKey[0], " IN ?")
						values = append(values, v)
					} else {
						whereSql += "1 = 0"
					}
				case []int:
					if len(v) > 0 {
						whereSql += fmt.Sprint(conditionKey[0], " IN ?")
						values = append(values, v)
					} else {
						whereSql += "1 = 0"
					}
				case []int64:
					if len(v) > 0 {
						whereSql += fmt.Sprint(conditionKey[0], " IN ?")
						values = append(values, v)
					} else {
						whereSql += "1 = 0"
					}
				case []any:
					if len(v) > 0 {
						whereSql += fmt.Sprint(conditionKey[0], " IN ?")
						values = append(values, v)
					} else {
						whereSql += "1 = 0"
					}
				default:
					// 其他类型使用等于
					whereSql += fmt.Sprint(conditionKey[0], " = ?")
					values = append(values, value)
				}
			}
		case 2:
			// 处理单条件
			field := conditionKey[0]
			switch conditionKey[1] {
			case "=":
				// 检查是否包含换行符，如果包含则转换为 IN 查询
				if strVal, ok := value.(string); ok && strings.Contains(strVal, "\n") {
					lines := strings.Split(strVal, "\n")
					var inValues []string
					for _, line := range lines {
						trimLine := strings.TrimSpace(line)
						if trimLine != "" {
							inValues = append(inValues, trimLine)
						}
					}

					if len(inValues) > 0 {
						whereSql += fmt.Sprint(field, " in ?")
						values = append(values, inValues)
					} else {
						whereSql += fmt.Sprint(field, " = ?")
						values = append(values, value)
					}
				} else {
					whereSql += fmt.Sprint(field, " = ?")
					values = append(values, value)
				}
			case ">", ">=", "<", "<=", "<>", "!=":
				whereSql += fmt.Sprint(field, " "+conditionKey[1]+" ?")
				values = append(values, value)
			case "in":
				// "code in" : "111,222,333" 或 []uint{111, 222, 333}
				whereSql += fmt.Sprint(field, " in ?")
				switch v := value.(type) {
				case string:
					inMap := strings.Split(v, ",")
					values = append(values, inMap)
				case []uint:
					values = append(values, v)
				case []string:
					values = append(values, v)
				default:
					return "", nil, fmt.Errorf("unsupported type for 'in' operator: %T", value)
				}
			case "notin":
				// "code notin" : "111,222,333" 或 []uint{111, 222, 333}
				whereSql += fmt.Sprint(field, " not in ?")
				switch v := value.(type) {
				case string:
					inMap := strings.Split(v, ",")
					values = append(values, inMap)
				case []uint:
					values = append(values, v)
				case []string:
					values = append(values, v)
				default:
					return "", nil, fmt.Errorf("unsupported type for 'notin' operator: %T", value)
				}
			case "like":
				// "code like" : "%111%"
				whereSql += fmt.Sprint(field, " like ?")
				values = append(values, value.(string))
			}
		default:
			// 处理 条件拼接,条件可以相同，也可以不同
			// created_by = ? or task_user_name = ?，need two arguments
			// "code = ? or name = ?" : "111,222"
			// "code like ? or name like ?" : "%111%,%222%"
			whereSql += fmt.Sprint(key)
			switch v := value.(type) {
			case string:
				inMap := strings.Split(v, ",")
				for _, val := range inMap {
					values = append(values, val)
				}
			case []uint:
				for _, val := range v {
					values = append(values, val)
				}
			default:
				return "", nil, fmt.Errorf("unsupported type for condition: %T", value)
			}

		}
	}

	return
}
