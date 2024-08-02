package platform

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
)

type PiPiXiaPlatform struct {
	Record Record
}

// PPXVideoJson 响应数据
type PPXVideoJson struct {
	Data struct {
		Data struct {
			Item struct {
				Content string `json:"content"`
				Cover   struct {
					UrlList []struct {
						Url string `json:"url"`
					} `json:"url_list"`
				} `json:"cover"`
				Note struct {
					MultiImage []struct {
						UrlList []struct {
							Url string `json:"url"`
						} `json:"url_list"`
					} `json:"multi_image"`
				} `json:"note"`
				OriginVideoDownload struct {
					UrlList []struct {
						Url string `json:"url"`
					} `json:"url_list"`
				} `json:"origin_video_download"`
			} `json:"item"`
		} `json:"data"`
	} `json:"data"`
}

func (ppx PiPiXiaPlatform) ParseOut() (record Record, err error) {

	// 重定向获取视频路径取编号
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 阻止重定向
		},
	}
	redirectRes, err := client.Head(ppx.Record.Link)
	if err != nil {
		return
	}
	re := regexp.MustCompile("/item/(.*)\\?")
	matches := re.FindStringSubmatch(redirectRes.Header.Get("Location"))
	if len(matches) < 2 {
		err = errors.New("获取视频编号失败")
		return
	}

	// 获取视频资源HTML数据
	detailURL := "https://is.snssdk.com/bds/cell/detail/?cell_type=1&aid=1319&app_name=super&cell_id=" + matches[1]
	detailResp, err := http.Get(detailURL)
	body, err := io.ReadAll(detailResp.Body)
	if err != nil {
		return
	}
	var js json.RawMessage
	isJson := json.Unmarshal(body, &js) == nil
	if !isJson {
		err = errors.New("视频资源数据不正确")
		return
	}
	var ppxVideo PPXVideoJson
	if err = json.Unmarshal(body, &ppxVideo); err != nil {
		return
	}

	// 构建响应数据
	item := ppxVideo.Data.Data.Item
	note := ppxVideo.Data.Data.Item.Note
	video := ppxVideo.Data.Data.Item.OriginVideoDownload
	// 图文
	if len(note.MultiImage) > 0 {
		var imageResource []string
		for _, v := range note.MultiImage {
			imageResource = append(imageResource, v.UrlList[0].Url)
		}
		ppx.Record.Type = 2
		ppx.Record.Title = item.Content
		ppx.Record.Cover = item.Cover.UrlList[0].Url
		ppx.Record.ResourcePath = imageResource
	}

	// 视频
	if len(video.UrlList) > 0 {
		ppx.Record.Type = 1
		ppx.Record.Title = item.Content
		ppx.Record.Cover = item.Cover.UrlList[0].Url
		ppx.Record.Video = video.UrlList[0].Url
		ppx.Record.ResourcePath = video.UrlList[0].Url
	}

	return ppx.Record, nil
}
