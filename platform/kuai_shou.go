package platform

import (
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

type QuickShouPlatform struct {
	Record Record
}

func (ks QuickShouPlatform) ParseOut() (record Record, err error) {

	// 重定向获取视频路径取编号
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 阻止重定向
		},
	}
	redirectRes, err := client.Head(ks.Record.Link)
	if err != nil {
		return
	}
	location := redirectRes.Header.Get("Location")
	if location == "" {
		err = errors.New("未获取到重定向地址")
		return
	}

	headers := map[string]string{
		"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-encoding":           "gzip, deflate, br, zstd",
		"accept-language":           "zh-CN,zh;q=0.9",
		"connection":                "keep-alive",
		"cookie":                    "kpf=PC_WEB; clientid=3; did=web_cee20d019f1c72467815c26358da2ee7; kpn=KUAISHOU_VISION",
		"host":                      "www.kuaishou.com",
		"sec-ch-ua":                 `"Not)A;Brand";v="99", "Google Chrome";v="127", "Chromium";v="127"`,
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        `"Windows"`,
		"sec-fetch-dest":            "document",
		"sec-fetch-mode":            "navigate",
		"sec-fetch-site":            "none",
		"sec-fetch-user":            "?1",
		"upgrade-insecure-requests": "1",
		"user-agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36",
	}
	req, err := http.NewRequest(http.MethodGet, location, nil)
	if err != nil {
		return
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client = &http.Client{}
	htmlRes, err := client.Do(req)
	if err != nil {
		return
	}
	body, err := getDecompressedReader(htmlRes)
	if err != nil {
		return
	}

	// 加载html内容, 查找视频资源信息
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return
	}
	jsonData := ""
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptText := s.Text()
		if strings.Contains(scriptText, "window.__APOLLO_STATE__") {
			start := strings.Index(scriptText, "{")
			end := strings.LastIndex(scriptText, "};") + 1
			jsonData = scriptText[start:end]
		}
	})
	if jsonData == "" {
		err = errors.New("暂时不支持该资源解析")
	}
	var js json.RawMessage
	isJson := json.Unmarshal([]byte(jsonData), &js) == nil
	if !isJson {
		err = errors.New("视频资源数据不正确")
		return
	}

	videoJson := map[string]interface{}{}
	err = json.Unmarshal([]byte(jsonData), &videoJson)
	if err != nil {
		return
	}
	fullData, ok := videoJson["defaultClient"]
	if !ok {
		err = errors.New("数据结构错误")
		return
	}

	for key, value := range fullData.(map[string]interface{}) {
		if strings.Contains(key, "VisionVideoDetailPhoto") && !strings.Contains(key, "manifest") {
			ks.Record.Type = 1
			if title, nok := value.(map[string]interface{})["caption"]; nok {
				ks.Record.Title = title.(string)
			}
			if coverUrl, nok := value.(map[string]interface{})["coverUrl"]; nok {
				ks.Record.Cover = coverUrl.(string)
			}
			if photoUrl, nok := value.(map[string]interface{})["photoUrl"]; nok {
				ks.Record.Video = photoUrl.(string)
				ks.Record.ResourcePath = photoUrl.(string)
			}
		}
	}

	return ks.Record, nil
}
