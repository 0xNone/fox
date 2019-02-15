package fox

import (
	"github.com/labstack/echo"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

var (
	// Echo 路由对象，默认会进行初始化
	RawRouter *echo.Echo
	// 用于绑定 api 的组，默认路径为空
	BindGroup *echo.Group

	// 默认的响应状态码与消息
	DefaultMessage = map[int]interface{}{
		http.StatusOK:           "Success",
		http.StatusCreated:      "Creation/update succeeded",
		http.StatusAccepted:     "Successful operation",
		http.StatusNoContent:    "",
		http.StatusBadRequest:   "Invalid request",
		http.StatusUnauthorized: "Unauthorized access",
		http.StatusForbidden:    "Access denied",
		http.StatusConflict:     "There are conflicts",
	}
)

func InStrSlice(str string, strSlice []string) bool {
	for i := range strSlice {
		if strings.ToUpper(strSlice[i]) == strings.ToUpper(str) {
			return true
		}
	}
	return false
}
func init() {
	RawRouter = echo.New()
}

// 新建一个 ModelView
// model 是一个结构体指针，modelSlice 为结构体切片
// api 会直接对 model 在数据库中的数据进行操作
func NewModelView(model interface{}, modelSlice interface{}) (*ModelView, error) {
	gormDrive, err := NewGORMDrive(model, modelSlice)
	if err != nil {
		return nil, err
	}

	refValModel := reflect.ValueOf(model)
	regStruName := regexp.MustCompile(`[\w_]+$`)
	struName := regStruName.FindString(refValModel.Type().String())
	struNameLow := strings.ToLower(struName)
	m := &ModelView{Name: struNameLow, Message: DefaultMessage, gormDrive: gormDrive}
	return m, nil
}

// ModelView 用于创建增删改查相关 api
// Name 是 api 名即：/xxx，在应用 api 前可进行修改，默认为 model 名称（小写）
type ModelView struct {
	Name      string
	Message   map[int]interface{}
	gormDrive *GORMDrive
}

func (mv *ModelView) InitBindGroup() {
	if BindGroup == nil {
		BindGroup = RawRouter.Group("")
	}
}

func (mv *ModelView) EnablePOST() {
	mv.InitBindGroup()
	BindGroup.POST("/"+mv.Name, mv.FuncPOST)
}

func (mv *ModelView) EnableDELETE() {
	mv.InitBindGroup()
	BindGroup.DELETE("/"+mv.Name, mv.FuncDELETE)
}

func (mv *ModelView) EnablePUT() {
	mv.InitBindGroup()
	BindGroup.PUT("/"+mv.Name, mv.FuncPUT)
}

func (mv *ModelView) EnableGET() {
	mv.InitBindGroup()
	BindGroup.GET("/"+mv.Name, mv.FuncGET)
}

func (mv *ModelView) FuncPOST(context echo.Context) error {
	data, err := context.FormParams()
	if err != nil {
		return context.JSON(mv.GenRetMapWithMsg(http.StatusBadRequest, err.Error()))
	}
	newModel, err := mv.gormDrive.Insert(data)
	if err != nil {
		return context.JSON(mv.GenRetMapWithMsg(http.StatusBadRequest, err.Error()))
	}
	return context.JSON(mv.GenRetMapWithData(http.StatusAccepted, newModel))
}

func (mv *ModelView) FuncDELETE(context echo.Context) error {
	query := context.QueryParams()
	rowsAffected, err := mv.gormDrive.Delete(query, context.QueryString())
	if err != nil {
		return context.JSON(mv.GenRetMapWithMsg(http.StatusBadRequest, err.Error()))
	}
	return context.JSON(mv.GenRetMapWithData(http.StatusAccepted, rowsAffected))
}

func (mv *ModelView) FuncPUT(context echo.Context) error {
	query := context.QueryParams()
	data, err := context.FormParams()
	if err != nil {
		return context.JSON(mv.GenRetMapWithMsg(http.StatusBadRequest, err.Error()))
	}
	rowsAffected, err := mv.gormDrive.Update(query, data, context.QueryString())
	if err != nil {
		return context.JSON(mv.GenRetMapWithMsg(http.StatusBadRequest, err.Error()))
	}
	return context.JSON(mv.GenRetMapWithData(http.StatusCreated, rowsAffected))
}

func (mv *ModelView) FuncGET(context echo.Context) error {
	query := context.QueryParams()
	result, resDB, err := mv.gormDrive.Select(query, context.QueryString())
	var count int
	resDB.Count(&count)
	respData := map[string]interface{}{"items": result, "items_count": reflect.ValueOf(result).Elem().Len(), "row_count": count}

	if err != nil {
		return context.JSON(mv.GenRetMapWithMsg(http.StatusBadRequest, err.Error()))
	}
	return context.JSON(mv.GenRetMapWithData(http.StatusOK, respData))
}

func (mv *ModelView) GenRetMapWithData(statusCode int, data interface{}) (int, map[string]interface{}) {
	retVal := mv.GenResponeMap(statusCode)
	retVal["data"] = data
	return statusCode, retVal
}

func (mv *ModelView) GenRetMapWithMsg(statusCode int, message interface{}) (int, map[string]interface{}) {
	retVal := mv.GenResponeMap(statusCode)
	if _, ok := retVal["message"]; ok {
		retVal["message"] = message
	} else {
		retVal["error"] = message
	}
	return statusCode, retVal
}

func (mv *ModelView) GenRetMapWithMsgData(statusCode int, message interface{}, data interface{}) (int, map[string]interface{}) {
	_, retVal := mv.GenRetMapWithMsg(statusCode, message)
	retVal["data"] = data
	return statusCode, retVal
}

func (mv *ModelView) GenResponeMap(statusCode int) (map[string]interface{}) {
	respData := map[string]interface{}{}
	var message interface{}
	if v, ok := mv.Message[statusCode]; ok {
		message = v
	} else {
		message = "Unknow status"
	}
	if statusCode < 400 {
		respData["message"] = message
	} else {
		respData["error"] = message
	}
	return respData
}

func (mv *ModelView) EnableDefault() {
	mv.EnablePOST()
	mv.EnableDELETE()
	mv.EnablePUT()
	mv.EnableGET()
}
