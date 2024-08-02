package platform

import (
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"strings"
)

type DouYinPlatform struct {
	Record Record
}

// VideoJson 响应数据
type VideoJson struct {
	LoaderData struct {
		VideoPage struct {
			VideoInfoRes struct {
				ItemList []struct {
					Desc   string `json:"desc"`
					Images []struct {
						UrlList []string `json:"url_list"`
					} `json:"images"`
					Video struct {
						Player struct {
							Uri string `json:"uri"`
						} `json:"play_addr"`
						Cover struct {
							UrlList []string `json:"url_list"`
						} `json:"cover"`
					} `json:"video"`
				} `json:"item_list"`
			} `json:"videoInfoRes"`
		} `json:"video_(id)/page"`
		NotePage struct {
			VideoInfoRes struct {
				ItemList []struct {
					Desc   string `json:"desc"`
					Images []struct {
						UrlList []string `json:"url_list"`
					} `json:"images"`
				} `json:"item_list"`
			} `json:"videoInfoRes"`
		} `json:"note_(id)/page"`
	} `json:"loaderData"`
}

func (dy DouYinPlatform) ParseOut() (record Record, err error) {

	// 构建请求获取视频编号
	client := &http.Client{}
	req, err := http.NewRequest("GET", dy.Record.Link, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Mobile Safari/537.36")
	htmlRes, err := client.Do(req)
	if err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {

		}
	}(htmlRes.Body)

	// 加载html内容, 查找视频资源信息
	doc, err := goquery.NewDocumentFromReader(htmlRes.Body)
	if err != nil {
		return
	}
	jsonData := ""
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptText := s.Text()
		if strings.Contains(scriptText, "window._ROUTER_DATA") {
			start := strings.Index(scriptText, "{")
			end := strings.LastIndex(scriptText, "}") + 1
			jsonData = scriptText[start:end]
		}
	})
	var js json.RawMessage
	isJson := json.Unmarshal([]byte(jsonData), &js) == nil
	if !isJson {
		err = errors.New("视频资源数据不正确")
		return
	}
	videoJson := VideoJson{}
	err = json.Unmarshal([]byte(jsonData), &videoJson)
	if err != nil {
		return

	}

	// 验证数据合法
	noteItemList := videoJson.LoaderData.NotePage.VideoInfoRes.ItemList
	videoItemList := videoJson.LoaderData.VideoPage.VideoInfoRes.ItemList
	if len(noteItemList) < 1 && len(videoItemList) < 1 {
		err = errors.New("解析数据异常,请向站点反馈分享链接排查")
		return
	}

	// 图文资源 -- 图文两种情况
	if len(videoItemList) > 0 && len(videoItemList[0].Images) > 0 {
		var imageResource []string
		for _, v := range videoItemList[0].Images {
			imageResource = append(imageResource, v.UrlList[0])
		}
		dy.Record.Type = 2
		dy.Record.Cover = videoItemList[0].Video.Cover.UrlList[0]
		dy.Record.Title = videoItemList[0].Desc
		dy.Record.ResourcePath = imageResource
	}
	if len(noteItemList) > 0 {
		var imageResource []string
		for _, v := range noteItemList[0].Images {
			imageResource = append(imageResource, v.UrlList[0])
		}
		dy.Record.Type = 2
		dy.Record.Title = noteItemList[0].Desc
		dy.Record.Cover = noteItemList[0].Images[0].UrlList[0]
		dy.Record.ResourcePath = imageResource
	}

	// 视频资源
	if len(videoItemList) > 0 && len(videoItemList[0].Images) == 0 {

		var redirectRes *http.Response
		var redirectUrl = "https://www.douyin.com/aweme/v1/play/?ratio=1080p&video_id=" + videoItemList[0].Video.Player.Uri

		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 阻止重定向
		}
		redirectRes, err = client.Head(redirectUrl)
		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {

			}
		}(redirectRes.Body)
		imageResource := redirectRes.Header.Get("Location")

		dy.Record.Type = 1
		dy.Record.Title = videoItemList[0].Desc
		dy.Record.Cover = videoItemList[0].Video.Cover.UrlList[0]
		dy.Record.Video = redirectUrl
		dy.Record.ResourcePath = imageResource
	}

	return dy.Record, nil
}
