package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"social/internal/store"
)

func Seed(store store.Storage, db *sql.DB) {
	ctx := context.Background()

	users := generateUsers(100)
	tx, _ := db.BeginTx(ctx, nil)

	for _, user := range users {
		err := store.Users.Create(ctx, tx, user)
		if err != nil {
			_ = tx.Rollback()
			log.Println("Error creating user:", err)
			return
		}
	}

	tx.Commit()

	posts := generatePosts(100, users)

	for _, post := range posts {
		err := store.Posts.Create(ctx, post)
		if err != nil {
			log.Println("Error creating post:", err)
			return
		}
	}

	commnets := generateComments(100, posts)

	for _, comment := range commnets {
		err := store.Comments.Create(ctx, comment)
		if err != nil {
			log.Println("Error creating comment:", err)
			return
		}
	}

	log.Println("Seed completed")
}

func generateUsers(count int) []*store.User {
	users := make([]*store.User, count)

	for i := 0; i < count; i++ {
		user := &store.User{
			Username: "username" + fmt.Sprint(i),
			Email:    "email" + fmt.Sprint(i) + "@example.com",
			RoleId:   1,
		}

		pass := &store.Password{}
		if err := pass.Set("password" + fmt.Sprint(i)); err != nil {
			panic(fmt.Sprintf("failed to set password for user %d: %v", i, err))
		}
		user.Password = *pass

		users[i] = user
	}

	return users
}

func generatePosts(count int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, count)
	for i := 0; i < count; i++ {
		user := users[rand.Intn(len(users))]
		posts[i] = &store.Post{
			Content: "content" + fmt.Sprint(i),
			Title:   "title" + fmt.Sprint(i),
			UserId:  user.ID,
			Tags:    []string{"tag1", "tag2", "tag3"},
			Version: 1,
		}
	}
	return posts
}

func generateComments(count int, posts []*store.Post) []*store.Comment {
	comments := make([]*store.Comment, count)
	for i := 0; i < count; i++ {
		post := posts[rand.Intn(len(posts))]
		comments[i] = &store.Comment{
			Content: "content" + fmt.Sprint(i),
			PostId:  post.ID,
			UserId:  post.UserId,
		}
	}
	return comments
}
