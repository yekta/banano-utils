package blog

import (
	"fmt"
	"strings"
	"time"

	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
	blogStructs "github.com/yekta/banano-price-service/blog/structs"
	sharedUtils "github.com/yekta/banano-price-service/shared"
)

var TYPESENSE_ADMIN_API_KEY = sharedUtils.GetEnv("TYPESENSE_ADMIN_API_KEY")
var GHOST_API_KEY = sharedUtils.GetEnv("GHOST_API_KEY")
var GHOST_TO_MEDIUM_SECRET = sharedUtils.GetEnv("GHOST_TO_MEDIUM_SECRET")
var MEDIUM_SECRET = sharedUtils.GetEnv("MEDIUM_SECRET")
var MEDIUM_USER_ID = sharedUtils.GetEnv("MEDIUM_USER_ID")
var blogApiUrl = "https://ghost.banano.cc/ghost/api/content"
var blogPostsForSitemap blogStructs.SGhostPostsForSitemapResponse
var blogPosts blogStructs.SGhostPostsResponse
var blogSlugToPost = make(map[string]blogStructs.SGhostPost)

var fields = [...]string{
	"id",
	"title",
	"slug",
	"created_at",
	"updated_at",
	"published_at",
	"excerpt",
	"custom_excerpt",
	"feature_image",
	"featured",
	"reading_time",
}
var formats = [...]string{"html", "plaintext"}
var include = [...]string{"tags"}
var fieldsStr = strings.Join(fields[:], ",")
var formatsStr = strings.Join(formats[:], ",")
var includeStr = strings.Join(include[:], ",")
var limit = 1000
var similarsLimit = 3
var blogEndpoint = fmt.Sprintf(`%s/posts/?key=%s&fields=%s&formats=%s&include=%s&limit=%v`, blogApiUrl, GHOST_API_KEY, fieldsStr, formatsStr, includeStr, limit)
var typesenseClient = typesense.NewClient(
	typesense.WithServer("https://typesense.banano.cc"),
	typesense.WithAPIKey(TYPESENSE_ADMIN_API_KEY),
	typesense.WithConnectionTimeout(60*time.Second))
var typesenseParams = &api.ImportDocumentsParams{
	Action:    action(),
	BatchSize: batchSize(),
}

const defaultPostLimit = 15

var schema = &api.CollectionSchema{
	Name: "blog-posts",
	Fields: []api.Field{
		{
			Name:  "title",
			Type:  "string",
			Infix: newTrue(),
		},
		{
			Name: "excerpt",
			Type: "string",
		},
		{
			Name: "slug",
			Type: "string",
		},
		{
			Name: "published_at",
			Type: "int64",
		},
		{
			Name:     "custom_excerpt",
			Type:     "string",
			Optional: newTrue(),
		},
		{
			Name: "plaintext",
			Type: "string",
		},
		{
			Name:     "feature_image",
			Type:     "string",
			Optional: newTrue(),
		},
	},
	DefaultSortingField: defaultSortingField(),
}

const secondThreshold = 60
