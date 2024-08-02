package platform

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type BiliBiliPlatform struct {
	Record Record
}

type BiLiBiLiJson struct {
	Data struct {
		Dash struct {
			Video []VideoStream `json:"video"`
			Audio []AudioStream `json:"audio"`
		} `json:"dash"`
	} `json:"data"`
	VideoData struct {
		Pic   string `json:"pic"`
		Title string `json:"title"`
	} `json:"videoData"`
}

// VideoStream struct to parse video stream data
type VideoStream struct {
	ID           int      `json:"id"`
	BaseURL      string   `json:"baseUrl"`
	BackupURL    []string `json:"backupUrl"`
	Bandwidth    int      `json:"bandwidth"`
	MimeType     string   `json:"mimeType"`
	Codecs       string   `json:"codecs"`
	Width        int      `json:"width"`
	Height       int      `json:"height"`
	FrameRate    string   `json:"frameRate"`
	SAR          string   `json:"sar"`
	StartWithSAP int      `json:"startWithSap"`
	SegmentBase  struct {
		Initialization string `json:"Initialization"`
		IndexRange     string `json:"indexRange"`
	} `json:"SegmentBase"`
}

// AudioStream struct to parse audio stream data
type AudioStream struct {
	ID          int      `json:"id"`
	BaseURL     string   `json:"baseUrl"`
	BackupURL   []string `json:"backupUrl"`
	Bandwidth   int      `json:"bandwidth"`
	MimeType    string   `json:"mimeType"`
	Codecs      string   `json:"codecs"`
	SegmentBase struct {
		Initialization string `json:"Initialization"`
		IndexRange     string `json:"indexRange"`
	} `json:"SegmentBase"`
}

func (bl BiliBiliPlatform) ParseOut() (record Record, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", bl.Record.Link, nil)
	if err != nil {
		return Record{}, err
	}

	setRequestHeaders(req)

	htmlRes, err := client.Do(req)
	if err != nil {
		return Record{}, err
	}
	defer htmlRes.Body.Close()

	reader, err := getDecompressedReader(htmlRes)
	if err != nil {
		return Record{}, err
	}

	videoJson, videoDataJson, err := parseHTML(reader)
	if err != nil {
		return Record{}, err
	}

	dirPath := createDirPath()
	bestVideoStream := findBestQualityStream(videoJson.Data.Dash.Video)
	bestAudioStream := findBestQualityAudioStream(videoJson.Data.Dash.Audio)

	videoInitPath := dirPath + "/" + generateRandomString(6) + "best_video_init.m4s"
	audioInitPath := dirPath + "/" + generateRandomString(6) + "best_audio_init.m4s"
	finalOutputPath := dirPath + "/" + generateRandomString(6) + "final_output.mp4"

	if err = downloadStream(bestVideoStream.BaseURL, videoInitPath); err != nil {
		return Record{}, err
	}
	if err = downloadStream(bestAudioStream.BaseURL, audioInitPath); err != nil {
		return Record{}, err
	}
	if err = mergeFiles(videoInitPath, audioInitPath, finalOutputPath); err != nil {
		return Record{}, err
	}

	bl.Record.Type = 1
	bl.Record.Title = videoDataJson.VideoData.Title
	bl.Record.Cover = videoDataJson.VideoData.Pic
	bl.Record.Video = "https://api.analysis.sancat.cn/" + finalOutputPath
	bl.Record.ResourcePath = "https://api.analysis.sancat.cn/" + finalOutputPath

	return bl.Record, nil
}

func setRequestHeaders(req *http.Request) {
	headers := map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"Accept-Encoding":           "gzip, deflate, br, zstd",
		"Accept-Language":           "zh-CN,zh;q=0.9",
		"Referer":                   "https://www.bilibili.com/",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "same-origin",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36",
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	// Adding cookies
	cookies := map[string]string{
		"CURRENT_QUALITY":   "112",
		"CURRENT_FNVAL":     "4048",
		"DedeUserID":        "456026034",
		"DedeUserID__ckMd5": "96048149e7029403",
		"SESSDATA":          "6cdf984c%2C1738027742%2C2d49f%2A82CjAy6TnYTCy2CnPJUohRXtPb5CKMHqaHf-j16QL3RA4xJMOBh0pZOr50DM0L-Y_4FZUSVk1JazE1bUR3LXIzWnpOM3h5cWVMUUxhZldsWDlCYU5QdlktUzJnaXRSOVBUXzZnaGwwOGNUckx5MWdvc3pKQ2Y2aDV3b3BIZ2kxZHg3a0p5dDloZy1BIIEC",
		"_uuid":             "E714F5BD-B163-2F104-AEB2-B6DE98191151085390infoc",
	}
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
}

func parseHTML(reader io.Reader) (BiLiBiLiJson, BiLiBiLiJson, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return BiLiBiLiJson{}, BiLiBiLiJson{}, err
	}

	var videoJson, videoDataJson BiLiBiLiJson
	var jsonData, baseJsonData string

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptText := s.Text()
		if strings.Contains(scriptText, "window.__playinfo__=") {
			jsonData = extractJson(scriptText, "}")
		}
		if strings.Contains(scriptText, "window.__INITIAL_STATE__=") {
			baseJsonData = extractJson(scriptText, "};")
		}
	})

	if err = json.Unmarshal([]byte(jsonData), &videoJson); err != nil {
		return BiLiBiLiJson{}, BiLiBiLiJson{}, errors.New("failed to unmarshal video JSON data")
	}
	if err = json.Unmarshal([]byte(baseJsonData), &videoDataJson); err != nil {
		return BiLiBiLiJson{}, BiLiBiLiJson{}, errors.New("failed to unmarshal video data JSON")
	}

	return videoJson, videoDataJson, nil
}

func extractJson(scriptText string, endMark string) string {
	start := strings.Index(scriptText, "{")
	end := strings.LastIndex(scriptText, endMark) + 1
	return scriptText[start:end]
}

func createDirPath() string {
	dirPath := "video/" + time.Now().Format("20060102150405")
	_ = os.MkdirAll(dirPath, os.ModePerm)
	return dirPath
}

func mergeFiles(videoPath, audioPath, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-i", audioPath, "-c:v", "copy", "-c:a", "copy", outputPath)
	return cmd.Run()
}

func downloadStream(url string, outputPath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	setRequestHeaders(req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	return err
}

func findBestQualityStream(videoStreams []VideoStream) VideoStream {
	bestStream := videoStreams[0]
	for _, stream := range videoStreams {
		if stream.Bandwidth > bestStream.Bandwidth {
			bestStream = stream
		}
	}
	return bestStream
}

func findBestQualityAudioStream(audioStreams []AudioStream) AudioStream {
	bestStream := audioStreams[0]
	for _, stream := range audioStreams {
		if stream.Bandwidth > bestStream.Bandwidth {
			bestStream = stream
		}
	}
	return bestStream
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	return string(result)
}
