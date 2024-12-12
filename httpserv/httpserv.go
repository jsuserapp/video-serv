package httpserv

import (
	"VideoServ/glb"
	"crypto/tls"
	"github.com/jsuserapp/ju"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/kataras/iris/v12/sessions/sessiondb/boltdb"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var sessDb *boltdb.Database

func init() {
	sessDb, _ = boltdb.New("./data/sessions.db", 0666)
}
func RunHttp(app *iris.Application) {
	conf := glb.GetConf()
	ju.LogGreen("begin http server in", conf.HTTP.Port)
	err := app.Run(
		iris.Addr(":"+conf.HTTP.Port),
		iris.WithConfiguration(iris.Configuration{
			PostMaxMemory: 32 << 22,
		}),
		iris.WithoutServerError(iris.ErrServerClosed), // Ignores err server closed log when CTRL/CMD+C pressed.
		iris.WithOptimizations,                        // Enables faster json serialization and more.
	)
	ju.CheckError(err)
}
func RunHttps(app *iris.Application) {
	conf := glb.GetConf()
	cer, err := tls.LoadX509KeyPair(conf.TLS.Pem, conf.TLS.Key)
	if ju.CheckFailure(err) {
		return
	}
	ju.LogGreen("begin https server in ", conf.TLS.Port)
	srv := &http.Server{
		Addr:         ":" + conf.TLS.Port,
		Handler:      app.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cer},
		},
	}

	//使用缺省的 server 会导致 80 端口自动运行转发服务器
	err = app.Run(
		iris.Server(srv),
		iris.WithConfiguration(iris.Configuration{
			PostMaxMemory: 32 << 22,
		}),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)
}

func initSessionDb() *sessions.Sessions {
	iris.RegisterOnInterrupt(func() {
		_ = sessDb.Close()
	})

	sessManager := sessions.New(sessions.Config{
		Cookie:  KeyCookieId,
		Expires: 0,
	})
	sessManager.UseDatabase(sessDb)
	return sessManager
}

func StartHttpServ() {
	conf := glb.GetConf()
	if conf.TLS.Port == "" {
		conf.TLS.Port = "443"
	}
	if conf.HTTP.Port == "" {
		conf.HTTP.Port = "80"
	}
	ju.CreateFolder("./html")
	irApp := iris.New()

	sessManager := initSessionDb()

	cf := ControllerFatory{
		App:  irApp,
		Sess: sessManager,
	}

	cf.Create("/api", new(ApiController), false)
	cf.Create("/video", new(VideoController), false)
	irApp.Get("/", func(ctx iris.Context) {
		err := ctx.CompressWriter(true)
		ju.CheckError(err)
		err = ctx.ServeFile("./html/index.html")
		ju.CheckError(err)
	})
	irApp.Get("/{path:path}", func(ctx iris.Context) {
		err := ctx.CompressWriter(true)
		ju.CheckError(err)
		fn := filepath.Join("./html", ctx.Path())
		var e error
		if _, err = os.Stat(fn); os.IsNotExist(err) {
			e = ctx.ServeFile("./html/index.html")
		} else {
			e = ctx.ServeFile(fn)
		}
		if e != nil {
			ju.LogToRed("http", e.Error())
		}
	})
	if conf.WS.Enable {
		go RunWs()
	}
	if conf.WSS.Enable {
		go RunWss()
	}
	//以下只能运行一个，因为是阻塞执行的
	if conf.TLS.Enable {
		RunHttps(irApp)
	}
	if conf.HTTP.Enable {
		RunHttp(irApp)
	}
}
