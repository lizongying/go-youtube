package youtubeServer

import (
	"errors"
	"github.com/lizongying/go-youtube/internal/config"
	"github.com/lizongying/go-youtube/internal/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YoutubeServer struct {
	key     string
	mongoDb *mongo.Database
	logger  *logger.Logger
	service *youtube.Service
}

type SearchType string

const SearchChannel SearchType = "channel"
const SearchPlaylist SearchType = "playlist"
const SearchVideo SearchType = "video"

func (y *YoutubeServer) getServiceWithToken(ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	y.service, err = youtube.NewService(ctx, option.WithAPIKey(y.key))
	if err != nil {
		y.logger.Error(err)
		return
	}

	return
}

func (y *YoutubeServer) Channels(ctx context.Context, part []string, id string) (channels []*youtube.Channel, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	call := y.service.Channels.List(part)
	call = call.Id(id)
	err = call.Pages(ctx, func(response *youtube.ChannelListResponse) (err error) {
		channels = response.Items
		return
	})
	if err != nil {
		y.logger.Error(err)
		return
	}

	return
}

func (y *YoutubeServer) Search(ctx context.Context, part []string, keyword string, searchType SearchType, pageMax int) (items []*youtube.SearchResult, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	call := y.service.Search.List(part)
	call = call.Q(keyword)
	if searchType != "" {
		call = call.Type(string(searchType))
	}
	page := 1
	err = call.Pages(ctx, func(response *youtube.SearchListResponse) (err error) {
		items = response.Items
		page++
		if page > pageMax {
			y.logger.Info("max page")
			return
		}
		return
	})
	if err != nil {
		y.logger.Error(err)
		return
	}

	return
}

func NewYoutubeServer(config *config.Config, logger *logger.Logger, mongoDb *mongo.Database) (youtubeServer *YoutubeServer, err error) {
	key := config.Youtube.Key
	if key == "" {
		err = errors.New("key is empty")
		logger.Error(err)
		return
	}

	youtubeServer = &YoutubeServer{
		key:     key,
		mongoDb: mongoDb,
		logger:  logger,
	}

	err = youtubeServer.getServiceWithToken(nil)
	if err != nil {
		logger.Error(err)
		return
	}

	return
}
