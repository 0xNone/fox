package fox

import (
	"github.com/creasty/defaults"
	"github.com/labstack/echo"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

var (
	RetCode *ResponseCode

	RawRouter *echo.Echo
	BindGroup *echo.Group

	ResponseCNMessage = map[int]string{
		0:    "成功",
		-255: "失败",
		-254: "超时",
		-253: "未知错误",
		-252: "请求过于频繁",
		-251: "此接口已不推荐使用",
		-249: "未找到",
		-248: "已存在",
		-239: "无权访问",
		-238: "权限申请失败",
		-229: "校验失败",
		-228: "缺少参数",
		-227: "缺少提交内容",
		-219: "非法参数",
		-218: "非法提交内容",
		1:    "Websocket 请求完成",}
)

func init() {
	var err error
	RawRouter = echo.New()
	RetCode, err = NewResponseCode(ResponseCNMessage)
	if err != nil {
		panic(err)
	}
}

type ResponseCode struct {
	Success          int `default:"0"`
	Failed           int `default:"-255"`
	Timeout          int `default:"-254"`
	Unknown          int `default:"-253"`
	TooFrequent      int `default:"-252"`
	Deprecated       int `default:"-251"`
	NotFound         int `default:"-249"`
	AlreadyExists    int `default:"-248"`
	PermissionDenied int `default:"-239"`
	InvalidRole      int `default:"-238"`
	CheckFailure     int `default:"-229"`
	QueryRequired    int `default:"-228"`
	PostdataRequired int `default:"-227"`
	InvalidParams    int `default:"-219"`
	InvalidPostdata  int `default:"-218"`
	WsDone           int `default:"1"`

	Message map[int]string
}

func (self *ResponseCode) ToMap(code int) map[string]interface{} {
	return map[string]interface{}{"code": code, "message": self.Message[code]}
}

func (self *ResponseCode) ToMapUsingData(code int, data interface{}) map[string]interface{} {
	return map[string]interface{}{"code": code, "message": self.Message[code], "data": data}
}

func (self *ResponseCode) ToMapUsingMessage(code int, message string) map[string]interface{} {
	return map[string]interface{}{"code": code, "message": message}
}

func (self *ResponseCode) ToMapUsingDataMessage(code int, message string, data interface{}) map[string]interface{} {
	return map[string]interface{}{"code": code, "message": message, "data": data}
}

func NewResponseCode(msg map[int]string) (*ResponseCode, error) {
	// 通过 field 的 index 获取 tag
	responseCode := &ResponseCode{}
	err := defaults.Set(responseCode)
	if err != nil {
		return nil, err
	}
	responseCode.Message = ResponseCNMessage
	return responseCode, nil
}

func Group(prefix string, m ...echo.MiddlewareFunc) (g *echo.Group) {
	return RawRouter.Group(prefix, m...)
}

func InStrSlice(str string, strSlice []string) bool {
	for i := range strSlice {
		if strings.ToUpper(strSlice[i]) == strings.ToUpper(str) {
			return true
		}
	}
	return false
}

func ModelRoute(model interface{}, modelSlice interface{}, disableApi ...string) error {
	if BindGroup == nil {
		BindGroup = RawRouter.Group("")
	}
	refValModel := reflect.ValueOf(model)
	regStruName := regexp.MustCompile(`[\w_]+$`)
	struName := regStruName.FindString(refValModel.Type().String())
	struNameLow := strings.ToLower(struName)

	gormDrive, err := NewGORMDrive(model, modelSlice)
	if err != nil {
		return err
	}

	if !InStrSlice(http.MethodPost, disableApi) {
		BindGroup.POST("/"+struNameLow, func(context echo.Context) error {
			//fmt.Println(context.FormParams())
			data, err := context.FormParams()
			if err != nil {
				return context.JSON(http.StatusBadRequest, RetCode.ToMap(RetCode.Failed))
			}
			newModel, err := gormDrive.Insert(data)
			if err != nil {
				return context.JSON(http.StatusBadRequest, RetCode.ToMapUsingMessage(RetCode.Failed, err.Error()))
			}
			return context.JSON(http.StatusCreated, RetCode.ToMapUsingData(RetCode.Success, newModel))
		})
	}

	if !InStrSlice(http.MethodGet, disableApi) {
		BindGroup.GET("/"+struNameLow, func(context echo.Context) error {
			query := context.QueryParams()
			result, resDB, err := gormDrive.Select(query)
			var count int
			resDB.Count(&count)
			respData := map[string]interface{}{"items": result, "items_count": reflect.ValueOf(result).Elem().Len(), "row_count": count}

			if err != nil {
				return context.JSON(http.StatusBadRequest, RetCode.ToMapUsingMessage(RetCode.Failed, err.Error()))
			}

			return context.JSON(http.StatusOK, RetCode.ToMapUsingData(RetCode.Success, respData))
		})
	}

	if !InStrSlice(http.MethodPut, disableApi) {
		BindGroup.PUT("/"+struNameLow, func(context echo.Context) error {
			query := context.QueryParams()
			data, err := context.FormParams()
			if err != nil {
				return context.JSON(http.StatusBadRequest, RetCode.ToMap(RetCode.Failed))
			}
			rowsAffected, err := gormDrive.Update(query, data)
			if err != nil {
				return context.JSON(http.StatusBadRequest, RetCode.ToMapUsingMessage(RetCode.Failed, err.Error()))
			}
			return context.JSON(http.StatusCreated, RetCode.ToMapUsingData(RetCode.Success, map[string]interface{}{"rows_affected": rowsAffected}))
		})
	}

	if !InStrSlice(http.MethodDelete, disableApi) {
		BindGroup.DELETE("/"+struNameLow, func(context echo.Context) error {
			query := context.QueryParams()
			rowsAffected, err := gormDrive.Delete(query)
			if err != nil {
				return context.JSON(http.StatusBadRequest, RetCode.ToMapUsingMessage(RetCode.Failed, err.Error()))
			}
			return context.JSON(http.StatusAccepted, RetCode.ToMapUsingData(RetCode.Success, map[string]interface{}{"rows_affected": rowsAffected}))
		})
	}
	return nil
}
