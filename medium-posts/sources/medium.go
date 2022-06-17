package mediumPostsSources

import (
	"encoding/xml"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/yekta/banano-price-service/medium-posts/structs"
)

func GetMediumPosts() mediumPostsStructs.MediumPosts {
 	mediumPostsURL := "https://medium.com/feed/banano"

	respMediumPosts, errMediumPosts := http.Get(mediumPostsURL)
	if errMediumPosts != nil {
		log.Fatalln(errMediumPosts)
	}
	var resultMediumPosts mediumPostsStructs.Rss
	xml.NewDecoder(respMediumPosts.Body).Decode(&resultMediumPosts)

	var mediumPosts mediumPostsStructs.MediumPosts

	var t, err = time.Parse(time.RFC1123,resultMediumPosts.Channel.LastBuildDate) 
	if err != nil {
		log.Fatalln(err)
	}

	mediumPosts.LastBuildTimestamp = t.UnixMilli()

	for _, post := range resultMediumPosts.Channel.Posts {
		var tPublish, err1 = time.Parse(time.RFC1123,post.PublishDate) 
		if err1 != nil {
			log.Fatalln(err)
		}
		var tUpdate, err2 = time.Parse(time.RFC3339,post.LastUpdateDate) 
		if err2 != nil {
			log.Fatalln(err)
		}
		split := strings.Split(post.Link, "/")
		splitFinal := strings.Split(split[len(split)-1], "?")

		p := strings.NewReader(post.ContentEncoded)
		doc, _ := goquery.NewDocumentFromReader(p)
		doc.Find("script").Each(func(i int, el *goquery.Selection) {
				el.Remove()
		})
		
		text := doc.Text()
		description := text[0:200]
		image := doc.Find("img").AttrOr("src", "")
		
		mediumPosts.Posts = append(mediumPosts.Posts, mediumPostsStructs.MediumPost{
			Title: post.Title,
			Content: post.ContentEncoded,
			Description: description,
			Image: image,
			Tags: post.Tags,
			PublishTimestamp: tPublish.UnixMilli(),
			LastUpdateTimestamp: tUpdate.UnixMilli(),
			Slug: splitFinal[0],
		})
	}
	log.Println("Medium posts done")
	
	return mediumPosts
}