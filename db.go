package main

import (
	"errors"
	"fmt"
	"os/exec"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func connectDb() (db *gorm.DB) {
	dsn := "root:root@tcp(127.0.0.1:3306)/websql?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	checkErr(err)
	err = db.AutoMigrate(&Users{})
	checkErr(err)
	err = db.AutoMigrate(&Admins{})
	checkErr(err)
	err = db.AutoMigrate(&News{})
	checkErr(err)
	return db
}

func testDb(db *gorm.DB) {
	rows, err := db.Raw("SHOW databases").Rows()
	checkErr(err)
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		fmt.Println(name)
	}
}

// TODO: Add frontend interface
func backupDb(dbName, username, password, backupFile string) error {
	cmd := fmt.Sprintf("mysqldump -u%s -p%s %s > %s", username, password, dbName, backupFile)
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return err
	}
	return nil
}

func restoreDb(dbName, username, password, backupFile string) error {
	cmd := fmt.Sprintf("mysql -u%s -p%s %s < %s", username, password, dbName, backupFile)
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return err
	}
	return nil
}

// User
func getUserPasswd(db *gorm.DB, name string) string {
	var user Users
	db.Select("password").Where("user_name = ?", name).First(&user)
	return user.Password
}

// 0-success, 1-user exist, 2-unknown error
func register(db *gorm.DB, name string, password string) int {
	user := Users{UserName: name, Password: getPasswordHash(password)}
	fmt.Println(user)
	if errors.Is(db.Where("user_name = ?", name).First(&Users{}).Error, gorm.ErrRecordNotFound) {
		res := db.Create(&user)
		if res.Error != nil {
			return 2
		} else {
			return 0
		}
	} else {
		return 1
	}
}

// 0-success, 1-user exist, 2-unknown error
func updateUser(db *gorm.DB, name string, newName string, password string) int {
	user := Users{UserName: newName, Password: getPasswordHash(password)}
	if errors.Is(db.Where("user_name = ?", newName).First(&Users{}).Error, gorm.ErrRecordNotFound) || name == newName {
		res := db.Model(&user).Where("user_name = ?", name).Updates(user)
		if res.Error != nil {
			return 2
		} else {
			return 0
		}
	} else {
		return 1
	}
}

// Admin
func getAdminPasswd(db *gorm.DB, name string) string {
	var admin Admins
	db.Select("password").Where("user_name = ?", name).First(&admin)
	return admin.Password
}

// 0-success, 1-user exist, 2-unknown error
//func updateAdmin(db *gorm.DB, name string, newName string, password string) int {
//	var admin = Admins{UserName: newName, Password: password}
//	if errors.Is(db.Where("user_name = ?", name).First(&Admins{}).Error, gorm.ErrRecordNotFound) {
//		res := db.Model(&admin).Where("user_name = ?", name).Updates(admin)
//		if res.Error != nil {
//			return 2
//		} else {
//			return 0
//		}
//	} else {
//		return 1
//	}
//}

// News
func addNews(db *gorm.DB, title string, content string, isShow bool, author string) bool {
	news := News{Title: title, Content: content, Time: time.Now(), IsShow: isShow, Author: author}
	res := db.Create(&news)
	if res.Error == nil {
		return true
	} else {
		return false
	}
}

func updateNews(db *gorm.DB, id int, newTitle string, content string, isShow bool) bool {
	news := News{Title: newTitle, Content: content, IsShow: isShow}
	res := db.Model(&news).Where("id = ?", id).Select("title", "content", "is_show").Updates(news)
	if res.Error == nil {
		return true
	} else {
		return false
	}
}

func deleteNews(db *gorm.DB, id int) bool {
	res := db.Where("id = ?", id).Delete(&News{})
	if res.Error == nil {
		return true
	} else {
		return false
	}
}

func getAllNews(db *gorm.DB) []News {
	var news []News
	db.Find(&news)
	return news
}

func getNews(db *gorm.DB, id string) (News, error) {
	var news News
	res := db.Where("id = ?", id).First(&news)
	err := res.Error
	return news, err
}

// Comment
func addComment(db *gorm.DB, userID int, newsID int, content string) bool {
	comment := Comment{UserID: userID, NewsID: newsID, Content: content, Timestamp: time.Now()}
	res := db.Create(&comment)
	if res.Error == nil {
		return true
	} else {
		return false
	}
}

func getComments(db *gorm.DB, newsID string) ([]Comment, error) {
	var comments []Comment
	res := db.Where("news_id = ?", newsID).Find(&comments)
	return comments, res.Error
}
