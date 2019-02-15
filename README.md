# FOX
![CI](https://travis-ci.org/0xNone/fox.svg?branch=master) ![Coverage Status](https://coveralls.io/repos/github/0xNone/fox/badge.svg?branch=master)

**TODO**

+ Model
  + ~~query、data 等数据进行基本的增删改查操作~~
+ View
  + ~~web 框架 echo 包装~~
  + ~~根据 model 自动创建 RUST API~~
  + 中间件
  + 返回状态码规范
+ Permission
  + 可 insert、update 用的 form data 检查
  + 可 query 的 query data 检查
  + 可读取的数据检查
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

```go
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
```go
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
```go
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

### query 查询

对各个查询进行了扩展支持

当在查询时加入逻辑操作符、对比操作符、扩展查询语句时框架会根据查询条件对结果进行控制，如：

```go
// 查询 id 大于 20 的结果集
GET http://localhost:8000/person?id.gt=20
// 响应内容
{
    "code": 0,
    "data": {
        "items": [
            {
                "ID": 21,
                "Name": "Li Na",
                "Age": 24
            },
            {
                "ID": 22,
                "Name": "Han meimei",
                "Age": 18
            },
        ],
        "items_count": 2,
        "row_count": 2
    },
    "message": "成功"
}

// 查询 id 等于 1 或 id 等于 2 或 age 小于 20 限制内容条数为5条的结果集
GET http://localhost:8000/person?id=1&id.or=2&age.or.lt=20
// 响应内容
{
    "code": 0,
    "data": {
        "items": [
            {
                "ID": 21,
                "Name": "Li Na",
                "Age": 24
            },
            {
                "ID": 22,
                "Name": "Han meimei",
                "Age": 18
            },
        ],
        "items_count": 2,
        "row_count": 2
    },
    "message": "成功"
}
```

### 插入/更新内容
插入/更新时会读取 form data 的内容，同样不区分大小写。如：
```go
// 插入 person 数据
POST http://localhost:8000/person
name=Li Lei&age=26

// 更新 person 中 id 小于 4 的数据
PUT http://localhost:8000/person?id.lt=4
name=Han Meimei&age=21
```


不区分大小写

**逻辑操作符**
+ `OR` 相当于 `OR`
+ `AND` 相当于 `AND`，默认

**对比操作符**
+ `EQ` 相当于 `=`，默认
+ `NE` 相当于 `!=`
+ `LT` 相当于 `<`
+ `LE` 相当于 `=`
+ `GT` 相当于 `>`
+ `GE` 相当于 `>=`
+ `IN` 相当于 `IN`
+ `NOT_IN` 相当于 `NOT IN`
+ `LIKE` 相当于 `LIKE`

**扩展查询语句**
+ `EXT_LIMIT` 相当于 `LIMIT`，默认20
+ `EXT_OFFSET` 相当于 `OFFSET`，默认0
+ `EXT_ORDER_BY` 相当于 `ORDER BY`
+ `EXT_UNSCOPED`，gorm 特性，仅在 delete 操作时有效

# API 文档

pass
