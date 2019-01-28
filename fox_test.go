package fox

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type User struct {
	ID       int64  `gorm:"primary_key"`
	Username string `gorm:"unique;not null;index"`
	Salt     []byte `gorm:"not null"`
	Password []byte `gorm:"not null"`

	Key     []byte `gorm:"unique;index"`
	KeyTime []byte

	Avatar   string
	Nickname string `gorm:"index"`
	Phone    string `gorm:"index"`
	Email    string `gorm:"index"`
	Group    int    `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Person struct {
	Name string
	Age  uint
}

func TestBindAPI(t *testing.T) {
	p := &Person{"george", 19}
	st := reflect.TypeOf(p).Elem()
	v, err := st.FieldByName("nihao")
	fmt.Println(v, err)
	v1, err := st.FieldByName("Name")
	fmt.Println(v1, err)
	//type args struct {
	//	model interface{}
	//}
	//tests := []struct {
	//	name string
	//	args args
	//}{
	//	{"Base test", args{&User{}}},
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		err := BindAPI(tt.args.model)
	//		fmt.Println(1009, err)
	//	})
	//}
}
