package blog

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	blogStructs "github.com/yekta/banano-price-service/blog/structs"
)

func IndexBlog() {
	log.Println("IndexBlog: Started Indexing...")
	GetAndSetBlogPosts()
	IndexTypesense()
	log.Println("IndexBlog: Finished Indexing!")
}

func GetAndSetBlogPosts() error {
	log.Println("GetAndSetBlogPosts: Getting...")

	resp, err := http.Get(blogEndpoint)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}

	var ghostPosts blogStructs.SGhostPostsResponse
	json.NewDecoder(resp.Body).Decode(&ghostPosts)

	blogPosts = ghostPosts

	for _, post := range blogPosts.Posts {
		newPost := post
		newPost.Similars = []blogStructs.SGhostPost{}
		if len(newPost.Tags) < 1 {
			continue
		}
		tagSlug := newPost.Tags[0].Slug
		for _, otherPost := range blogPosts.Posts {
			if otherPost.Slug == newPost.Slug || len(otherPost.Tags) < 1 {
				continue
			}
			otherTagSlug := otherPost.Tags[0].Slug
			if otherTagSlug == tagSlug {
				newPost.Similars = append(newPost.Similars, blogStructs.SGhostPost{
					Title:         otherPost.Title,
					Slug:          otherPost.Slug,
					FeatureImage:  otherPost.FeatureImage,
					Excerpt:       otherPost.Excerpt,
					CustomExcerpt: otherPost.CustomExcerpt,
					ReadingTime:   otherPost.ReadingTime,
				})
			}
			if len(newPost.Similars) >= similarsLimit {
				break
			}
		}
		blogSlugToPost[newPost.Slug] = newPost
	}

	log.Println("GetAndSetBlogPosts: Set!")
	return err
}

func IndexTypesense() error {
	log.Println("TypesenseHandler: Started Indexing...")

	var blogPostsForTypesense []interface{}
	for _, post := range blogPosts.Posts {
		t, _ := time.Parse("2006-01-02T15:04:05.000+15:04", post.PublishedAt)
		blogPostsForTypesense = append(blogPostsForTypesense, blogStructs.SBlogPostForTypesense{
			Id:            post.Id,
			Title:         post.Title,
			Slug:          post.Slug,
			CustomExcerpt: post.CustomExcerpt,
			Excerpt:       post.Excerpt,
			PlainText:     post.Plaintext,
			FeatureImage:  post.FeatureImage,
			PublishedAt:   uint64(t.UnixMilli()),
		})
	}

	_, errDel := typesenseClient.Collection("blog-posts").Delete()

	if errDel != nil {
		log.Println("TypesenseHandler: Error deleting collection:", errDel)
	} else {
		log.Println("TypesenseHandler: Typesense collection deleted...")
	}

	_, errCreate := typesenseClient.Collections().Create(schema)
	if errCreate != nil {
		log.Printf("Got error %s", errCreate)
	} else {
		log.Println("TypesenseHandler: New Typesense collection created...")
	}

	_, errImport := typesenseClient.Collection("blog-posts").Documents().Import(blogPostsForTypesense, typesenseParams)

	if errImport != nil {
		log.Printf("Got error %s", errImport)
	} else {
		log.Printf("TypesenseHandler: Imported documents to Typesense...")
	}
	return errImport
}
