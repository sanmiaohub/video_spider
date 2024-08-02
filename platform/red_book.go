package platform

import (
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

type RedBookPlatform struct {
	Record Record
}

type RedBookVideoJson struct {
	Note struct {
		FirstNoteId   string `json:"firstNoteId"`
		NoteDetailMap map[string]struct {
			Note struct {
				Type  string `json:"type"`
				Title string `json:"title"`
				Video struct {
					Media struct {
						Stream struct {
							H265 []struct {
								MasterUrl string `json:"masterUrl"`
							} `json:"h265"`
							H264 []struct {
								MasterUrl string `json:"masterUrl"`
							} `json:"h264"`
						} `json:"stream"`
					} `json:"media"`
				} `json:"video"`
				ImageList []struct {
					UrlDefault string `json:"urlDefault"`
				} `json:"imageList"`
			} `json:"note"`
		} `json:"noteDetailMap"`
	} `json:"note"`
}

func (rb RedBookPlatform) ParseOut() (record Record, err error) {

	// 重定向获取视频路径取编号
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 阻止重定向
		},
	}
	redirectRes, err := client.Head(rb.Record.Link)
	if err != nil {
		return
	}
	// 获取视频资源HTML数据
	detailURL := redirectRes.Header.Get("Location")
	detailResp, err := http.Get(detailURL)
	if err != nil {
		return
	}
	// 加载html内容, 查找视频资源信息
	doc, err := goquery.NewDocumentFromReader(detailResp.Body)
	if err != nil {
		return
	}
	jsonData := ""
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptText := s.Text()
		if strings.Contains(scriptText, "window.__INITIAL_STATE__") {
			start := strings.Index(scriptText, "{")
			end := strings.LastIndex(scriptText, "}") + 1
			jsonData = scriptText[start:end]
			jsonData = strings.ReplaceAll(jsonData, "undefined", "\"\"")
		}
	})
	var js json.RawMessage
	isJson := json.Unmarshal([]byte(jsonData), &js) == nil
	if !isJson {
		err = errors.New("视频资源数据不正确")
		return
	}
	var rbVideo RedBookVideoJson
	if err = json.Unmarshal([]byte(jsonData), &rbVideo); err != nil {
		return
	}

	// 构建响应数据
	note, ok := rbVideo.Note.NoteDetailMap[rbVideo.Note.FirstNoteId]
	if !ok {
		err = errors.New("数据解析错误")
		return
	}

	// 图文
	if note.Note.Type == "normal" {
		var imageResource []string
		for _, v := range note.Note.ImageList {
			imageResource = append(imageResource, v.UrlDefault)
		}
		rb.Record.Type = 2
		rb.Record.Title = note.Note.Title
		rb.Record.Cover = note.Note.ImageList[0].UrlDefault
		rb.Record.ResourcePath = imageResource
	}

	// 视频
	if note.Note.Type == "video" {
		video := ""
		if len(note.Note.Video.Media.Stream.H264) > 0 {
			video = note.Note.Video.Media.Stream.H264[0].MasterUrl
		}
		if len(note.Note.Video.Media.Stream.H265) > 0 {
			video = note.Note.Video.Media.Stream.H265[0].MasterUrl
		}
		rb.Record.Type = 1
		rb.Record.Title = note.Note.Title
		rb.Record.Cover = note.Note.ImageList[0].UrlDefault
		rb.Record.Video = video
		rb.Record.ResourcePath = video
	}

	return rb.Record, nil
}
