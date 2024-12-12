package httpserv

import (
	"VideoServ/glb"
	"github.com/jsuserapp/ju"
	"golang.org/x/net/websocket"
	"net/http"
)

func RunWs() {
	conf := glb.GetConf()
	ju.LogGreen("begin ws server in", conf.WS.Port)
	http.Handle("/ws", websocket.Handler(wsHandler))
	err := http.ListenAndServe(":"+conf.WS.Port, nil)
	ju.CheckError(err)
}
func RunWss() {
	conf := glb.GetConf()
	ju.LogGreen("begin wss server in", conf.WSS.Port)
	http.Handle("/wss", websocket.Handler(wsHandler))
	err := http.ListenAndServeTLS(":"+conf.WSS.Port, conf.TLS.Pem, conf.TLS.Key, nil)
	ju.CheckError(err)
}
func wsHandler(ws *websocket.Conn) {
	req := ws.Request()
	var userName string

	cookie, er := req.Cookie(KeyCookieId)
	if ju.CheckSuccess(er) {
		loginId := sessDb.Get(cookie.Value, KeyLoginId)
		un, ok := loginId.(string)
		if ok {
			userName = un
		}
	}
	if userName == "" {
		msg := "error not login"
		ju.LogRed(msg)
		obj := struct {
			Error string `json:"error"`
		}{Error: msg}
		_, _ = ws.Write(ju.JsonToBytes(obj, false))
		_ = ws.Close()
		return
	}
}
