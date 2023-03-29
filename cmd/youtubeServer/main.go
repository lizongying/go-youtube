package main

import (
	"context"
	"fmt"
	"github.com/lizongying/go-youtube/internal/config"
	"github.com/lizongying/go-youtube/internal/logger"
	"github.com/lizongying/go-youtube/internal/mongodb"
	"github.com/lizongying/go-youtube/internal/youtubeServer"
	"go.uber.org/fx"
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

func search(logger *logger.Logger, youtube *youtubeServer.YoutubeServer) (err error) {
	searchType := youtubeServer.SearchChannel
	pageMax := 1
	for _, v := range []string{
		"youtube",
	} {
		items, e := youtube.Search(context.Background(), []string{}, v, searchType, pageMax)
		if e != nil {
			logger.Error(e)
			continue
		}
		for _, item := range items {
			js, e := item.MarshalJSON()
			if e != nil {
				logger.Error(e)
				continue
			}
			logger.Info(string(js))

			switch item.Id.Kind {
			case "youtube#video":
			case "youtube#channel":
				_, e = youtube.Channels(context.Background(), []string{"statistics", "contentOwnerDetails"}, item.Id.ChannelId)
				if e != nil {
					logger.Error(e)
					continue
				}
			case "youtube#playlist":
			default:
			}
			break
		}
	}
	return
}

func main() {
	fx.New(
		fx.Provide(
			config.NewConfig,
			mongodb.NewMongoDb,
			logger.NewLogger,
			youtubeServer.NewYoutubeServer,
		),
		fx.Invoke(func(logger *logger.Logger, youtube *youtubeServer.YoutubeServer, shutdowner fx.Shutdowner) (err error) {

			err = search(logger, youtube)

			err = shutdowner.Shutdown()
			if err != nil {
				logger.Error(err)
				return
			}

			return
		}),
	).Run()
}
