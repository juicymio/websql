package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
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
		user := session.Get("user")
		if user == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		news := getAllNews(db)
		out_news := []News{}
		// nid := []int{}

		for _, mynew := range news {
			if mynew.IsShow {
				// titles = append(titles, mynew.Title)
				// nid = append(nid, mynew.Id)
				out_news = append(out_news, mynew)
			}
		}
		fmt.Println(out_news)
		c.HTML(http.StatusOK, "index.html", out_news)
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
		user := session.Get("user")
		if user == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		id := c.Param("id")
		news, err := getNews(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve news article"})
			return
		}

		comments, err := getComments(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve comments"})
			return
		}

		if news.IsShow {
			c.HTML(http.StatusOK, "news.html", map[string]interface{}{
				"news":     news,
				"comments": comments,
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
		res := register(db, user.UserName, user.Password)
		if res == 0 {
			c.JSON(http.StatusOK, gin.H{"message": "register success"})
		} else if res == 1 {
			c.JSON(http.StatusOK, gin.H{"message": "user exist"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "register fail"})
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

		if getPasswordHash(user.Password) == getUserPasswd(db, user.UserName) {
			session := sessions.Default(c)
			session.Set("user", user.UserName)
			session.Set("isAdmin", false)
			session.Save()
			c.JSON(http.StatusOK, gin.H{"message": "login successfully"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "username or password error"})
		}
	})

	r.POST("/api/update_user", func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("user")
		if username == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Login first!"})
			return
		}

		var user Users
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res := updateUser(db, username.(string), user.UserName, user.Password)
		if res == 0 {
			session.Clear()
			session.Save()
			c.SetCookie("session", "", -1, "/", "127.0.0.1", false, false)
			c.JSON(http.StatusOK, gin.H{"message": "update success"})
		} else if res == 1 {
			c.JSON(http.StatusOK, gin.H{"message": "user exist"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "register fail"})
		}
	})

	r.POST("/api/admin", func(c *gin.Context) {
		var admin Admins
		if err := c.ShouldBindJSON(&admin); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if getPasswordHash(admin.Password) == getAdminPasswd(db, admin.UserName) {
			session := sessions.Default(c)
			session.Set("user", admin.UserName)
			session.Set("isAdmin", true)
			session.Save()
			c.JSON(http.StatusOK, gin.H{"message": "login successfully"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "username or password error"})
		}
	})

	//r.POST("/api/update_admin", func(c *gin.Context) {
	//	session := sessions.Default(c)
	//	username := session.Get("user")
	//	isAdmin := session.Get("isAdmin")
	//	if isAdmin != true {
	//		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login first!"})
	//		return
	//	}
	//
	//	var admin Admins
	//	if err := c.ShouldBindJSON(&admin); err != nil {
	//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	//		return
	//	}
	//	res := updateAdmin(db, username.(string), admin.UserName, admin.Password)
	//	if res == 0 {
	//		c.JSON(http.StatusOK, gin.H{"message": "update success"})
	//	} else if res == 1 {
	//		c.JSON(http.StatusOK, gin.H{"message": "user exist"})
	//	} else {
	//		c.JSON(http.StatusOK, gin.H{"message": "register fail"})
	//	}
	//})

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
		author := session.Get("user").(string)
		res := addNews(db, news.Title, news.Content, news.IsShow, author)
		if res {
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
		res := updateNews(db, news.Id, news.Title, news.Content, news.IsShow)
		if res {
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
		res := deleteNews(db, news.Id)
		if res {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "Failed"})
		}
	})

	r.POST("/api/add_comment", func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user").(int)
		newsID_str := c.PostForm("news_id")
		newsID, err := strconv.Atoi(newsID_str)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "news_id should be an integer"})
			return
		}
		content := c.PostForm("content")

		if addComment(db, userID, newsID, content) {
			c.JSON(http.StatusOK, gin.H{"message": "Comment added successfully"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to add comment"})
		}
	})

	err := r.Run("0.0.0.0:12345")
	if err != nil {
		return
	}
}
