//todo:如果用户注册名和返回的信息一样？？

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

var usersCache map[string]uint = make(map[string]uint)

var load uint

//须保证user的id不为0
func AuthUser(c *gin.Context) (uint, string) {
	session := sessions.Default(c)
	v := session.Get("user")
	usr, ok := v.(string)
	if usr == "" || !ok {
		return 0, ""
	}

	id, ok := usersCache[usr]
	if !ok {
		log.Println("!ok")
		return 0, usr
	}
	return id, usr
}

func SetUserSession(c *gin.Context, usr string) {
	session := sessions.Default(c)
	session.Set("user", usr)
	session.Save()
}

func OnLogin(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	var usr User

	error := c.ShouldBind(&usr)
	if error != nil {
		log.Println(error.Error())
		c.String(200, "failure")
		return
	}

	if res := CheckUser(usr.Name, usr.Password); res > 0 {
		SetUserSession(c, usr.Name)
		usersCache[usr.Name] = res
		c.String(200, "OK")
	} else {
		c.String(200, "failure")
	}
}

func OnLogup(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	var usr User

	if err := c.ShouldBind(&usr); err != nil {
		log.Println(err)
		c.String(200, "failure")
		return
	}

	if FindUser(usr.Name) != 0 {
		c.String(200, "duplicated")
		return
	}

	if err := AddUser(usr); err != nil {
		c.String(200, "failure")
	} else {
		SetUserSession(c, usr.Name)
		id := FindUser(usr.Name)
		usersCache[usr.Name] = id
		log.Println("new user Id:", id)
		c.String(200, "OK")
	}
}

func OnOptions(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.String(200, "OK")
}

type SubmitData struct {
	UserName string     `json:"username"`
	Coins    uint       `json:"coins"`
	Diamonds uint       `json:"diamonds"`
	Recs     []RecordIn `json:"recs"`
}

func OnSubmit(c *gin.Context) {

	c.Header("Access-Control-Allow-Origin", "*")
	id, usrname := AuthUser(c)

	if id == 0 {
		c.String(http.StatusOK, "unauth")
		return
	}

	var sd SubmitData

	if err := c.ShouldBindJSON(&sd); err != nil {
		log.Println("error:", err.Error())
		c.String(200, "error")
		return
	}

	if sd.UserName != usrname {
		c.String(200, "userdismatch")
		return
	}

	log.Println("uploaded:", sd.Recs)

	if err := CommitRecordIns(id, sd.Recs); err != nil {
		log.Println("error:", err.Error())
		c.String(200, "error")
	} else {
		//todo:缺错误处理
		UpdateAssets(id, sd.Coins, sd.Diamonds)
		c.String(200, "OK")
	}
}

func ServeDataNormal(c *gin.Context) {
	//支持跨域访问
	c.Header("Access-Control-Allow-Origin", "*")

	id, username := AuthUser(c)

	if id == 0 {
		c.String(http.StatusOK, "unauth")
		return
	}

	tl := TotalLearned(id)
	tasks := LoadTasks(id, load, tl)
	//log.Println("outer tasks", tasks)
	if len(tasks) == 0 {
		c.String(http.StatusOK, "nodata")
		return
	}

	coins, diamonds := QueryAssets(id)

	c.JSON(http.StatusOK, gin.H{"username": username, "learned": tl,
		"coins": coins, "diamonds": diamonds, "data": tasks})
}

func ServeDataFallible(c *gin.Context) {
	//支持跨域访问
	c.Header("Access-Control-Allow-Origin", "*")

	id, username := AuthUser(c)

	if id == 0 {
		c.String(http.StatusOK, "unauth")
		return
	}

	tl := TotalLearned(id)
	tasks := Fallible(id, load)
	//log.Println("outer tasks", tasks)
	if len(tasks) == 0 {
		c.String(http.StatusOK, "nodata")
		return
	}

	coins, diamonds := QueryAssets(id)

	c.JSON(http.StatusOK, gin.H{"username": username, "learned": tl,
		"coins": coins, "diamonds": diamonds, "data": tasks})
}

func main() {
	//UpdateTaskInfoDB()
	//UpdateRecData()
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	// Logging to a file.
	//f, _ := os.Create("gin.log")
	//gin.DefaultWriter = io.MultiWriter(f)

	loadPtr := flag.Uint("load", 60, "每次练习的任务量")
	flag.Parse()

	load = *loadPtr

	if load < 10 {
		load = 10
	}

	log.Println("每次练习的任务量为", load)

	router := gin.Default()

	defer CloseDB()

	store := cookie.NewStore([]byte("dwnbk"))
	store.Options(sessions.Options{Path: "/", MaxAge: 86400 * 10})

	router.Use(sessions.Sessions("gksession", store))

	router.Use(static.Serve("/", static.LocalFile("./public", true)))
	router.Use(gin.Logger())

	router.GET("/data.json", ServeDataNormal)

	router.GET("/fallible.json", ServeDataFallible)

	router.POST("/login", OnLogin)

	router.POST("/logup", OnLogup)

	router.POST("/submit", OnSubmit)

	router.OPTIONS("/*path", OnOptions)

	router.GET("sounds/:audio", func(c *gin.Context) {
		audio := c.Param("audio")
		c.Header("Access-Control-Allow-Origin", "*")
		c.File("./sounds/" + audio)
	})

	router.Run(":4000")
}
