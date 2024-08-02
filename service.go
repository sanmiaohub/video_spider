package main

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"regexp"
	"strings"
	"video_spider/platform"
)

func hello(c echo.Context) error {
	return success(c, "Hello, World!")
}

func analysis(c echo.Context) error {
	// 解析链接
	linkText := c.FormValue("share_link")
	if linkText == "" {
		return errors.New("分享链接不能为空")
	}

	link := regexp.MustCompile(`https?://\S+`).FindString(linkText)

	// 解析链接匹配解析器, 绑定结果到Model
	var err error
	var record platform.Record
	switch {
	case strings.Contains(link, "v.douyin.com"):
		record, err = platform.DouYinPlatform{
			Record: platform.Record{Link: link},
		}.ParseOut()
	case strings.Contains(link, "h5.pipix.com"):
		record, err = platform.PiPiXiaPlatform{
			Record: platform.Record{Link: link},
		}.ParseOut()
	case strings.Contains(link, "isee.weishi.qq.com"):
		record, err = platform.WeiShiPlatform{
			Record: platform.Record{Link: link},
		}.ParseOut()
	case strings.Contains(link, "xhslink.com"):
		record, err = platform.RedBookPlatform{
			Record: platform.Record{Link: link},
		}.ParseOut()
	case strings.Contains(link, "www.bilibili.com"):
		record, err = platform.BiliBiliPlatform{
			Record: platform.Record{Link: link},
		}.ParseOut()
	default:
		err = errors.New("暂不支持该平台资源解析")
	}
	if err != nil {
		return err
	}

	return success(c, map[string]interface{}{
		"type":          record.Type,
		"title":         record.Title,
		"cover":         record.Cover,
		"video":         record.Video,
		"resource_path": record.ResourcePath,
	})
}

// Success 响应封装
func success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":    data,
		"message": "success",
	})
}
