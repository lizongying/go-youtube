package youtubeSpider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lizongying/go-youtube/internal/config"
	"github.com/lizongying/go-youtube/internal/logger"
	"github.com/lizongying/go-youtube/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type YoutubeSpider struct {
	proxy                 string
	timeout               time.Duration
	collectionYoutubeUser *mongo.Collection
	logger                *logger.Logger
	client                *http.Client

	urlSearch    string
	urlSearchApi string
	urlUserApi   string

	apiKey          string
	searchJsonRe    *regexp.Regexp
	apiKeyRe        *regexp.Regexp
	emailRe         *regexp.Regexp
	urlRe           *regexp.Regexp
	floatRe         *regexp.Regexp
	intRe           *regexp.Regexp
	publishedTimeRe *regexp.Regexp
}

func (y *YoutubeSpider) getClient() (err error) {
	proxy, err := url.Parse(y.proxy)
	if err != nil {
		y.logger.Error(err)
		return
	}

	tr := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}
	y.client = &http.Client{
		Transport: tr,
		Timeout:   y.timeout,
	}

	return
}

func (y *YoutubeSpider) Search(ctx context.Context, meta MetaSearch) (err error) {
	y.logger.Info("Search", utils.JsonStr(meta))

	if ctx == nil {
		ctx = context.Background()
	}

	keyword := url.QueryEscape(meta.Keyword)
	req, err := http.NewRequest("GET", fmt.Sprintf(y.urlSearch, keyword), nil)

	if err != nil {
		y.logger.Error(err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

	resp, err := y.client.Do(req)
	if err != nil {
		y.logger.Error(err)
		return
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		y.logger.Error(err)
		return
	}
	r := y.searchJsonRe.FindSubmatch(body)
	if len(r) != 2 {
		err = errors.New("not find content")
		y.logger.Error(err)
		return
	}
	var respSearch RespSearch
	err = json.Unmarshal(r[1], &respSearch)
	if err != nil {
		y.logger.Error(err)
		return
	}
	token := ""
	for _, v := range respSearch.Contents.TwoColumnSearchResultsRenderer.PrimaryContents.SectionListRenderer.Contents {
		continuationCommand := v.ContinuationItemRenderer.ContinuationEndpoint.ContinuationCommand
		if continuationCommand.Request == "CONTINUATION_REQUEST_TYPE_SEARCH" {
			token = continuationCommand.Token
		} else {
			for _, v1 := range v.ItemSectionRenderer.Contents {
				if v1.VideoRenderer.VideoID == "" {
					continue
				}

				runs := v1.VideoRenderer.OwnerText.Runs
				if len(runs) < 1 {
					y.logger.Error("runs err")
					continue
				}
				e := y.UserApi(context.Background(), MetaUser{
					KeyWord:  meta.Keyword,
					Id:       strings.TrimPrefix(runs[0].NavigationEndpoint.BrowseEndpoint.CanonicalBaseURL, "/@"),
					Key:      runs[0].NavigationEndpoint.BrowseEndpoint.BrowseID,
					UserName: runs[0].Text,
				})
				if e != nil {
					y.logger.Error(e)
					continue
				}
			}
		}
	}

	r = y.apiKeyRe.FindSubmatch(body)
	if len(r) != 2 {
		err = errors.New("not find api-key")
		y.logger.Error(err)
		return
	}

	y.apiKey = string(r[1])

	meta.Page++
	if meta.MaxPage > 0 && meta.Page > meta.MaxPage {
		y.logger.Info("max page")
		return
	}
	meta.NextPageToken = token
	err = y.SearchApi(context.Background(), meta)
	if err != nil {
		y.logger.Error(err)
		return
	}

	return
}

func (y *YoutubeSpider) SearchApi(ctx context.Context, meta MetaSearch) (err error) {
	y.logger.Info("SearchApi", utils.JsonStr(meta))

	if ctx == nil {
		ctx = context.Background()
	}

	bs := []byte(fmt.Sprintf(`{"context":{"client":{"hl":"en","gl":"US","clientName":"WEB","clientVersion":"2.20230327.01.00"}},"continuation":"%s"}`, meta.NextPageToken))
	req, err := http.NewRequest("POST", fmt.Sprintf(y.urlSearchApi, y.apiKey), bytes.NewReader(bs))

	if err != nil {
		y.logger.Error(err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

	resp, err := y.client.Do(req)
	if err != nil {
		y.logger.Error(err)
		return
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		y.logger.Error(err)
		return
	}

	var respSearch RespSearchApi
	err = json.Unmarshal(body, &respSearch)
	if err != nil {
		y.logger.Error(err)
		return
	}

	token := ""
	onResponseReceivedCommands := respSearch.OnResponseReceivedCommands
	if len(onResponseReceivedCommands) < 1 {
		err = errors.New("onResponseReceivedCommands err")
		y.logger.Error(err)
		return
	}

	for _, v := range onResponseReceivedCommands[0].AppendContinuationItemsAction.ContinuationItems {
		continuationCommand := v.ContinuationItemRenderer.ContinuationEndpoint.ContinuationCommand
		if continuationCommand.Request == "CONTINUATION_REQUEST_TYPE_SEARCH" {
			token = continuationCommand.Token
		} else {
			for _, v1 := range v.ItemSectionRenderer.Contents {
				if v1.VideoRenderer.VideoID == "" {
					continue
				}

				runs := v1.VideoRenderer.OwnerText.Runs
				if len(runs) < 1 {
					y.logger.Error("runs err")
					continue
				}
				e := y.UserApi(context.Background(), MetaUser{
					KeyWord:  meta.Keyword,
					Id:       strings.TrimPrefix(runs[0].NavigationEndpoint.BrowseEndpoint.CanonicalBaseURL, "/@"),
					Key:      runs[0].NavigationEndpoint.BrowseEndpoint.BrowseID,
					UserName: runs[0].Text,
				})
				if e != nil {
					y.logger.Error(e)
					continue
				}
			}
		}
	}

	if token != "" {
		meta.Page++
		if meta.MaxPage > 0 && meta.Page > meta.MaxPage {
			y.logger.Info("max page")
			return
		}
		meta.NextPageToken = token
		err = y.SearchApi(context.Background(), meta)
		if err != nil {
			y.logger.Error(err)
			return
		}
	}

	return
}

func (y *YoutubeSpider) UserApi(ctx context.Context, meta MetaUser) (err error) {
	y.logger.Info("UserApi", utils.JsonStr(meta))

	if ctx == nil {
		ctx = context.Background()
	}

	bs := []byte(fmt.Sprintf(`{"context":{"client":{"hl":"en","gl":"US","clientName":"WEB","clientVersion":"2.20230327.01.00"}},"browseId":"%s"}`, meta.Key))
	req, err := http.NewRequest("POST", fmt.Sprintf(y.urlUserApi, y.apiKey), bytes.NewReader(bs))

	if err != nil {
		y.logger.Error(err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

	resp, err := y.client.Do(req)
	if err != nil {
		y.logger.Error(err)
		return
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		y.logger.Error(err)
		return
	}

	var respUser RespUserApi
	err = json.Unmarshal(body, &respUser)
	if err != nil {
		y.logger.Error(err)
		return
	}

	viewAvg := 0
	viewTotal := 0
	ok := false
	begin := time.Now().AddDate(0, -3, 0)
	for _, v := range respUser.Contents.TwoColumnBrowseResultsRenderer.Tabs {
		if v.TabRenderer.Title != "Home" {
			continue
		}

		for _, v1 := range v.TabRenderer.Content.SectionListRenderer.Contents {
			for _, v2 := range v1.ItemSectionRenderer.Contents {
				i := 0
				for _, v3 := range v2.ShelfRenderer.Content.HorizontalListRenderer.Items {
					videoID := v3.GridVideoRenderer.VideoID
					if videoID == "" {
						continue
					}

					viewCountText := v3.GridVideoRenderer.ViewCountText.SimpleText
					viewCount := 0
					if viewCountText != "" {
						viewCountInt, e := strconv.Atoi(strings.Join(y.intRe.FindAllString(viewCountText, -1), ""))
						if e != nil {
							y.logger.Error(e, "viewCount", viewCountText)
							continue
						}
						viewCount = viewCountInt
					}

					t := time.Now().Unix()
					r := y.publishedTimeRe.FindAllString(v3.GridVideoRenderer.PublishedTimeText.SimpleText, -1)
					y.logger.Info("PublishedTime", v3.GridVideoRenderer.PublishedTimeText.SimpleText, r)
					if len(r) == 2 {
						i1, _ := strconv.Atoi(r[0])
						switch r[1] {
						case "year":
							t -= int64(i1 * 60 * 60 * 24 * 30 * 365)
							y.logger.Info("PublishedTime", r[1], v3.GridVideoRenderer.PublishedTimeText.SimpleText, t)
						case "month":
							t -= int64(i1 * 60 * 60 * 24 * 30)
							y.logger.Info("PublishedTime", r[1], v3.GridVideoRenderer.PublishedTimeText.SimpleText, t)
						case "week":
							t -= int64(i1 * 60 * 60 * 24 * 7)
							y.logger.Info("PublishedTime", r[1], v3.GridVideoRenderer.PublishedTimeText.SimpleText, t)
						case "day":
							t -= int64(i1 * 60 * 60 * 24)
							y.logger.Info("PublishedTime", r[1], v3.GridVideoRenderer.PublishedTimeText.SimpleText, t)
						case "hour":
							t -= int64(i1 * 60 * 60)
							y.logger.Info("PublishedTime", r[1], v3.GridVideoRenderer.PublishedTimeText.SimpleText, t)
						case "minute":
							t -= int64(i1 * 60)
							y.logger.Info("PublishedTime", r[1], v3.GridVideoRenderer.PublishedTimeText.SimpleText, t)
						case "second":
							t -= int64(i1)
							y.logger.Info("PublishedTime", r[1], v3.GridVideoRenderer.PublishedTimeText.SimpleText, t)
						default:
							y.logger.Error("PublishedTime", v3.GridVideoRenderer.PublishedTimeText.SimpleText)
						}
					}
					if time.Unix(t, 0).After(begin) {
						ok = true
					}

					i++
					viewTotal += viewCount
					viewAvg = viewTotal / i
					if i > 10 {
						break
					}
				}
			}
		}
	}

	if !ok {
		y.logger.Info("out date")
		return
	}

	subscriber := respUser.Header.C4TabbedHeaderRenderer.SubscriberCountText.SimpleText
	index := strings.Index(subscriber, " ")
	followers := 0
	if index < 1 {
		y.logger.Error("subscriber", subscriber)
	} else {
		followersText := subscriber[0:index]
		followers64, e := strconv.ParseFloat(strings.Join(y.floatRe.FindAllString(followersText, -1), ""), 64)
		if e != nil {
			y.logger.Error(e, "followers64", subscriber)
		}
		if strings.HasSuffix(followersText, "T") {
			followers = int(followers64 * 1000 * 1000 * 1000 * 1000)
		} else if strings.HasSuffix(followersText, "G") {
			followers = int(followers64 * 1000 * 1000 * 1000)
		} else if strings.HasSuffix(followersText, "M") {
			followers = int(followers64 * 1000 * 1000)
		} else if strings.HasSuffix(followersText, "K") {
			followers = int(followers64 * 1000)
		} else {
			followers = int(followers64)
		}
	}

	description := strings.TrimSpace(respUser.Metadata.ChannelMetadataRenderer.Description)
	email := ""
	r := y.emailRe.FindAllString(description, -1)
	if len(r) > 0 {
		email = r[0]
	}

	link := ""
	urls := y.urlRe.FindAllString(description, -1)
	if len(urls) > 0 {
		link = urls[0]
	}

	if viewAvg > 1000 && viewAvg < 100000 {
		data := Data{
			Id:          meta.Id,
			UserName:    meta.UserName,
			Description: description,
			Link:        link,
			Email:       email,
			Followers:   followers,
			ViewAvg10:   viewAvg,
			Keyword:     meta.KeyWord,
		}
		y.logger.Info(utils.JsonStr(data))
		err = y.save(context.Background(), &data)
		if err != nil {
			y.logger.Error(err)
			return
		}
	}

	return
}

func (y *YoutubeSpider) save(ctx context.Context, data *Data) (err error) {
	if err != nil {
		ctx = context.Background()
	}

	bs, err := bson.Marshal(data)
	if err != nil {
		y.logger.Error(err)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, y.timeout)
	defer cancel()

	res, err := y.collectionYoutubeUser.InsertOne(ctx, bs)
	if err != nil {
		y.logger.Error(err)
		return
	}
	y.logger.Info("insert success", res.InsertedID)

	return
}

func NewYoutubeSpider(config *config.Config, logger *logger.Logger, mongoDb *mongo.Database) (youtubeSpider *YoutubeSpider, err error) {
	collectionYoutubeUser := config.Mongo.CollectionYoutubeUser
	if collectionYoutubeUser == "" {
		err = errors.New("collectionYoutubeUser is empty")
		logger.Error(err)
		return
	}

	youtubeSpider = &YoutubeSpider{
		proxy:                 "http://127.0.0.1:33210",
		timeout:               time.Second * 30,
		collectionYoutubeUser: mongoDb.Collection(collectionYoutubeUser),
		logger:                logger,
		urlSearch:             "https://www.youtube.com/results?search_query=%s",
		urlSearchApi:          "https://www.youtube.com/youtubei/v1/search?key=%s",
		urlUserApi:            "https://www.youtube.com/youtubei/v1/browse?key=%s",

		apiKey:          "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8",
		searchJsonRe:    regexp.MustCompile(`ytInitialData = (.+);</script>`),
		apiKeyRe:        regexp.MustCompile(`"INNERTUBE_API_KEY":"([^"]+)`),
		emailRe:         regexp.MustCompile(`(\w+[-+.]*\w+@\w+[-.]*\w+\.\w+[-.]*\w+)`),
		urlRe:           regexp.MustCompile(`(?i)\b((?:https?://|www\d{0,3}[.]|[a-z0-9.-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s\` + "`" + `!()\[\]{};:'".,<>?«»“”‘’]))`),
		floatRe:         regexp.MustCompile(`[\d.]`),
		intRe:           regexp.MustCompile(`\d`),
		publishedTimeRe: regexp.MustCompile(`(\d+)\s*(year|month|week|day|hour|minute|second)`),
	}

	err = youtubeSpider.getClient()
	if err != nil {
		logger.Error(err)
		return
	}

	return
}
