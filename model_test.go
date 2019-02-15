package fox

import (
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Person struct {
	ID   uint
	Name string
	Age  uint
}

type TestModel struct {
	Bool        bool
	Int8        int8
	Int16       int16
	Int32       int32
	Int64       int64
	Int         int
	Uint8       uint8
	Uint16      uint16
	Uint32      uint32
	Uint64      uint64
	Uint        uint
	UintPtr     *uint
	Float32     float32
	Float64     float64
	String      string
	ByteSlice   []byte
	StringSlice []string
	MapString   map[string]string
	Time        time.Time
	TimePtr     *time.Time
}

type Pet struct {
	gorm.Model
	Name string
	Type string
}

// todo: 多类型测试后续补上
var (
	auint uint = 99489541

	testModel1 = TestModel{
		Bool:    false,
		Int8:    15,
		Int16:   30245,
		Int32:   9975468,
		Int64:   9147483647,
		Int:     1575468,
		Uint8:   185,
		Uint16:  50245,
		Uint32:  3457483647,
		Uint64:  16223372036854775807,
		Uint:    6579483647,
		UintPtr: &auint,
		Float32: 6579483647.4849,
		Float64: 657948347849.4849,
		String:  "asdwqfcfqw"}
)

func init() {
	var err error
	DB, err = gorm.Open("sqlite3", "database.db")
	if err != nil {
		panic("数据库未连接")
	}
	DB.AutoMigrate(&Person{})
}

func TestParseOperator(t *testing.T) {
	type args struct {
		lOperator []string
	}
	tests := []struct {
		name            string
		args            args
		wantSLogical    string
		wantSComparison string
	}{
		{"baseTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.or", -1)}, "OR", "="},
		{"baseTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.and", -1)}, "AND", "="},
		{"baseTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.eq", -1)}, "AND", "="},
		{"baseTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.ne", -1)}, "AND", "!="},
		{"baseTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.lt", -1)}, "AND", "<"},
		{"baseTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.le", -1)}, "AND", "<="},
		{"baseTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.ge", -1)}, "AND", ">="},
		{"baseTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.gt", -1)}, "AND", ">"},
		{"baseTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.in", -1)}, "AND", "IN"},
		{"baseTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.not_in", -1)}, "AND", "NOT IN"},
		{"baseTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.like", -1)}, "AND", "LIKE"},
		{"MultipleOperatorsTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.or.ne", -1)}, "OR", "!="},
		{"MultipleOperatorsTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.or.eq", -1)}, "OR", "="},
		{"MultipleOperatorsTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.and.in", -1)}, "AND", "IN"},
		{"MultipleOperatorsTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.or.eq.and", -1)}, "OR", "="},
		{"MultipleOperatorsTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.or.and.in.like", -1)}, "OR", "IN"},
		{"MultipleOperatorsTest", args{regexp.MustCompile(`\.(\w+)`).FindAllString("id.and.eq.and.like", -1)}, "AND", "="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSLogical, gotSComparison := ParseOperator(tt.args.lOperator)
			if gotSLogical != tt.wantSLogical {
				t.Errorf("ParseOperator() gotSLogicalOperator = %v, want %v", gotSLogical, tt.wantSLogical)
			}
			if gotSComparison != tt.wantSComparison {
				t.Errorf("ParseOperator() gotSComparisonOperator = %v, want %v", gotSComparison, tt.wantSComparison)
			}
		})
	}
}

func TestNewGORMDrive(t *testing.T) {
	str := "Person"

	type args struct {
		model      interface{}
		modelSlice interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *GORMDrive
		wantErr bool
	}{
		{"BaseTest", args{&Person{}, []Person{}}, &GORMDrive{&Person{}, []Person{}}, false},
		{"BaseTest", args{&Pet{}, []Pet{}}, &GORMDrive{&Pet{}, []Pet{}}, false},
		{"ErrorTest", args{Person{}, []Person{}}, nil, true},
		{"ErrorTest", args{&str, []Person{}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGORMDrive(tt.args.model, tt.args.modelSlice)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGORMDrive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGORMDrive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGORMDrive_GenModel(t *testing.T) {
	tmpPson := &Person{}

	type fields struct {
		Model      interface{}
		ModelSlice interface{}
	}
	tests := []struct {
		name      string
		fields    fields
		wantModel reflect.Value
	}{
		{"BaseTest", fields{tmpPson, []Person{}}, reflect.ValueOf(tmpPson)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := &GORMDrive{
				Model:      tt.fields.Model,
				ModelSlice: tt.fields.ModelSlice,
			}
			if gotModel := self.GenModel(); reflect.DeepEqual(gotModel, tt.wantModel) {
				t.Errorf("a new model should have been generated. but GORMDrive.GenModel() = %v = %v", gotModel, tt.wantModel)
			}
		})
	}
}

func TestGORMDrive_GenModels(t *testing.T) {
	tmpPsonSlice := []Person{}

	type fields struct {
		Model      interface{}
		ModelSlice interface{}
	}
	tests := []struct {
		name       string
		fields     fields
		wantModels reflect.Value
	}{
		{"BaseTest", fields{&Person{}, tmpPsonSlice}, reflect.ValueOf(tmpPsonSlice)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := &GORMDrive{
				Model:      tt.fields.Model,
				ModelSlice: tt.fields.ModelSlice,
			}
			if gotModels := self.GenModels(); reflect.DeepEqual(gotModels, tt.wantModels) {
				t.Errorf("a new model slice should have been generated. GORMDrive.GenModels() = %v = %v", gotModels, tt.wantModels)
			}
		})
	}
}

func TestGORMDrive_Insert(t *testing.T) {
	lastPerson := Person{}
	DB.Last(&lastPerson)
	lastPet := Pet{}
	DB.Last(&lastPet)

	type fields struct {
		Model      interface{}
		ModelSlice interface{}
	}

	personField := fields{&Person{}, []Person{}}
	petField := fields{&Pet{}, []Pet{}}

	type args struct {
		data url.Values
	}

	// todo: 继承的结构体无法赋值
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{"BaseTest", personField, args{url.Values{"id": []string{strconv.FormatUint(uint64(lastPerson.ID+1), 10)}, "name": []string{"Li Lei"}, "age": []string{"18"}}}, &Person{ID: lastPerson.ID + 1, Name: "Li Lei", Age: 18}, false},
		{"BaseTest", personField, args{url.Values{"id": []string{strconv.FormatUint(uint64(lastPerson.ID+2), 10)}, "name": []string{"Da Ming"}, "age": []string{"23"}}}, &Person{ID: lastPerson.ID + 2, Name: "Da Ming", Age: 23}, false},
		{"BaseTest", personField, args{url.Values{"id": []string{strconv.FormatUint(uint64(lastPerson.ID+3), 10)}, "name": []string{"Han Meimei"}, "age": []string{"41"}}}, &Person{ID: lastPerson.ID + 3, Name: "Han Meimei", Age: 41}, false},
		{"BaseTest", personField, args{url.Values{"id": []string{strconv.FormatUint(uint64(lastPerson.ID+4), 10)}, "name": []string{"Han Keke"}, "age": []string{"12"}}}, &Person{ID: lastPerson.ID + 4, Name: "Han Keke", Age: 12}, false},
		{"BaseTest", personField, args{url.Values{"id": []string{strconv.FormatUint(uint64(lastPerson.ID+5), 10)}, "name": []string{"Han Xixi"}, "age": []string{"26"}}}, &Person{ID: lastPerson.ID + 5, Name: "Han Xixi", Age: 26}, false},
		{"BaseTest", personField, args{url.Values{"id": []string{strconv.FormatUint(uint64(lastPerson.ID+6), 10)}, "name": []string{"Mike"}, "age": []string{"35"}}}, &Person{ID: lastPerson.ID + 6, Name: "Mike", Age: 35}, false},
		{"BaseTest", personField, args{url.Values{"id": []string{strconv.FormatUint(uint64(lastPerson.ID+7), 10)}, "name": []string{"Kate"}, "age": []string{"46"}}}, &Person{ID: lastPerson.ID + 7, Name: "Kate", Age: 46}, false},
		{"BaseTest", personField, args{url.Values{"id": []string{strconv.FormatUint(uint64(lastPerson.ID+8), 10)}, "name": []string{"Anna"}, "age": []string{"52"}}}, &Person{ID: lastPerson.ID + 8, Name: "Anna", Age: 52}, false},
		{"BaseTest", personField, args{url.Values{"id": []string{strconv.FormatUint(uint64(lastPerson.ID+9), 10)}, "name": []string{"Jean"}, "age": []string{"38"}}}, &Person{ID: lastPerson.ID + 9, Name: "Jean", Age: 38}, false},
		{"BaseTest", personField, args{url.Values{"id": []string{strconv.FormatUint(uint64(lastPerson.ID+10), 10)}, "name": []string{"Jim"}}}, &Person{ID: lastPerson.ID + 10, Name: "Jim", Age: 0}, false},
		{"BaseTest", petField, args{url.Values{"name": []string{"Cheese"}, "type": []string{"dog"}}}, &Pet{Name: "Cheese", Type: "dog"}, false},
		{"BaseTest", petField, args{url.Values{"name": []string{"Bobi"}, "type": []string{"cat"}}}, &Pet{Name: "Bobi", Type: "cat"}, false},
		{"BaseTest", petField, args{url.Values{"name": []string{"Pee"}, "type": []string{"mouse"}}}, &Pet{Name: "Pee", Type: "mouse"}, false},
		{"BaseTest", petField, args{url.Values{"name": []string{"Mut"}, "type": []string{"fish"}, "atta": []string{"asss"}}}, &Pet{Name: "Mut", Type: "fish"}, false},
		{"BaseTest", petField, args{url.Values{"name": []string{"Gee"}, "": []string{""}}}, &Pet{Name: "Gee", Type: ""}, false},
		{"BaseTest", petField, args{url.Values{"name": []string{"Cookie"}}}, &Pet{Name: "Cookie", Type: ""}, false},
		{"BaseTest", petField, args{url.Values{"name": []string{"Kitty"}}}, &Pet{Name: "Kitty", Type: ""}, false},
		{"ErrorTest", personField, args{url.Values{"id": []string{strconv.FormatUint(uint64(lastPerson.ID+1), 10)}, "name": []string{"Li Lei"}, "age": []string{"68"}}}, nil, true},
		{"ErrorTest", personField, args{url.Values{"id": []string{"hakaka"}, "name": []string{"Li Lei"}, "age": []string{"75"}}}, nil, true},
		{"ErrorTest", personField, args{url.Values{"id": []string{strconv.FormatUint(uint64(lastPerson.ID+3), 10)}, "name": []string{"Li Lei"}, "age": []string{"wkai"}}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := &GORMDrive{
				Model:      tt.fields.Model,
				ModelSlice: tt.fields.ModelSlice,
			}
			got, err := self.Insert(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GORMDrive.Insert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			switch reflect.TypeOf(got) {
			case reflect.TypeOf(&Pet{}):
				gotElem := got.(*Pet)
				wantElem := tt.want.(*Pet)
				if gotElem.Name != wantElem.Name || gotElem.Type != wantElem.Type {
					t.Errorf("GORMDrive.Insert() = %v, want %v", got, tt.want)
				}
			case reflect.TypeOf(&Person{}):
				//gotElem := got.(*Person)
				if !reflect.DeepEqual(got, tt.want) {
				}
			}

			//if gotElem.ID != wantElem.ID || gotElem.Name != wantElem.Name || gotElem.Age != wantElem.Age {
			//	t.Errorf("GORMDrive.Insert() = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestGORMDrive_Select(t *testing.T) {
	type fields struct {
		Model      interface{}
		ModelSlice interface{}
	}

	personField := fields{&Person{}, []Person{}}

	type args struct {
		query url.Values
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		//wantQueryResult map[string]interface{}
		wantErr bool
		wantLen int
	}{
		{"BaseTest", personField, args{url.Values{"id": []string{"1"}, "ext_limit": []string{"4"}}}, false, 1},
		{"BaseTest", personField, args{url.Values{"id.lt": []string{"4"}, "id.and.ge": []string{"2"}, "ext_limit": []string{"4"}}}, false, 2},
		{"BaseTest", personField, args{url.Values{"id.lt": []string{"8"}, "ext_limit": []string{"4"}, "ext_offset": []string{"2"}}}, false, 4},
		{"BaseTest", personField, args{url.Values{"id.lt": []string{"5"}, "ext_order_by": []string{"id"}}}, false, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := &GORMDrive{
				Model:      tt.fields.Model,
				ModelSlice: tt.fields.ModelSlice,
			}
			gotQueryResult, _, err := self.Select(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("GORMDrive.Select() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if resLen := reflect.ValueOf(gotQueryResult).Elem().Len(); resLen != tt.wantLen {
				t.Errorf("The result length of GORMDrive.Select() = %v, wantLen %v", resLen, tt.wantLen)
				return
			}
			//if !reflect.DeepEqual(gotQueryResult, tt.wantQueryResult) {
			//	t.Errorf("GORMDrive.Select() = %v, want %v", gotQueryResult, tt.wantQueryResult)
			//}
		})
	}
}

func TestGORMDrive_Update(t *testing.T) {
	type fields struct {
		Model      interface{}
		ModelSlice interface{}
	}
	personField := fields{&Person{}, []Person{}}

	type args struct {
		query url.Values
		data  url.Values
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{"BaseTest", personField, args{url.Values{"id.lt": []string{"4"}}, url.Values{"age": []string{"45"}}}, 3, false},
		{"BaseTest", personField, args{url.Values{"id.eq": []string{"4"}, "age.gt": []string{"199"}}, url.Values{"age": []string{"45"}}}, 0, false},
		{"ErrorTest", personField, args{url.Values{"id.lt": []string{"4"}}, url.Values{"age": []string{"asdw"}}}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := &GORMDrive{
				Model:      tt.fields.Model,
				ModelSlice: tt.fields.ModelSlice,
			}
			got, err := self.Update(tt.args.query, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GORMDrive.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GORMDrive.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGORMDrive_Delete(t *testing.T) {
	lastPerson := Person{}
	DB.Last(&lastPerson)
	lastPet := Pet{}
	DB.Last(&lastPet)

	type fields struct {
		Model      interface{}
		ModelSlice interface{}
	}
	personField := fields{&Person{}, []Person{}}
	petField := fields{&Pet{}, []Pet{}}

	type args struct {
		query url.Values
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{"BaseTest", personField, args{url.Values{"id.gt": []string{strconv.FormatUint(uint64(lastPerson.ID-4), 10)}}}, 4, false},
		{"BaseTest", personField, args{url.Values{"id.gt": []string{strconv.FormatUint(uint64(lastPerson.ID-4), 10)}}}, 0, false},
		{"BaseTest", petField, args{url.Values{"id.gt": []string{strconv.FormatUint(uint64(lastPet.ID-2), 10)}}}, 2, false},
		{"BaseTest", petField, args{url.Values{"id.gt": []string{strconv.FormatUint(uint64(lastPet.ID-2), 10)}, "EXT_UNSCOPED": []string{"true"}}}, 2, false},
		{"BaseTest", petField, args{url.Values{"id.gt": []string{strconv.FormatUint(uint64(lastPet.ID-4), 10)}, "EXT_UNSCOPED": []string{"false"}}}, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := &GORMDrive{
				Model:      tt.fields.Model,
				ModelSlice: tt.fields.ModelSlice,
			}
			got, err := self.Delete(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("GORMDrive.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GORMDrive.Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}
