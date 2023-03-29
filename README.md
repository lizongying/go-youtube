# go-youtube

通过youtubeApi或youtubeSpider获取youtube数据

## youtube

[youtube](https://developers.google.com/youtube/v3/quickstart/go)

## Server（api）

```shell
export https_proxy=http://127.0.0.1:33210 http_proxy=http://127.0.0.1:33210 all_proxy=socks5://127.0.0.1:33211 && go run cmd/youtubeServer/*.go -c example.yml
```

## Spider

```shell
export https_proxy=http://127.0.0.1:33210 http_proxy=http://127.0.0.1:33210 all_proxy=socks5://127.0.0.1:33211 && go run cmd/youtubeSpider/*.go -c example.yml
```