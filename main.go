package main

import (
	"errors"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

func main() {
	db := connectDb()
	r := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("session", store))

	r.Static("/assets", "./assets")
	r.Static("/uploads", "./uploads")
	r.LoadHTMLGlob("templates/*")

	// frontend
	r.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		uid := session.Get("uid")
		if uid == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		isAdmin := session.Get("isAdmin").(bool)
		news := getAllNews(db)
		popular, _ := getOrderNews(db)
		var outNews []RenderNews
		var outPopuler []News

		for _, mynew := range news {
			if mynew.IsShow || isAdmin {
				_, author := userChange(db, mynew.UID, "")
				rate, _ := getAverageRate(db, mynew.ID)
				render := RenderNews{ID: mynew.ID, Title: mynew.Title, Author: author,
					Content: template.HTML(truncateHTML(mynew.Content, 100)),
					Rate:    rate, Timestamp: mynew.Timestamp}
				outNews = append(outNews, render)
			}
		}

		for _, mynew := range popular {
			if mynew.IsShow || isAdmin {
				outPopuler = append(outPopuler, mynew)
			}
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"news":    outNews,
			"popular": outPopuler,
			"isAdmin": isAdmin,
		})
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})

	r.GET("/update_user", func(c *gin.Context) {
		session := sessions.Default(c)
		uid := session.Get("uid")
		if uid == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}
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
		isAdmin := session.Get("isAdmin").(bool)
		id := c.Param("id")
		news, err := getNews(db, id)
		if err != nil {
			c.HTML(http.StatusNotFound, "404.html", nil)
			return
		}
		//news.Content = n
		comments, err := getComments(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve comments"})
			return
		}

		var outComments []RenderComments
		for _, myComments := range comments {
			_, author := userChange(db, myComments.UID, "")
			outComments = append(outComments, RenderComments{Author: author, Content: myComments.Content, Timestamp: myComments.Timestamp})
		}

		popular, _ := getOrderNews(db)
		var outPopuler []News
		rate, _ := getRate(db, news.ID, uid.(int))
		for _, mynew := range popular {
			if mynew.IsShow || isAdmin {
				outPopuler = append(outPopuler, mynew)
			}
		}

		if news.IsShow || isAdmin {
			_, NewsAuthor := userChange(db, news.UID, "")
			renderNews := RenderNews{ID: news.ID, Title: news.Title, Author: NewsAuthor, Content: template.HTML(news.Content), Timestamp: news.Timestamp}
			c.HTML(http.StatusOK, "news.html", gin.H{
				"news":     renderNews,
				"isAdmin":  isAdmin,
				"popular":  outPopuler,
				"rate":     rate,
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

	r.GET("/search", func(c *gin.Context) {
		session := sessions.Default(c)
		//isAdmin := session.Get("isAdmin")
		uid := session.Get("uid")
		if uid == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		// TODO 搜索
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

	r.GET("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.SetCookie("session", "", -1, "/", "127.0.0.1", false, false)
		c.Redirect(http.StatusFound, "/login")
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

	r.POST("/api/rate", func(c *gin.Context) {
		session := sessions.Default(c)
		uid := session.Get("uid")
		if uid == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Login first!"})
			return
		}

		var rate RateNews
		if err := c.ShouldBindJSON(&rate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		rate = RateNews{UID: uid.(int), NID: rate.NID, Rate: rate.Rate}
		if updateRate(db, rate) == nil {
			c.JSON(http.StatusOK, gin.H{"message": "Rate added successfully"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to add rate"})
		}
	})

	//r.POST("/api/like", func(c *gin.Context) {
	//	session := sessions.Default(c)
	//	uid := session.Get("uid")
	//	if uid == nil {
	//		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login first!"})
	//		return
	//	}
	//
	//	var like LikeComment
	//	if err := c.ShouldBindJSON(&like); err != nil {
	//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	//	}
	//	like = LikeComment{UID: uid.(int), CID: like.CID, Value: like.Value}
	//	if addLike(db, like) == nil {
	//		c.JSON(http.StatusOK, gin.H{"message": "Like added successfully"})
	//	} else {
	//		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to add like"})
	//	}
	//})

	r.POST("/api/uploads", func(c *gin.Context) {
		session := sessions.Default(c)
		isAdmin := session.Get("isAdmin")
		if isAdmin != true {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not admin!"})
			return
		}

		f, err := c.FormFile("wangeditor-uploaded-image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errno": 1, "message": "上传失败!"})
			return
		} else {
			fileExt := strings.ToLower(path.Ext(f.Filename))
			if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".gif" && fileExt != ".jpeg" {
				c.JSON(http.StatusBadRequest, gin.H{"errno": 1, "message": "上传失败!只允许png,jpg,gif,jpeg文件"})
				return
			}
			fileName := getMd5(fmt.Sprintf("%s%s", f.Filename, time.Now().String()))
			fildDir := fmt.Sprintf("uploads/%d%s/", time.Now().Year(), time.Now().Month().String())
			_, res := os.Stat(fildDir)
			if os.IsNotExist(res) {
				os.Mkdir(fildDir, os.ModePerm)
			}
			filepath := fmt.Sprintf("%s%s%s", fildDir, fileName, fileExt)
			c.SaveUploadedFile(f, filepath)
			c.JSON(http.StatusOK, gin.H{"errno": 0, "data": gin.H{"url": "/" + filepath}})
		}
	})

	err := r.Run("0.0.0.0:12345")
	if err != nil {
		return
	}
}
