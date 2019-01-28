package fox

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ExtraExpression    []string          = []string{"LIMIT", "OFFSET", "ORDER_BY"}
	BehaviorExpression []string          = []string{"LOAD_FK"}
	ComparisonOperator map[string]string = map[string]string{"EQ": "=", "NE": "!=", "LT": "<", "LE": "<=", "GE": ">=", "GT": ">", "IN": "IN", "LIKE": "LIKE"}
	LogicalOperator    map[string]string = map[string]string{"OR": "OR", "AND": "AND", "NOT": "NOT"}
)

func IsExtraExpression(key string) bool {
	for i := range ExtraExpression {
		if key == ExtraExpression[i] {
			return true
		}
	}
	return false
}

func IsComparisonOperator(key string) bool {
	for i := range ComparisonOperator {
		if key == ComparisonOperator[i] {
			return true
		}
	}
	return false
}

func IsLogicalOperator(key string) bool {
	for i := range LogicalOperator {
		if key == LogicalOperator[i] {
			return true
		}
	}
	return false
}

func SoftOperator() {

}

// 主要用于 SQL 语句的生成
type GORMDrive struct {
	model interface{}

	SqlStatement *SQLStatement
}

// 新建数据接口
func (self *GORMDrive) Insert(formData map[string]string) (reflect.Value, error) {
	sv := reflect.ValueOf(self.model)
	ref := reflect.New(sv.Elem().Type())
	ref.Elem().Set(sv.Elem())
	if err := Set(ref.Interface(), formData); err != nil {
		return ref, err
	}
	return ref, nil
}

// 查询数据接口
func (self *GORMDrive) Select(query map[string]string) () {
}

// 更新数据接口
func (self *GORMDrive) Update(query map[string]string) () {
}

// 删除数据接口
func (self *GORMDrive) Delete(query map[string]string) () {
}

type SQLStatement struct {
	ExtraExpression map[string]string
	WhereExpression map[string]string
}

func NewSQLStatement(model interface{}, query map[string]string) *SQLStatement {
	newSql := &SQLStatement{ExtraExpression: make(map[string]string), WhereExpression: make(map[string]string)}
	isFirstCondition := true
	for k, v := range query {
		if IsExtraExpression(k) {
			newSql.ExtraExpression[k] = v
		} else {
			reg := regexp.MustCompile(`(\w+)(\.\w+)?(\.\w+)?`)
			s := reg.FindStringSubmatch(k)
			mv := reflect.TypeOf(model).Elem()
			if _, isExist := mv.FieldByName(s[1]); !isExist {
				continue
			}
			condition := ""
			if !isFirstCondition {
				condition = "AND "
			}
			switch len(s) {
			case 2:
				condition = fmt.Sprintf("%s%s = ?", condition, s[1])
				newSql.WhereExpression[condition] = v
			case 3:
				switch {
				case IsComparisonOperator(s[2]):
					condition = fmt.Sprintf("%s%s %s ?", condition, s[1], s[2])
				case IsLogicalOperator(s[2]):
					condition = fmt.Sprintf("%s %s = ?", s[2], s[1])
				}

			}
			newSql.WhereExpression[s[1]] = v
			isFirstCondition = false
		}
	}
	return newSql
}

// 主要用于 WHERE 表达式的生成
type WhereExpression struct {
}

func NewWhereExpression(query map[string]string) (*WhereExpression, error) {
	sqlStatement := &WhereExpression{}
	return sqlStatement, nil
}

func (self *WhereExpression) Parse(query map[string]string) {
	//reg := regexp.MustCompile(`^(\w+).(\w+)$`)
	//first := true
	//for k, v := range query {
	//	result := reg.FindStringSubmatch(k)
	//	if first {
	//
	//	}
	//}
}

//func (self *WhereExpression) IsComparisonOperator(oper string) bool {
//	_, ok := self.ComparisonOperator[oper]
//	return ok
//}
//
//func (self *WhereExpression) IsLogicalOperator(oper string) bool {
//	_, ok := self.LogicalOperator[oper]
//	return ok
//}

//func (self *WhereExpression) NewCondition(logical, key, comparison, value string) (*Condition, error) {
//	newCond := &Condition{}
//	defaults.Set(newCond)
//	if logical == "" {
//		logical = "AND"
//	} else if v, ok := newCond.LogicalOperator[strings.ToUpper(logical)]; ok {
//		logical = v
//	} else {
//		return nil, errors.New("this")
//	}
//	if comparison == "" {
//		logical = "="
//	}
//	return
//}

type Condition struct {
	Logical            string
	Key                string
	Comparison         string
	Value              string
	ComparisonOperator map[string]string `default:"{\"EQ\":\"=\",\"NE\":\"!=\",\"LT\":\"<\",\"LE\":\"<=\",\"GE\":\">=\",\"GT\":\">\",\"IN\":\"IN\",\"NOT\":\"NOT\",\"IS\":\"IS\",\"LIKE\":\"LIKE\"}"`
	LogicalOperator    map[string]string `default:"{\"OR\":\"OR\",\"AND\":\"AND\",\"NOT\":\"NOT\"}"`
}

//func NewCondition(logical, key, comparison, value string) (*Condition, error) {
//	newCond := &Condition{}
//	defaults.Set(newCond)
//	if logical == "" {
//		logical = "AND"
//	} else if v, ok := newCond.LogicalOperator[strings.ToUpper(logical)]; ok {
//		logical = v
//	} else {
//		return nil, errors.New("this")
//	}
//	if comparison == "" {
//		logical = "="
//	}
//	return
//}

func (self *Condition) GenLogicalExpression() (string, string) {
	return fmt.Sprintf("%s %s %s ?", self.Logical, self.Key, self.Comparison), self.Value
}

var (
	ErrorNotStructPtr = errors.New("not a struct pointer")
)

type ModelAPI struct {
	model interface{}
}

func Set(model interface{}, formdata map[string]string) error {
	if reflect.TypeOf(model).Kind() != reflect.Ptr {
		return ErrorNotStructPtr
	}

	v := reflect.ValueOf(model).Elem()
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return ErrorNotStructPtr
	}

	for i := 0; i < t.NumField(); i++ {
		fmt.Println(1, strings.ToLower(t.Field(i).Name))
		fmt.Println(2, formdata[strings.ToLower(t.Field(i).Name)])
		if strings.ToLower(t.Field(i).Name) == "username" {
			fmt.Println(1111)
			v.Field(i).Set(reflect.ValueOf("thisisadmin").Convert(v.Field(i).Type()))
		}
		if inputVal, ok := formdata[strings.ToLower(t.Field(i).Name)]; ok {
			if err := setField(v.Field(i), inputVal); err != nil {
				return err
			}
		}
	}

	return nil
}

type Setter interface {
	SetDefaults()
}

func callSetter(v interface{}) {
	if ds, ok := v.(Setter); ok {
		ds.SetDefaults()
	}
}

func setField(field reflect.Value, inputVal string) error {
	if !field.CanSet() {
		return nil
	}
	if field.Kind() != reflect.Struct || inputVal == "" {
		return nil
	}

	if reflect.DeepEqual(reflect.Zero(field.Type()).Interface(), field.Interface()) {
		switch field.Kind() {
		case reflect.Bool:
			if val, err := strconv.ParseBool(inputVal); err == nil {
				field.Set(reflect.ValueOf(val).Convert(field.Type()))
			}
		case reflect.Int:
			if val, err := strconv.ParseInt(inputVal, 10, 64); err == nil {
				field.Set(reflect.ValueOf(int(val)).Convert(field.Type()))
			}
		case reflect.Int8:
			if val, err := strconv.ParseInt(inputVal, 10, 8); err == nil {
				field.Set(reflect.ValueOf(int8(val)).Convert(field.Type()))
			}
		case reflect.Int16:
			if val, err := strconv.ParseInt(inputVal, 10, 16); err == nil {
				field.Set(reflect.ValueOf(int16(val)).Convert(field.Type()))
			}
		case reflect.Int32:
			if val, err := strconv.ParseInt(inputVal, 10, 32); err == nil {
				field.Set(reflect.ValueOf(int32(val)).Convert(field.Type()))
			}
		case reflect.Int64:
			if val, err := time.ParseDuration(inputVal); err == nil {
				field.Set(reflect.ValueOf(val).Convert(field.Type()))
			} else if val, err := strconv.ParseInt(inputVal, 10, 64); err == nil {
				field.Set(reflect.ValueOf(val).Convert(field.Type()))
			}
		case reflect.Uint:
			if val, err := strconv.ParseUint(inputVal, 10, 64); err == nil {
				field.Set(reflect.ValueOf(uint(val)).Convert(field.Type()))
			}
		case reflect.Uint8:
			if val, err := strconv.ParseUint(inputVal, 10, 8); err == nil {
				field.Set(reflect.ValueOf(uint8(val)).Convert(field.Type()))
			}
		case reflect.Uint16:
			if val, err := strconv.ParseUint(inputVal, 10, 16); err == nil {
				field.Set(reflect.ValueOf(uint16(val)).Convert(field.Type()))
			}
		case reflect.Uint32:
			if val, err := strconv.ParseUint(inputVal, 10, 32); err == nil {
				field.Set(reflect.ValueOf(uint32(val)).Convert(field.Type()))
			}
		case reflect.Uint64:
			if val, err := strconv.ParseUint(inputVal, 10, 64); err == nil {
				field.Set(reflect.ValueOf(val).Convert(field.Type()))
			}
		case reflect.Uintptr:
			if val, err := strconv.ParseUint(inputVal, 10, 64); err == nil {
				field.Set(reflect.ValueOf(uintptr(val)).Convert(field.Type()))
			}
		case reflect.Float32:
			if val, err := strconv.ParseFloat(inputVal, 32); err == nil {
				field.Set(reflect.ValueOf(float32(val)).Convert(field.Type()))
			}
		case reflect.Float64:
			if val, err := strconv.ParseFloat(inputVal, 64); err == nil {
				field.Set(reflect.ValueOf(val).Convert(field.Type()))
			}
		case reflect.String:
			field.Set(reflect.ValueOf(inputVal).Convert(field.Type()))

		case reflect.Slice:
			ref := reflect.New(field.Type())
			ref.Elem().Set(reflect.MakeSlice(field.Type(), 0, 0))
			switch field.Type().Elem().Kind() {
			case reflect.Uint8:
				if h, err := hex.DecodeString(inputVal); err == nil {
					field.Set(reflect.ValueOf(h).Convert(field.Type()))
				} else {
					return err
				}

			default:
				if inputVal != "" && inputVal != "[]" {
					if err := json.Unmarshal([]byte(inputVal), ref.Interface()); err != nil {
						return err
					}
				}
				field.Set(ref.Elem().Convert(field.Type()))
			}
		case reflect.Map:
			ref := reflect.New(field.Type())
			ref.Elem().Set(reflect.MakeMap(field.Type()))
			if inputVal != "" && inputVal != "{}" {
				if err := json.Unmarshal([]byte(inputVal), ref.Interface()); err != nil {
					return err
				}
			}
			field.Set(ref.Elem().Convert(field.Type()))
		case reflect.Struct:
			ref := reflect.New(field.Type())
			if inputVal != "" && inputVal != "{}" {
				if err := json.Unmarshal([]byte(inputVal), ref.Interface()); err != nil {
					return err
				}
			}
			field.Set(ref.Elem())
		case reflect.Ptr:
			field.Set(reflect.New(field.Type().Elem()))
		}
	}

	if field.Kind() == reflect.Ptr {
		setField(field.Elem(), inputVal)
		callSetter(field.Interface())
	}

	return nil
}

func BindAPI(model interface{}) (reflect.Value, error) {
	formData := map[string]string{"username": "admin", "password": "1213ba"}

	sv := reflect.ValueOf(model)
	ref := reflect.New(sv.Elem().Type())
	ref.Elem().Set(sv.Elem())
	//if err := Set(ref.Interface(), formData); err != nil {
	//	return err
	//}
	//callSetter(ref.Interface())
	if err := Set(ref.Interface(), formData); err != nil {
		return ref, err
	}
	fmt.Println(10001, ref)
	return ref, nil
}
