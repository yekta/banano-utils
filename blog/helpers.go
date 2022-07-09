package blog

import (
	"fmt"
	"strings"

	blogStructs "github.com/yekta/banano-price-service/blog/structs"
	"golang.org/x/exp/slices"
)

func GhostToMediumHtmlConverter(html string, title string) string {
	s1 := strings.ReplaceAll(html, "<h2", "<h1")
	s2 := strings.ReplaceAll(s1, "</h2>", "</h1>")
	s3 := strings.ReplaceAll(s2, "<h3", "<h2")
	s4 := strings.ReplaceAll(s3, "</h3>", "</h2>")
	resHtml := fmt.Sprintf(`<h1>%s</h1>%s`, title, s4)
	return resHtml
}

func newTrue() *bool {
	b := true
	return &b
}

func defaultSortingField() *string {
	f := "published_at"
	return &f
}

func batchSize() *int {
	f := 500
	return &f
}

func action() *string {
	f := "create"
	return &f
}

func filterByFields(post blogStructs.SGhostPost, fields []string) blogStructs.SGhostPost {
	if !slices.Contains(fields, "id") {
		post.Id = ""
	}
	if !slices.Contains(fields, "slug") {
		post.Slug = ""
	}
	if !slices.Contains(fields, "title") {
		post.Title = ""
	}
	if !slices.Contains(fields, "published_at") {
		post.PublishedAt = ""
	}
	if !slices.Contains(fields, "excerpt") {
		post.Excerpt = ""
	}
	if !slices.Contains(fields, "custom_excerpt") {
		post.CustomExcerpt = ""
	}
	if !slices.Contains(fields, "html") {
		post.Html = ""
	}
	if !slices.Contains(fields, "plaintext") {
		post.Plaintext = ""
	}
	if !slices.Contains(fields, "slug") {
		post.Slug = ""
	}
	if !slices.Contains(fields, "feature_image") {
		post.FeatureImage = ""
	}
	if !slices.Contains(fields, "created_at") {
		post.CreatedAt = ""
	}
	if !slices.Contains(fields, "updated_at") {
		post.UpdatedAt = ""
	}
	if !slices.Contains(fields, "reading_time") {
		post.ReadingTime = 0
	}
	if !slices.Contains(fields, "tags") {
		post.Tags = []blogStructs.SGhostPostTag{}
	}
	if !slices.Contains(fields, "similars") {
		post.Similars = []blogStructs.SGhostPost{}
	}
	return post
}
