package main

import (
	"errors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"time"
)

func main() {
	db := connectDb()
	r := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("session", store))

	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")

	// frontend
	r.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		uid := session.Get("uid")
		if uid == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		news := getAllNews(db)
		var outNews []News

		for _, mynew := range news {
			if mynew.IsShow {
				outNews = append(outNews, mynew)
			}
		}
		c.HTML(http.StatusOK, "index.html", outNews)
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})

	r.GET("/update_user", func(c *gin.Context) {
		// TODO 鉴权
		c.HTML(http.StatusOK, "update_user.html", nil)
	})

	r.GET("/admin", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin.html", nil)
	})

	r.GET("/news/:id", func(c *gin.Context) {
		session := sessions.Default(c)
		uid := session.Get("uid")
		if uid == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		id := c.Param("id")
		news, err := getNews(db, id)
		if err != nil {
			c.HTML(http.StatusNotFound, "404.html", nil)
			return
		}
		comments, err := getComments(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve comments"})
			return
		}
		_, NewsAuthor := userChange(db, news.UID, "")

		var outComments []Render

		for _, myComments := range comments {
			_, author := userChange(db, myComments.UID, "")
			outComments = append(outComments, Render{Author: author, Content: myComments.Content, Timestamp: myComments.Timestamp})
		}

		if news.IsShow {
			c.HTML(http.StatusOK, "news.html", gin.H{
				"news":     news,
				"Author":   NewsAuthor,
				"comments": outComments,
			})
		} else {
			c.HTML(http.StatusNotFound, "404.html", nil)
		}
	})

	r.GET("/add_news", func(c *gin.Context) {
		session := sessions.Default(c)
		isAdmin := session.Get("isAdmin")
		if isAdmin != true {
			c.Redirect(http.StatusFound, "/index")
			return
		}
		c.HTML(http.StatusOK, "add.html", nil)
	})

	r.GET("/edit/:id", func(c *gin.Context) {
		session := sessions.Default(c)
		isAdmin := session.Get("isAdmin")
		if isAdmin != true {
			c.Redirect(http.StatusFound, "/index")
			return
		}
		id := c.Param("id")
		news, err := getNews(db, id)

		if err == nil {
			c.HTML(http.StatusOK, "edit.html", news)
		} else {
			c.HTML(http.StatusNotFound, "404.html", nil)
		}
	})

	// 后端
	r.POST("/api/register", func(c *gin.Context) {
		var user Users
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		match, _ := regexp.MatchString(`^[A-Za-z0-9]{3,32}$`, user.UserName)
		if !match {
			c.JSON(http.StatusOK, gin.H{"message": "invalid user name"})
		} else {
			user = Users{UserName: user.UserName, Password: getPasswordHash(user.Password)}
			err := register(db, user)
			if err == nil {
				c.JSON(http.StatusOK, gin.H{"message": "register success"})
			} else if errors.Is(err, errors.New("user already exists")) {
				c.JSON(http.StatusOK, gin.H{"message": "user exist"})
			} else {
				c.JSON(http.StatusOK, gin.H{"message": "register fail"})
			}
		}
	})

	r.POST("/api/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.SetCookie("session", "", -1, "/", "127.0.0.1", false, false)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	r.POST("/api/login", func(c *gin.Context) {
		var user Users
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		uid, err := checkUser(db, user)
		if err == nil {
			session := sessions.Default(c)
			session.Set("uid", uid)
			session.Set("isAdmin", false)
			session.Save()
			c.JSON(http.StatusOK, gin.H{"message": "login successfully"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "username or password error"})
		}
	})

	r.POST("/api/update_user", func(c *gin.Context) {
		session := sessions.Default(c)
		uid := session.Get("uid")
		if uid == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Login first!"})
			return
		}

		var user Users
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user = Users{ID: uid.(int), UserName: user.UserName, Password: user.Password}
		err := updateUser(db, user)
		if err == nil {
			session.Clear()
			session.Save()
			c.SetCookie("session", "", -1, "/", "127.0.0.1", false, false)
			c.JSON(http.StatusOK, gin.H{"message": "update success"})
		} else if errors.Is(err, errors.New("user already exists")) {
			c.JSON(http.StatusOK, gin.H{"message": "user exist"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "update fail"})
		}
	})

	r.POST("/api/admin", func(c *gin.Context) {
		var user Users
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		uid, err := checkAdmin(db, user)
		if err == nil {
			session := sessions.Default(c)
			session.Set("uid", uid)
			session.Set("isAdmin", true)
			session.Save()
			c.JSON(http.StatusOK, gin.H{"message": "login successfully"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "username or password error"})
		}
	})

	r.POST("/api/add_news", func(c *gin.Context) {
		session := sessions.Default(c)
		isAdmin := session.Get("isAdmin")
		if isAdmin != true {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not admin!"})
			return
		}

		var news News
		if err := c.ShouldBindJSON(&news); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		uid := session.Get("uid").(int)
		news = News{UID: uid, Title: news.Title, Content: news.Content, IsShow: news.IsShow, Timestamp: time.Now()}
		if addNews(db, news) == nil {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "Failed"})
		}
	})

	r.POST("/api/update_news", func(c *gin.Context) {
		session := sessions.Default(c)
		isAdmin := session.Get("isAdmin")
		if isAdmin != true {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not admin!"})
			return
		}

		var news News
		if err := c.ShouldBindJSON(&news); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		uid := session.Get("uid").(int)
		news = News{ID: news.ID, UID: uid, Title: news.Title, Content: news.Content, IsShow: news.IsShow, Timestamp: time.Now()}
		if updateNews(db, news) == nil {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "Failed"})
		}
	})

	r.POST("/api/delete_news", func(c *gin.Context) {
		session := sessions.Default(c)
		isAdmin := session.Get("isAdmin")
		if isAdmin != true {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not admin!"})
			return
		}

		var news News
		if err := c.ShouldBindJSON(&news); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if deleteNews(db, news.ID) == nil {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "Failed"})
		}
	})

	r.POST("/api/add_comment", func(c *gin.Context) {
		session := sessions.Default(c)
		uid := session.Get("uid")
		if uid == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Login first!"})
			return
		}

		var comment Comments
		if err := c.ShouldBindJSON(&comment); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		comment = Comments{UID: uid.(int), NID: comment.NID, Content: comment.Content, Timestamp: time.Now()}
		if addComment(db, comment) == nil {
			c.JSON(http.StatusOK, gin.H{"message": "Comments added successfully"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to add comment"})
		}
	})

	err := r.Run("0.0.0.0:12345")
	if err != nil {
		return
	}
}
