package fox

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type (
	DataHandlerFunc func(key string, value []string, refValModel reflect.Value, refTypModel reflect.Type) (string, interface{}, error)
)

//type DataHandlerFunction func(key string, value []string, refValModel reflect.Value, refTypModel reflect.Type) (string, interface{}, error)

var (
	DB *gorm.DB

	ExtraQuery         []string          = []string{"EXT_LIMIT", "EXT_OFFSET", "EXT_ORDER_BY", "EXT_UNSCOPED"}
	SwitchQuery        []string          = []string{"LAST", "FRIST"}
	FalseValue         []string          = []string{"false", "f", "0", "no", "n", "null", "none"}
	ComparisonOperator map[string]string = map[string]string{"EQ": "=", "NE": "!=", "LT": "<", "LE": "<=", "GE": ">=", "GT": ">", "IN": "IN", "NOT_IN": "NOT IN", "LIKE": "LIKE"}
	LogicalOperator    map[string]string = map[string]string{"OR": "OR", "AND": "AND"}

	ErrorDBNotConnet     = errors.New("database not connet")
	ErrorNotStruct       = errors.New("not a struct")
	ErrorNotPtr          = errors.New("not a pointer")
	ErrorCannotSet       = errors.New("cannot set")
	ErrorType            = errors.New("error type")
	ErrorUnsupportedType = errors.New("unsupported type")
)

func IsExtraQuery(key string) bool {
	for i := range ExtraQuery {
		if strings.ToUpper(key) == ExtraQuery[i] {
			return true
		}
	}
	return false
}

func IsFalseValue(key string) bool {
	for i := range FalseValue {
		if strings.ToUpper(key) == FalseValue[i] {
			return true
		}
	}
	return false
}

func ParseOperator(lOperator []string) (sLogical, sComparison string) {
	for i := range lOperator {
		if sLogical != "" && sComparison != "" {
			break
		}
		sOperator := strings.ToUpper(lOperator[i][1:])
		if v, ok := LogicalOperator[sOperator]; sLogical == "" && ok {
			sLogical = v
		} else if v, ok := ComparisonOperator[sOperator]; sComparison == "" && ok {
			sComparison = v
		} else {
			continue
		}
	}
	if sLogical == "" {
		sLogical = "AND"
	}
	if sComparison == "" {
		sComparison = "="
	}
	return sLogical, sComparison
}

// ORM 驱动
type GORMDrive struct {
	Model      interface{}
	ModelSlice interface{}
}

func NewGORMDrive(model interface{}, modelSlice interface{}) (*GORMDrive, error) {
	if reflect.TypeOf(model).Kind() != reflect.Ptr {
		return nil, ErrorNotPtr
	}

	v := reflect.ValueOf(model).Elem()
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return nil, ErrorNotStruct
	}

	if reflect.ValueOf(DB).IsNil() {
		return nil, ErrorDBNotConnet
	}

	newDrive := &GORMDrive{model, modelSlice}
	DB.AutoMigrate(newDrive.Model)
	return newDrive, nil
}

func (self *GORMDrive) GenModel() (model reflect.Value) {
	sv := reflect.ValueOf(self.Model)
	model = reflect.New(sv.Elem().Type())
	model.Elem().Set(sv.Elem())
	return model
}

func (self *GORMDrive) GenModels() (models reflect.Value) {
	models = reflect.New(reflect.TypeOf(self.ModelSlice))
	return models
}

// 新建数据接口
func (self *GORMDrive) Insert(data url.Values) (interface{}, error) {
	refValModel := self.GenModel()

	result, err := self.DataHanlder(data, []DataHandlerFunc{ExistsField, StringConvert}...)
	if err != nil {
		return nil, err
	}

	model := refValModel.Interface()

	//refTypModel := reflect.TypeOf(model)
	modelElm := refValModel.Elem()
	modelType := modelElm.Type()
	for i := 0; i < modelType.NumField(); i++ {
		// todo: 没有考虑到内嵌结构体部分，这里写的捞一些
		if modelElm.Field(i).Kind() == reflect.Struct {
			embedStruField := modelElm.Field(i)
			//fmt.Println(1100, embedStruField)
			embedStruType := embedStruField.Type()
			for j := 0; j < embedStruField.NumField(); j++ {
				key := strings.ToLower(embedStruType.Field(j).Name)
				if inputVal, ok := result[key]; ok {
					embedStruField.Field(i).Set(reflect.ValueOf(inputVal).Convert(embedStruField.Field(i).Type()))
				}
			}
		}
		key := strings.ToLower(modelType.Field(i).Name)
		if inputVal, ok := result[key]; ok {
			modelElm.Field(i).Set(reflect.ValueOf(inputVal).Convert(modelElm.Field(i).Type()))
		}
	}

	resDB := DB.Create(model)
	if resDB.Error != nil {
		return nil, resDB.Error
	}
	return model, nil
}

// 查询数据接口
func (self *GORMDrive) Select(query url.Values) (interface{}, *gorm.DB, error) {
	refValueModels := self.GenModels()

	lModelsPtr := refValueModels.Interface()
	resDB, err := self.QueryParse(DB, query)
	if err != nil {
		return nil, resDB, err
	}
	var count uint
	resDB.Count(&count)

	result, err := self.DataHanlder(query, []DataHandlerFunc{KeyToUpper}...)
	limit := 20
	offset := 0
	order := ""
	if v, ok := result["EXT_LIMIT"]; ok {
		tmpValue, err := strconv.Atoi(v.([]string)[0])
		if err == nil {
			limit = tmpValue
		}
	}
	if v, ok := result["EXT_OFFSET"]; ok {
		tmpValue, err := strconv.Atoi(v.([]string)[0])
		if err == nil {
			offset = tmpValue
		}
	}
	if v, ok := result["EXT_ORDER_BY"]; ok {
		if err == nil {
			order = v.([]string)[0]
		}
	}

	resDB.Offset(offset).Limit(limit).Order(order).Find(lModelsPtr)
	return lModelsPtr, resDB, resDB.Error
}

// 更新数据接口
func (self *GORMDrive) Update(query, data url.Values) (int64, error) {
	result, err := self.DataHanlder(data, []DataHandlerFunc{ExistsField, StringConvert}...)
	if err != nil {
		return 0, err
	}

	resDB, err := self.QueryParse(DB.Model(self.Model), query)
	if err != nil {
		return 0, err
	}
	resDB = resDB.Updates(result)
	return resDB.RowsAffected, resDB.Error
}

// 删除数据接口
func (self *GORMDrive) Delete(query url.Values) (int64, error) {
	resDB, err := self.QueryParse(DB, query)
	if err != nil {
		return 0, err
	}
	result, err := self.DataHanlder(query, []DataHandlerFunc{KeyToUpper}...)
	if v, ok := result["EXT_UNSCOPED"]; ok {
		if !IsFalseValue(v.([]string)[0]) {
			resDB = resDB.Unscoped()
		}
	}
	resDB = resDB.Delete(self.Model)
	return resDB.RowsAffected, resDB.Error
}

func (self *GORMDrive) QueryParse(baseDB *gorm.DB, query url.Values) (resDB *gorm.DB, err error) {
	sWhereStatement := ""
	lWhereArgs := make([]interface{}, 0)

	isFirstCondition := true
	mv := reflect.TypeOf(self.Model).Elem()

	for k, v := range query {
		regFieldName := regexp.MustCompile(`^\w+`)
		sFieldName := strings.ToLower(regFieldName.FindString(k))

		if _, isExist := mv.FieldByNameFunc(func(s string) bool {
			return strings.ToLower(s) == sFieldName
		}); !isExist {
			continue
		}

		regOperator := regexp.MustCompile(`\.(\w+)`)
		sLogicalOperator, sComparisonOperator := ParseOperator(regOperator.FindAllString(k, -1))
		if isFirstCondition {
			sLogicalOperator = ""
		} else {
			sLogicalOperator += " "
		}

		if sComparisonOperator == "" {
			sComparisonOperator = ComparisonOperator["EQ"]
		}

		sWhereStatement += fmt.Sprintf("%s%s%s? ", sLogicalOperator, sFieldName, sComparisonOperator)
		lWhereArgs = append(lWhereArgs, v)
		isFirstCondition = false
	}

	resDB = DB.Model(self.Model).Where(sWhereStatement, lWhereArgs...)
	return resDB, resDB.Error
}

func (self *GORMDrive) DataHanlder(data url.Values, handlers ...DataHandlerFunc) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	mt := reflect.TypeOf(self.Model).Elem()
	mv := reflect.ValueOf(self.Model).Elem()
	for k, v := range data {
		for _, handler := range handlers {
			newKey, newValue, err := handler(k, v, mv, mt)
			if err != nil {
				return map[string]interface{}{}, err
			}
			if newKey == "" && newValue == "" {
				continue
			}
			result[newKey] = newValue
		}
	}
	return result, nil
}
