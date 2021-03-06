package main

import (
	"github.com/BurntSushi/toml"
	. "github.com/cool2645/kotori-ng/config"
	"github.com/cool2645/kotori-ng/database"
	"github.com/cool2645/kotori-ng/pluginmanager"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
	"github.com/yanzay/log"
	"net/http"
	"strconv"
	"github.com/cool2645/kotori-ng/handler"
	"github.com/cool2645/kotori-ng/status"
)

const (
	pluginPath = "plugins/"
	BaseApi    = "/"
)

var (
	r   = mux.NewRouter().StrictSlash(true)
	api = r.PathPrefix(BaseApi).Subrouter()
)

func main() {
	// Load global config
	_, err := toml.DecodeFile("config.toml", &GlobCfg)
	if err != nil {
		panic(err)
	}
	// Init DB connection
	db, err := gorm.Open("sqlite3", GlobCfg.DB_FILE)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Infof("Database init done")
	database.DB = db

	// Init global router
	// Strict slash
	r.StrictSlash(GlobCfg.USE_STRICT_SLASH)
	// 404
	r.NotFoundHandler = http.HandlerFunc(handler.NotFoundHandler)
	r.MethodNotAllowedHandler = http.HandlerFunc(handler.MethodNotAllowedHandler)
	log.Infof("Router init done")

	// Load base services
	RegRouter()
	InitDB(database.DB)

	// Load plugins
	pluginmanager.PM = pluginmanager.NewPluginManager(pluginPath, api, database.DB)
	err = pluginmanager.PM.LoadPlugins()
	if err != nil {
		log.Fatal(err)
		return
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   GlobCfg.ALLOW_ORIGIN,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
		AllowCredentials: true,
	})
	h := c.Handler(r)

	n := negroni.New()
	n.Use(status.Stat)
	n.UseHandler(h)

	log.Infof("Start listening at port %d", GlobCfg.PORT)
	http.ListenAndServe(":"+strconv.FormatInt(GlobCfg.PORT, 10), n)
	return
}
