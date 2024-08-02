# video_spider
短视频分享链接解析，短视频去水印：抖音，皮皮虾，微视，小红书，哔哩哔哩等，不断增加中。

# 快捷体验

目前这个服务已经部署到服务器，客户端使用的小程序，可以**扫码体验**。

![video_spider](https://resources.linghanghuiye.com/image/console/20240725/66a24f542f7da1721913172.webp)

# 项目依赖

如果需要使用哔哩哔哩解析下载，需要下载ffmpeg，目前测试过5.*版本和7.*版本，都可以正常合并生成视频。

ffmpeg的安装教程各linux的安装方法不一致，请自行百度安装。

# 部署说明

代码开发以golang1.22.5版本为主，构建也建议使用此版本或更高版本，以部署到linux服务器为例，执行下面的命令得到二进制文件，放到服务器直接运行启动。

建议使用宝塔面板golang部署管理。

```shell
go env -w GOOS=linux

go build -o analysis

./analysis

```

# 请求示例
```
curl --location --request POST 'http://127.0.0.1:8080/analysis' \
--header 'User-Agent: Apifox/1.0.0 (https://apifox.com)' \
--header 'Accept: */*' \
--header 'Host: 127.0.0.1:8080' \
--header 'Connection: keep-alive' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'share_link=32 【童趣ins春日清新显白可爱美甲 - 沃小妮妮 | 小红书 - 你的生活指南】 😆 mp9w4UUBKZ3eKDi 😆 http://xhslink.com/98JQgR'
```

# 响应示例

#### 响应字段说明

| 字段说明 | 说明 | 额外解释 |
|  ----  | ----  | ---- |
| type  | 1为视频 2为图集 |  |
| cover | 封面图 |  |
| title | 视频或者图文的文案 |  |
| video | 视频地址播放 type为1有值 |  |
| resource_path | 视频下载地址、图文下载地址 | 视频为字符串、图片为数组 |

#### 响应数据结构示例

```json
{
    "data": {
        "cover": "http://sns-webpic-qc.xhscdn.com/202408021008/310772040177f73d98119d109e690768/1000g00822835degfq0004a471n6uaht53bs3hl0!nd_dft_wlteh_jpg_3",
        "resource_path": [
            "http://sns-webpic-qc.xhscdn.com/202408021008/310772040177f73d98119d109e690768/1000g00822835degfq0004a471n6uaht53bs3hl0!nd_dft_wlteh_jpg_3",
            "http://sns-webpic-qc.xhscdn.com/202408021008/a7a3565e88cd4c012204f7a28f76ea4b/1000g00822835degfq00g4a471n6uaht5pc9omf0!nd_dft_wlteh_jpg_3",
            "http://sns-webpic-qc.xhscdn.com/202408021008/f90e528088b67bdddb7603b201902385/1000g00822835degfq0104a471n6uaht5gnepj70!nd_dft_wlteh_jpg_3",
            "http://sns-webpic-qc.xhscdn.com/202408021008/e6d8b9fe18e90e69b318c6d81ed3e306/1000g00822835degfq01g4a471n6uaht5ro4chu0!nd_dft_wlteh_jpg_3"
        ],
        "title": "童趣ins春日清新显白可爱美甲",
        "type": 2,
        "video": ""
    },
    "message": "success"
}
```

## FAQ

**关于配置项**

目前代码中需要配置的只有一个端口号，现在写死的8080，根据需要自己修改即可。如果要更换成配置文件形式，自行开发，去使用yaml配置文件（比较简单就自己开发吧）。

**关于解析速度**

目前各平台解析的速度在600毫秒左右，bilibili特殊，需要下载m4s源文件，进行ffmpeg合并生成。所有比较慢。

**关于有些视频平台解析失败**

有些平台需要cookie，请手动更新cookie，如果还是解析失败，请提交issues

示例：哔哩哔哩，登录哔哩哔哩官网，在F12打开检查，打开应用，找到COOKIE，找到bilibili网址，找到代码中需要的cookie，替换对应key的值即可，哔哩哔哩下载的清晰度和登录人的会员等级有关。

**关于客户端代码**

目前暂不提供前端代码（代码太丑了），后面重构后会进行开源公布。

# 免责声明
本仓库只为学习研究，如涉及侵犯个人或者团体利益，请与我取得联系，我将主动删除一切相关资料，谢谢！