package mediumPostsStructs

type MediumPosts struct {
	LastBuildTimestamp int64 `json:"lastBuildTimestamp"`
	Posts []MediumPost `json:"posts"`
}

type MediumPost struct {
	Title string `json:"title"`
	Description string `json:"description"`
	Image string `json:"image"`
	Content string `json:"content"`
	Tags []string `json:"tags"`
	PublishTimestamp int64 `json:"publishTimestamp"`
	LastUpdateTimestamp int64 `json:"lastUpdateTimestamp"`
	Slug string `json:"slug"`
}