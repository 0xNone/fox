# FOX
![CI](https://travis-ci.org/0xNone/fox.svg?branch=master)

**TODO**

+ Model
  + ~~query、data 等数据进行基本的增删改查操作~~
+ View
  + ~~web 框架 echo 包装~~
  + ~~根据 model 自动创建 RUST API~~
  + 中间件
  + 返回状态码规范
+ Permission
+ API 文档

之前用了 fy0 使用 python 写的 [slim](https://github.com/fy0/slim)，开发速度提高了很多，最近对 go 比较上心，写一个 go 版本的 fox

该框架基于 gorm 和 echo，可以自动添加增删改查 API，开发者可以从这些基础接口释放了。虽然还很粗糙，以后在摸鱼的时间慢慢磨。可以说是懒得要死了

# 特性
+ RESTful API 规范
+ 在数据库中自动创建相关 model 表
+ 开箱即用的增删改查 API

# 安装
```bash
$ go get -u https://github.com/0xNone/fox.git
```

# 快速开始

```cgo
package main

import (
	"github.com/0xNone/fox"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Person struct {
	ID   uint
	Name string
	Age  uint
}

func main() {
	var err error
	fox.DB, err = gorm.Open("sqlite3", "database.db")
	err = fox.ModelRoute(&Person{}, []Person{})
	if err != nil {
		panic(err)
	}
	fox.RawRouter.Start(":8000")
}
```

运行 `go run main.go`，启动服务

现在使用 GET、POST、PUT、DELETE 访问 http://localhost:8000/person，即可对数据库进行操作。默认会存在 `POST`, `GET`, `PUT`, `DELETE` 这四个方法，分别对应增、查、改、删。

### 定制路由
```cgo
package main

import (
	"github.com/0xNone/fox"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Person struct {
	ID   uint
	Name string
	Age  uint
}

func main() {
	var err error
	fox.DB, err = gorm.Open("sqlite3", "database.db")
	// 在绑定model route之前设置绑定组
	fox.BindGroup = fox.RawRouter.Group("/api")

	err = fox.ModelRoute(&Person{}, []Person{})
	if err != nil {
		panic(err)
	}
	fox.RawRouter.Start(":8000")
}
```
运行 `go run main.go`，启动服务

现在使用 GET、POST、PUT、DELETE 访问 http://localhost:8000/person 即可对数据进行操作。

### 取消方法
```cgo
package main

import (
	"github.com/0xNone/fox"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Person struct {
	ID   uint
	Name string
	Age  uint
}

func main() {
	var err error
	fox.DB, err = gorm.Open("sqlite3", "database.db")

	err = fox.ModelRoute(&Person{}, []Person{}, []string{"post"}...)
	if err != nil {
		panic(err)
	}
	fox.RawRouter.Start(":8000")
}
```
运行 `go run main.go`，启动服务

现在使用 GET、PUT、DELETE 访问 http://localhost:8000/person 即可对数据进行操作。

# API 文档

pass
