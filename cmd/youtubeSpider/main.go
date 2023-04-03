package main

import (
	"fmt"
	"github.com/lizongying/go-youtube/internal/config"
	"github.com/lizongying/go-youtube/internal/logger"
	"github.com/lizongying/go-youtube/internal/mongodb"
	"github.com/lizongying/go-youtube/internal/youtubeSpider"
	"go.uber.org/fx"
	"golang.org/x/net/context"
	"log"
	"runtime"
)

var (
	buildBranch string
	buildCommit string
	buildTime   string
)

func init() {
	info := fmt.Sprintf("Branch: %s, Commit: %s, Time: %s, GOVersion: %s, OS: %s, ARCH: %s", buildBranch, buildCommit, buildTime, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	log.Println(info)
}

func search(youtube *youtubeSpider.YoutubeSpider) (err error) {
	for _, v := range []string{
		"makeupvideo", "makeuptutorial", "beautytips", "beauty", "fashion", "fashionstyle", "fashiondiaries", "fashiontrends", "springfashion", "outfits", "ootd", "outfitoftheday",
	} {
		_ = youtube.Search(context.Background(), youtubeSpider.MetaSearch{
			Keyword: v,
		})
	}
	return
}

func searchApi(youtube *youtubeSpider.YoutubeSpider) (err error) {
	err = youtube.SearchApi(context.Background(), youtubeSpider.MetaSearch{
		Keyword:       "youtube",
		Page:          1,
		MaxPage:       2,
		NextPageToken: "",
	})
	return
}

func userApi(youtube *youtubeSpider.YoutubeSpider) (err error) {
	err = youtube.UserApi(context.Background(), youtubeSpider.MetaUser{
		Key: "UCYJhto4Of0p8eKKxmB2un9g",
	})
	return
}

func videos(youtube *youtubeSpider.YoutubeSpider) (err error) {
	err = youtube.Videos(context.Background(), youtubeSpider.MetaUser{
		Id: "sierramarie",
	})
	return
}

func main() {
	fx.New(
		fx.Provide(
			config.NewConfig,
			mongodb.NewMongoDb,
			logger.NewLogger,
			youtubeSpider.NewYoutubeSpider,
		),
		fx.Invoke(func(youtube *youtubeSpider.YoutubeSpider, shutdowner fx.Shutdowner) (err error) {
			err = search(youtube)
			if err != nil {
				return
			}

			err = shutdowner.Shutdown()
			if err != nil {
				return
			}

			return
		}),
	).Run()
}
