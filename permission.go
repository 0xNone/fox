package fox

import (
	"github.com/pelletier/go-toml"
)

var ()

type Tree struct {
	value map[string]map[string]int
}

//func (t *Tree) Get(key string) interface{} {
//	keySlice := strings.Split(key, ".")
//	if len(keySlice) != 2 {
//		return nil
//	}
//	if _, ok := t.value[keySlice[0]]; !ok {
//		return nil
//	} else {
//		if _, ok := t.value[keySlice[0]][keySlice[1]]; !ok {
//			return nil
//		}
//	}
//}

type Permissions struct {
	tree *toml.Tree
}

func NewPermissions() {
	//tree, _ := toml.Load("")
}

func (p *Permissions) SetTable() {
}
