package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type indexPage struct {
	FeaturedPosts []postCardData
	RecentPosts   []postCardData
}

type postCardData struct {
	PostID      string `db:"post_id"`
	Headline    string `db:"title"`
	Subheadline string `db:"subtitle"`
	AuthorName  string `db:"author_name"`
	AuthorPhoto string `db:"author_photo_url"`
	PublishDate string `db:"publish_date"`
	ImageUrl    string `db:"image_url"`
	LabelText   string `db:"label_text"`
	Featured    byte   `db:"featured"`
	Content     string `db:"content"`
}

type postData struct {
	Headline    string `db:"title"`
	Subheadline string `db:"subtitle"`
	PostImg     string `db:"image_url"`
	Content     string `db:"content"`
}

func index(db *sqlx.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		featuredPosts, recentPosts, err := getPosts(db)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		ts, err := template.ParseFiles("pages/index.html")
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err.Error())
			return
		}

		data := indexPage{
			FeaturedPosts: featuredPosts,
			RecentPosts:   recentPosts,
		}

		err = ts.Execute(w, data)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err.Error())
			return
		}

		log.Println("Request completed successfully")
	}
}

func getPosts(db *sqlx.DB) ([]postCardData, []postCardData, error) {
	const queryFeatured = `
		SELECT
			post_id,
			title,
			subtitle,
			author_name,
			author_photo_url,
			publish_date,
			image_url,
			label_text,
			content
		FROM
			post
		WHERE featured = 1
	`

	const queryRecent = `
		SELECT
			post_id,
			title,
			subtitle,
			author_name,
			author_photo_url,
			publish_date,
			image_url,
			label_text,
			content
		FROM
			post
		WHERE featured = 0
	`

	var featuredPosts []postCardData
	var recentPosts []postCardData

	errorFeatured := db.Select(&featuredPosts, queryFeatured)
	if errorFeatured != nil {
		return nil, nil, errorFeatured
	}

	errorRecent := db.Select(&recentPosts, queryRecent)
	if errorRecent != nil {
		return nil, nil, errorRecent
	}

	return featuredPosts, recentPosts, nil
}

func post(db *sqlx.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		postIDStr := mux.Vars(r)["postID"]

		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.Error(w, "Invalid post id", 403)
			log.Println(err)
			return
		}

		order, err := postByID(db, postID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Post not found", 404)
				log.Println(err)
				return
			}

			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		ts, err := template.ParseFiles("pages/post.html")
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		err = ts.Execute(w, order)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		log.Println("Request completed successfully")
	}
}

func postByID(db *sqlx.DB, postID int) (postData, error) {
	const query = `
		SELECT
			title,
			subtitle,
			image_url,
			content
		FROM
			` + "`post`" + `
		WHERE
			post_id = ?
	`

	var post postData

	err := db.Get(&post, query, postID)
	if err != nil {
		return postData{}, err
	}

	return post, nil
}
