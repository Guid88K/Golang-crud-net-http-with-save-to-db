package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewPosts() *Posts {
	var arr Posts
	return &arr
}

type Post struct {
	UserId int    `json:"userId"`
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type Posts []Post

func NewComments() *Comments {
	var arr Comments
	return &arr
}

type Comment struct {
	PostId int    `json:"postId"`
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
}

type Comments []Comment

func (p *Posts) getUserComments(url string, i int) {
	number := strconv.Itoa(i)

	resp, err := http.Get(url + "/" + number)

	if err != nil {
		fmt.Println("No response from request")
	}

	body, err := ioutil.ReadAll(resp.Body)

	errGetBody := resp.Body.Close()
	if errGetBody != nil {
		return
	}

	str := string(body)

	var el Post

	_ = json.Unmarshal([]byte(str), &el)

	if el.UserId == 7 {
		*p = append(*p, el)
	}
}

func (c *Comments) getCommentByPost(id int) {
	number := strconv.Itoa(id)

	resp, err := http.Get("https://jsonplaceholder.typicode.com/comments?postId=" + number)

	if err != nil {
		fmt.Println("No response from request")
	}

	body, err := ioutil.ReadAll(resp.Body)

	err = resp.Body.Close()
	if err != nil {
		return
	}

	str := string(body)
	fmt.Println(str)

	var cm Comments

	err = json.Unmarshal([]byte(str), &cm)

	for _, com := range cm {
		*c = append(*c, com)
	}
}

func main() {
	url := "https://jsonplaceholder.typicode.com/posts"

	data, errGetData := http.Get(url)

	if errGetData != nil {
		fmt.Print("Cant get data form url")
	}

	defer func(Body io.ReadCloser) {
		errCloseDataBody := Body.Close()
		if errCloseDataBody != nil {
			return
		}
	}(data.Body)

	body, errReadData := ioutil.ReadAll(data.Body)

	if errReadData != nil {
		return
	}

	fmt.Println(string(body))

	p := NewPosts()

	for i := 1; i <= 100; i++ {
		go p.getUserComments(url, i)
	}

	fmt.Println("Подождите")
	for i := 1; i <= 10; i++ {
		time.Sleep(time.Second * 1)
		fmt.Print("...")
	}

	dsn := "user:ucUeB2dlViI48Gfk%@tcp(127.0.0.1:3306)/publish_part2?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect databases")
	}

	err = db.AutoMigrate(&Posts{}, &Comments{})
	if err != nil {
		panic("can't migrate the schema")
	}

	for _, pp := range *p {
		fmt.Println("Пост:")
		fmt.Println(pp.Id, pp.UserId, pp.Title, pp.Body)
		fmt.Println("Коменти:")

		comments := NewComments()

		comments.getCommentByPost(pp.Id)

		for _, cc := range *comments {

			if dbc := db.Create(&Comment{
				PostId: cc.PostId,
				Id:     cc.Id,
				Name:   cc.Name,
				Email:  cc.Email,
				Body:   cc.Body,
			}); dbc.Error != nil {
				panic("Can't save data")
			}
		}
	}

	fmt.Println("Done!!")

	http.HandleFunc("/create/comment", func(w http.ResponseWriter, r *http.Request) {
		handleComment(w, r, db)
	})

	http.HandleFunc("/create/post", func(w http.ResponseWriter, r *http.Request) {
		handlePost(w, r, db)
	})

	http.HandleFunc("/json/comments", func(w http.ResponseWriter, r *http.Request) {
		getComments(w, r, db)
	})

	http.HandleFunc("/xml/comments", func(w http.ResponseWriter, r *http.Request) {
		getComments(w, r, db)
	})

	http.HandleFunc("/json/posts", func(w http.ResponseWriter, r *http.Request) {
		getPosts(w, r, db)
	})

	http.HandleFunc("/xml/posts", func(w http.ResponseWriter, r *http.Request) {
		getPosts(w, r, db)
	})

	err = http.ListenAndServe(":8080", nil)

	log.Fatal(err)
}

func getPosts(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var p []Post

	db.Raw("SELECT * FROM posts").Scan(&p)

	if "/json/posts" == r.URL.Path {
		json.NewEncoder(w).Encode(&p)
	} else {
		xml.NewEncoder(w).Encode(&p)
	}
}

func handleComment(w http.ResponseWriter, r *http.Request, d *gorm.DB) {
	length := r.ContentLength
	body := make([]byte, length)
	_, err := r.Body.Read(body)
	if err != nil {
		return
	}
	var c Comment
	json.Unmarshal(body, &c)
	createComment(c, d)
	w.WriteHeader(201)
}

func handlePost(w http.ResponseWriter, r *http.Request, d *gorm.DB) {
	length := r.ContentLength
	body := make([]byte, length)
	_, err := r.Body.Read(body)
	if err != nil {
		return
	}
	var p Post
	json.Unmarshal(body, &p)
	createPost(p, d)
	w.WriteHeader(201)
}

func getComments(w http.ResponseWriter, r *http.Request, d *gorm.DB) {
	var c []Comment

	d.Raw("SELECT * FROM comments").Scan(&c)

	if "/json/comments" == r.URL.Path {
		json.NewEncoder(w).Encode(&c)
	} else {
		xml.NewEncoder(w).Encode(&c)
	}
}

func createComment(c Comment, d *gorm.DB) {
	d.Create(&Comment{
		PostId: c.PostId,
		Id:     c.Id,
		Name:   c.Name,
		Email:  c.Email,
		Body:   c.Body,
	})
}

func createPost(c Post, d *gorm.DB) {
	d.Create(&Post{
		Id:     c.Id,
		UserId: c.UserId,
		Body:   c.Body,
		Title:  c.Title,
	})
}
