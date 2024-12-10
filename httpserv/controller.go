package httpserv

import (
	"encoding/json"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"ju"
	"net/url"
	"strconv"
)

type ControllerFatory struct {
	App  *iris.Application
	Sess *sessions.Sessions
}

func (cp *ControllerFatory) Create(root string, ctrl interface{}, login bool, param ...interface{}) {
	m := mvc.Configure(cp.App.Party(root))

	count := len(param)
	p := make([]interface{}, 1, count+1)
	p[0] = cp.Sess.Start
	p = append(p, param...)
	m.Register(p...)

	m.Router.Use(func(ctx *context.Context) {
		//中间件，用于处理未登录跳转，登录后跳回
		//只处理 GET 请求
		if "GET" == ctx.Method() && login {
			//启用登录控制
		} else {
			ctx.Next()
		}
	})
	m.Handle(ctrl)
}

type Controller struct {
	Ctx     iris.Context
	Session *sessions.Session
}

func (c *Controller) Err(code int, msg string) mvc.Response {
	return mvc.Response{
		Code:    code,
		Content: []byte(msg),
	}
}
func (c *Controller) SetLogin(username string) {
	c.Session.Set(KeyLoginId, username)
}
func (c *Controller) GetUsername() string {
	return c.Session.GetString(KeyLoginId)
}
func (c *Controller) IsLoggedIn() bool {
	return c.GetUsername() != ""
}
func (c *Controller) Logout() {
	if c.GetUsername() != "" {
		c.Session.Destroy()
	}
}

type RequestProc func(js ju.JsonObject) string

// JsonRequest 分析 JsonBody 请求, 无需登录
func (c *Controller) JsonRequest(obj interface{}, proc RequestProc) mvc.Result {
	js := ju.JsonObject{}
	for {
		if obj != nil {
			data, err := c.Ctx.GetBody()
			if err != nil {
				js.SetValue(KeyErr, err.Error())
				break
			}
			if len(data) > 0 {
				err = json.Unmarshal(data, obj)
				if err != nil {
					js.SetValue(KeyErr, "数据必须以 json 格式提交")
					break
				}
			}
		}
		ers := proc(js)
		if ers != "" {
			js.SetValue(KeyErr, ers)
		}
		break
	}
	return mvc.Response{
		Text: js.String(),
	}
}

// requestData 获取请求的参数数据, 所有数据都是字符串格式
// http 传递数据共有 4 种方式:
// 1. post form, 这种方式可以使用 ctx.FormValue 获取
// 2. urlform, 即 get 方式的 url 里的参数, 这种方式可以通过 url.ParseQuery 获取
// 3. body 直接传递, 这种方式可以通过 ctx.GetBody() 获取, body 上传的时候可以说json格式, 但是有一点和其它不同,
// 返回的 map 可能会有 keys 以外的字段, 而确实部分字段, 这依赖于用户上传时的数据. 而且其它方式, 用户没有上传的数据会
// 是空串, 而且不会包含 keys 以外的字段数据
// 4. header 传递
func requestData(ctx iris.Context, keys []string) map[string]string {

	data := map[string]string{}
	hasValue := false

	// 1. 获取 POST Form 数据
	for _, key := range keys {
		val := ctx.FormValue(key)
		if val != "" {
			hasValue = true
		}
		data[key] = val
	}
	if hasValue {
		return data
	}

	// 2. 获取 URL Query 参数
	for _, key := range keys {
		val := ctx.URLParam(key)
		if val != "" {
			hasValue = true
		}
		data[key] = val
	}
	if hasValue {
		return data
	}

	// 3. 获取 Body 数据
	d, _ := ctx.GetBody()
	values, err := url.ParseQuery(string(d))
	if err == nil {
		// 如果 Body 数据是 URL 编码格式
		for _, key := range keys {
			val := values.Get(key)
			if val != "" {
				hasValue = true
			}
			data[key] = val
		}
		if hasValue {
			return data
		}
	}

	// 如果 Body 数据是 JSON 格式
	_ = json.Unmarshal(d, &data)

	// 4. 获取 Header 数据
	for _, key := range keys {
		val := ctx.GetHeader(key)
		if val != "" {
			data[key] = val
		}
	}

	return data
}

func requestLoginData(ctx iris.Context) (string, string) {
	//http 传递数据共有 4 种方式:
	//1. post form, 这种方式可以使用 ctx.FormValue 获取
	//2. urlform, 即 get 方式的 url 里的参数, 这种方式可以通过 url.ParseQuery 获取
	//3. body 直接传递, 这种方式可以通过 ctx.GetBody() 获取
	//4. header 传递, 此处没有解析
	user := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	user.Username = ctx.FormValue(KeyUsername)
	user.Password = ctx.FormValue(KeyPassword)
	if user.Username != "" || user.Password != "" {
		return user.Username, user.Password
	}

	d, _ := ctx.GetBody()

	values, err := url.ParseQuery(string(d))
	if err == nil {
		user.Username = values.Get(KeyUsername)
		user.Password = values.Get(KeyPassword)
	}
	if user.Username != "" && user.Password != "" {
		return user.Username, user.Password
	}
	_ = json.Unmarshal(d, &user)
	return user.Username, user.Password
}
func StrToInt(str string) int64 {
	i, _ := strconv.ParseInt(str, 10, 64)
	return i
}
