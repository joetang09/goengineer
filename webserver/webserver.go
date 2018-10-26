package webserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"gitlab.dev.okapp.cc/golang/utils"
)

const (
	MethodGET     = "GET"
	MethodPOST    = "POST"
	MethodPUT     = "PUT"
	MethodDELETE  = "DELETE"
	MethodHEAD    = "HEAD"
	MethodPATCH   = "PATCH"
	MethodOPTIONS = "OPTIONS"
	MethodANY     = "ANY"

	PathTag = "path"
	PermTag = "perm"

	defaultAddr  = ":8080"
	defaultMode  = DEBUG
	defaultPprof = true
	defaultHost  = "http://127.0.0.1"

	DEBUG   = "debug"
	RELEASE = "release"
)

var (
	svrs = []*http.Server{}

	config = Config{
		Addr:  defaultAddr,
		Mode:  defaultMode,
		Pprof: defaultPprof,
		Host:  defaultHost,
	}

	router *gin.Engine

	logger = log.New(os.Stderr, "", log.LstdFlags)

	registerControllerShouldBeValueType = errors.New("Register Controller Should be Value Type")

	actionType = reflect.TypeOf(func(*gin.Context) {})

	routePerms = make(map[interface{}]string)

	methodSupport = map[string]func(*gin.RouterGroup, string, func(*gin.Context)){
		MethodGET:     func(r *gin.RouterGroup, p string, h func(*gin.Context)) { r.GET(p, h) },
		MethodPOST:    func(r *gin.RouterGroup, p string, h func(*gin.Context)) { r.POST(p, h) },
		MethodPUT:     func(r *gin.RouterGroup, p string, h func(*gin.Context)) { r.PUT(p, h) },
		MethodDELETE:  func(r *gin.RouterGroup, p string, h func(*gin.Context)) { r.DELETE(p, h) },
		MethodHEAD:    func(r *gin.RouterGroup, p string, h func(*gin.Context)) { r.HEAD(p, h) },
		MethodPATCH:   func(r *gin.RouterGroup, p string, h func(*gin.Context)) { r.PATCH(p, h) },
		MethodOPTIONS: func(r *gin.RouterGroup, p string, h func(*gin.Context)) { r.OPTIONS(p, h) },
		MethodANY:     func(r *gin.RouterGroup, p string, h func(*gin.Context)) { r.Any(p, h) },
	}

	actionParser = func(f reflect.StructField, v reflect.Value, r *gin.RouterGroup) bool {

		if f.Type != actionType {
			return false
		}

		p := f.Tag.Get(PathTag)
		if p == "-" {
			return false
		}

		if p == "" {
			p = "/"
		}
		name := f.Name
		h := v.Interface().(func(*gin.Context))
		gr := guessMethod(r, 3, name, h, p) || guessMethod(r, 4, name, h, p) || guessMethod(r, 5, name, h, p) || guessMethod(r, 6, name, h, p)
		if gr {
			perm := strings.TrimSpace(f.Tag.Get(PermTag))
			if perm != "" {
				routePerms[utils.NameOfFunction(h)] = perm
			}
		}
		return gr

	}

	guessMethod = func(r *gin.RouterGroup, n int, name string, h func(*gin.Context), p string) bool {
		nameB := []byte(name)
		if len(nameB) < n {
			return false
		}
		m := nameB[:n]
		if sm, ok := methodSupport[strings.ToUpper(string(m))]; ok {
			sm(r, p, h)
			return true
		}
		return false
	}
)

type Controller interface {
	Name() string
}

type WebSocketConfig struct {
	HandshakeTimeout  int
	ReadBufferSize    int
	WriteBufferSize   int
	Subprotocols      []string
	EnableCompression bool
	CheckOrigin       bool
	PongWait          int
	WriteWait         int
	MaxMessageSize    int64
	DeadInNoPongNum   int
}

type Config struct {
	Addr       string
	Mode       string
	Pprof      bool
	Host       string
	WebSockets map[string]WebSocketConfig
}

type WebServer struct {
}

func (WebServer) Init(options ...interface{}) error {

	if len(options) != 0 {
		c, ok := options[0].(*Config)
		if ok {
			config = *c
		}
	}

	if config.Mode == DEBUG {
		gin.SetMode(DEBUG)
	}
	router = gin.New()

	return nil
}

func Use(middlewares ...gin.HandlerFunc) {
	if router == nil {
		return
	}
	router.Use(middlewares...)
}

type group struct {
	r *gin.RouterGroup
}

func (g *group) Controller(c ...Controller) {
	for _, tc := range c {

		t := reflect.TypeOf(tc)
		v := reflect.ValueOf(tc)

		fN := t.NumField()

		for i := 0; i < fN; i++ {
			actionParser(t.Field(i), v.Field(i), g.r)
		}
	}

}

func RegisterGroup(g string, middlewares ...gin.HandlerFunc) *group {

	if router == nil {
		return nil
	}

	r := &router.RouterGroup
	r = r.Group(g)
	r.Use(middlewares...)
	return &group{
		r: r,
	}
}

func RegisterController(c Controller, middlewares ...gin.HandlerFunc) {

	if router == nil {
		return
	}

	t := reflect.TypeOf(c)
	v := reflect.ValueOf(c)
	fN := t.NumField()

	router.RouterGroup.Use(middlewares...)

	for i := 0; i < fN; i++ {
		actionParser(t.Field(i), v.Field(i), &router.RouterGroup)
	}

}

func RegisterStatic(relativePath, root string) {
	if router == nil {
		return
	}
	router.Static(relativePath, root)
}

func RegisterStaticFS(relativePath string, fs http.FileSystem) {
	if router == nil {
		return
	}
	router.StaticFS(relativePath, fs)
}

func RegisterStaticFile(relativePath, filepath string) {
	if router == nil {
		return
	}
	router.StaticFile(relativePath, filepath)
}

func buildPPROFSrv(port int) *http.Server {

FINDPORT:
	if utils.PortUsed(port) {
		port++
		goto FINDPORT
	}

	h := http.NewServeMux()

	h.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	h.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	h.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	h.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	h.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	return &http.Server{Addr: ":" + strconv.Itoa(port), Handler: h}
}

func buildAppSrv() *http.Server {
	app := &http.Server{}
	app.Addr = config.Addr
	app.Handler = router

	return app
}

func (WebServer) CfgKey() string {
	return "webserver"
}
func (WebServer) CfgType() interface{} {
	return Config{}
}

func (WebServer) CfgUpdate(interface{}) {

}

func (WebServer) Stop() {
	fmt.Println("Shutdown Server.")
	for _, s := range svrs {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.Shutdown(ctx); err != nil {
				panic(fmt.Sprintf("Server Shutdown error[%v] \n", err))
			}
		}()

	}

	fmt.Println("Exist Normally.")
}

func SetLogger(l *log.Logger) {
	logger = l
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Set("Content-Type", "application/json;charset=UTF-8")

		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			originUrl, _ := url.Parse(origin)
			c.Writer.Header().Set("Access-Control-Allow-Origin", originUrl.Scheme+"://"+originUrl.Host)
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Authorization")
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, HEAD, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
	}
}

func GET(rp string, handlers ...gin.HandlerFunc) {
	if router == nil {
		return
	}
	router.GET(rp, handlers...)
}

func POST(rp string, handlers ...gin.HandlerFunc) {
	if router == nil {
		return
	}
	router.POST(rp, handlers...)
}

func PUT(rp string, handlers ...gin.HandlerFunc) {
	if router == nil {
		return
	}
	router.PUT(rp, handlers...)
}

func DELETE(rp string, handlers ...gin.HandlerFunc) {
	if router == nil {
		return
	}
	router.DELETE(rp, handlers...)
}

func HEAD(rp string, handlers ...gin.HandlerFunc) {
	if router == nil {
		return
	}
	router.HEAD(rp, handlers...)
}

func PATCH(rp string, handlers ...gin.HandlerFunc) {
	if router == nil {
		return
	}
	router.PATCH(rp, handlers...)
}

func OPTIONS(rp string, handlers ...gin.HandlerFunc) {
	if router == nil {
		return
	}
	router.OPTIONS(rp, handlers...)
}

func Any(rp string, handlers ...gin.HandlerFunc) {
	if router == nil {
		return
	}
	router.Any(rp, handlers...)
}

func RoutePerms() map[interface{}]string {
	return routePerms
}

func RoutePerm(methodName string) string {
	perm, ok := routePerms[methodName]
	if !ok {
		return ""
	}
	return perm
}

func BindRequest(c *gin.Context, obj interface{}) error {
	return c.ShouldBindWith(obj, binding.Default(c.Request.Method, c.ContentType()))
}

type AccessRecorder func(time.Time, time.Time, time.Duration, int, string, string, string)

func AccessRecordMiddleware(out AccessRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		c.Next()
		end := time.Now()
		latency := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		if raw != "" {
			path = path + "?" + raw
		}
		out(start, end, latency, statusCode, clientIP, method, path)

	}
}

func RegisterWebSocket(path string, ws string, callback WebSocketCallback) (*WebSocketController, error) {
	if router == nil {
		return nil, errors.New("router not init")
	}
	wsCfg, ok := config.WebSockets[ws]
	if !ok {
		wsCfg = WebSocketConfig{}
	}

	handler := NewWebSocketHandler(wsCfg, callback)
	router.GET(path, func(c *gin.Context) {
		handler.HandleConn(c.Writer, c.Request)
	})
	return handler.controller, nil
}

func EnablePing() {

	Any("/ping", func(*gin.Context) {})

}
