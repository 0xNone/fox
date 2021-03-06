package fox

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"net/url"
	"reflect"
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
		sOperator := strings.ToUpper(lOperator[i])
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

func (gormDri *GORMDrive) GenModel() (model reflect.Value) {
	sv := reflect.ValueOf(gormDri.Model)
	model = reflect.New(sv.Elem().Type())
	model.Elem().Set(sv.Elem())
	return model
}

func (gormDri *GORMDrive) GenModels() (models reflect.Value) {
	models = reflect.New(reflect.TypeOf(gormDri.ModelSlice))
	return models
}

// 新建数据接口
func (gormDri *GORMDrive) Insert(data url.Values) (interface{}, error) {
	refValModel := gormDri.GenModel()

	result, err := gormDri.DataHanlder(data, []DataHandlerFunc{ExistsField, StringConvert}...)
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
func (gormDri *GORMDrive) Select(query url.Values, queryString string) (interface{}, *gorm.DB, error) {
	refValueModels := gormDri.GenModels()

	lModelsPtr := refValueModels.Interface()
	resDB, err := gormDri.QueryParse(DB, query, queryString)

	if err != nil {
		return nil, resDB, err
	}
	var count uint
	resDB.Count(&count)

	result, err := gormDri.DataHanlder(query, []DataHandlerFunc{KeyToUpper}...)
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
func (gormDri *GORMDrive) Update(query, data url.Values, queryString string) (int64, error) {
	result, err := gormDri.DataHanlder(data, []DataHandlerFunc{ExistsField, StringConvert}...)
	if err != nil {
		return 0, err
	}

	resDB, err := gormDri.QueryParse(DB.Model(gormDri.Model), query, queryString)
	if err != nil {
		return 0, err
	}
	resDB = resDB.Updates(result)
	return resDB.RowsAffected, resDB.Error
}

// 删除数据接口
func (gormDri *GORMDrive) Delete(query url.Values, queryString string) (int64, error) {
	resDB, err := gormDri.QueryParse(DB, query, queryString)
	if err != nil {
		return 0, err
	}
	result, err := gormDri.DataHanlder(query, []DataHandlerFunc{KeyToUpper}...)
	if v, ok := result["EXT_UNSCOPED"]; ok {
		if !IsFalseValue(v.([]string)[0]) {
			resDB = resDB.Unscoped()
		}
	}
	resDB = resDB.Delete(gormDri.Model)
	return resDB.RowsAffected, resDB.Error
}

func (gormDri *GORMDrive) QueryParse(baseDB *gorm.DB, query url.Values, queryString string) (resDB *gorm.DB, err error) {
	sWhereStatement := ""
	lWhereArgs := make([]interface{}, 0)

	isFirstCondition := true
	mv := reflect.TypeOf(gormDri.Model).Elem()

	querySlice := strings.Split(queryString, "&")
	for i := range querySlice {
		qk := strings.Split(querySlice[i], "=")[0]
		//regFieldName := regexp.MustCompile(`^\w+`)
		qkSlice := strings.Split(qk, ".")
		//sFieldName := strings.ToLower(qkSlice[0])

		if _, isExist := mv.FieldByNameFunc(func(s string) bool {
			return strings.ToLower(s) == qkSlice[0]
		}); !isExist {
			continue
		}

		//regOperator := regexp.MustCompile(`\.(\w+)`)
		sLogicalOperator, sComparisonOperator := ParseOperator(qkSlice)
		if isFirstCondition {
			sLogicalOperator = ""
		} else {
			sLogicalOperator += " "
		}

		sWhereStatement += fmt.Sprintf("%s%s%s? ", sLogicalOperator, qkSlice[0], sComparisonOperator)
		lWhereArgs = append(lWhereArgs, query[qk][0])
		isFirstCondition = false
	}
	resDB = DB.Model(gormDri.Model).Where(sWhereStatement, lWhereArgs...)
	return resDB, resDB.Error
}

func (gormDri *GORMDrive) DataHanlder(data url.Values, handlers ...DataHandlerFunc) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	mt := reflect.TypeOf(gormDri.Model).Elem()
	mv := reflect.ValueOf(gormDri.Model).Elem()
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
