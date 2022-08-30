package blog

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	blogStructs "github.com/yekta/banano-price-service/blog/structs"
)

func IndexBlog(initial bool) {
	start := time.Now()

	log.Println("-- IndexBlog: Started Indexing...")

	GetAndSetBlogPosts()

	var wg sync.WaitGroup
	if !initial {
		wg.Add(1)
		go func() {
			defer wg.Done()
			TriggerDeploys()
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		IndexTypesense()
	}()
	wg.Wait()

	log.Printf("-- IndexBlog: Finished Indexing in %s!", time.Since(start))
}

func GetAndSetBlogPosts() error {
	start := time.Now()
	log.Println("GetAndSetBlogPosts: Getting...")

	resp, err := http.Get(blogEndpoint)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}

	var ghostPosts blogStructs.SGhostPostsResponse
	json.NewDecoder(resp.Body).Decode(&ghostPosts)

	blogPosts = ghostPosts

	for index, post := range blogPosts.Posts {
		// set feature image if post doesn't have one
		if post.FeatureImage == "" {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(post.Html))
			if err != nil {
				log.Fatal(err)
			}
			img := doc.Find("img").First()
			post.FeatureImage = img.AttrOr("src", "")
			blogPosts.Posts[index].FeatureImage = post.FeatureImage
		}
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

	log.Printf("GetAndSetBlogPosts: Set in %s!", time.Since(start))
	return err
}

func IndexTypesense() error {
	start := time.Now()
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

	log.Printf("TypesenseHandler: Finished Indexing in %s!", time.Since(start))
	return errImport
}

func TriggerDeploys() {
	log.Printf(`TriggerDeploys: Started...`)
	endpoints := []blogStructs.WebhookEndpoint{
		{
			Name: "Cloudflare Pages",
			Url:  "https://api.cloudflare.com/client/v4/pages/webhooks/deploy_hooks/7fb014e8-6f1a-420f-89a7-919693ac5337",
		},
		{
			Name: "Vercel",
			Url:  "https://api.vercel.com/v1/integrations/deploy/prj_cR1PYJ509eWSNjFaV58m3UxODzWX/nXlojEcSZu",
		},
		{
			Name: "Netlify",
			Url:  "https://api.netlify.com/build_hooks/62caf49d858ea74e4d4dc3de",
		},
	}

	for _, endpoint := range endpoints {
		go TriggerDeploy(endpoint)
	}

	log.Printf(`TriggerDeploys: Ended!`)
}

func TriggerDeploy(endpoint blogStructs.WebhookEndpoint) {
	log.Printf(`TriggerDeploy: Started for "%s"...`, endpoint.Name)
	_, err := http.Post(endpoint.Url, "application/json", nil)
	if err != nil {
		log.Printf(`TriggerDeploy: Got error "%s"`, err)
	} else {
		log.Printf(`TriggerDeploy: Success for "%s"!`, endpoint.Name)
	}
}
