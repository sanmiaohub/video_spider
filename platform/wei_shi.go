package platform

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type WeiShiPlatform struct {
	Record Record
}

type WeiShiJson struct {
	Ret  int    `json:"ret"`
	Msg  string `json:"msg"`
	Data struct {
		Feeds []struct {
			VideoUrl   string `json:"video_url"`
			FeedDesc   string `json:"feed_desc"`
			VideoCover struct {
				StaticCover struct {
					Url string `json:"url"`
				} `json:"static_cover"`
			} `json:"video_cover"`
		} `json:"feeds"`
	} `json:"data"`
}

func (ws WeiShiPlatform) ParseOut() (record Record, err error) {

	// 根据视频ID获取视频内容
	matches := regexp.MustCompile(`id=([^&]+)`).FindStringSubmatch(ws.Record.Link)
	if len(matches) < 2 {
		err = errors.New("无法解析该分享链接")
		return
	}
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 阻止重定向
		},
	}
	req, err := client.Get(fmt.Sprintf("https://h5.weishi.qq.com/webapp/json/weishi/WSH5GetPlayPage?feedid=%s", matches[1]))
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return
	}

	// 数据结构绑定
	var js json.RawMessage
	isJson := json.Unmarshal(body, &js) == nil
	if !isJson {
		err = errors.New("视频资源数据不正确")
		return
	}
	videoJson := WeiShiJson{}
	err = json.Unmarshal(body, &videoJson)
	if err != nil {
		return
	}
	feeds := videoJson.Data.Feeds
	if len(feeds) == 0 {
		err = errors.New("视频数据获取失败")
		return
	}

	ws.Record.Type = 1
	ws.Record.Title = feeds[0].FeedDesc
	ws.Record.Cover = feeds[0].VideoCover.StaticCover.Url
	ws.Record.Video = feeds[0].VideoUrl
	ws.Record.ResourcePath = feeds[0].VideoUrl

	return ws.Record, nil
}
