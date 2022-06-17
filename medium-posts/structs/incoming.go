package mediumPostsStructs

type Rss struct {
	Channel struct {
		LastBuildDate string `xml:"lastBuildDate"`
		Posts          []struct {
			Title string `xml:"title"`
			Link  string `xml:"link"`
			Guid  struct {
				Text        string `xml:",chardata"`
				IsPermaLink string `xml:"isPermaLink,attr"`
			} `xml:"guid"`
			Tags []string `xml:"category"`
			PublishDate  string   `xml:"pubDate"`
			LastUpdateDate  string   `xml:"updated"`
			ContentEncoded  string   `xml:"encoded"`
		} `xml:"item"`
	} `xml:"channel"`
} 
