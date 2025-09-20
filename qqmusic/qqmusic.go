package qqmusic

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"log/slog"

	"github.com/meloshub/meloshub/adapter"
	"github.com/meloshub/meloshub/model"
)

var (
	ErrSearchSong      = errors.New("[QQMusicAdapter] Failed to search song")
	ErrGetPlayURL      = errors.New("[QQMusicAdapter] Failed to get play url")
	ErrEmptyPlayURL    = errors.New("[QQMusicAdapter] play url is empty")
	ErrGetAlbumDetail  = errors.New("[QQMusicAdapter] Failed to album detail")
	ErrParseTimeString = errors.New("[QQMusicAdapter] Failed to parse time string")
	ErrStatusCode      = errors.New("[QQMusicAdapter] Http status error")
	ErrUnmarshaling    = errors.New("[QQMusicAdapter] Failed to unmarshaling body bytes")
	ErrBase64Decode    = errors.New("[QQMusicAdapter] Failed to decode base64")
)

// 搜索结果响应体结构体
type searchResponseBody struct {
	Code int64 `json:"code"`
	Data struct {
		Keyword string `json:"keyword"`
		Song    struct {
			List []songItem `json:"list"`
		} `json:"song"`
	} `json:"data"`
}

type songItem struct {
	Albumid   int64        `json:"albumid"`
	Albummid  string       `json:"albummid"`
	Albumname string       `json:"albumname"`
	Docid     string       `json:"docid"`
	Pubtime   int64        `json:"pubtime"`
	Size128   int64        `json:"size128"`
	Size320   int64        `json:"size320"`
	Sizeape   int64        `json:"sizeape"`
	Sizeflac  int64        `json:"sizeflac"`
	Sizeogg   int64        `json:"sizeogg"`
	Songid    int64        `json:"songid"`
	Songmid   string       `json:"songmid"`
	Songname  string       `json:"songname"`
	Stream    int64        `json:"stream"`
	Singer    []singerItem `json:"singer"`
	Pay       struct {
		Payalbum      int64 `json:"payalbum"`
		Payalbumprice int64 `json:"payalbumprice"`
		Paydownload   int64 `json:"paydownload"`
		Payinfo       int64 `json:"payinfo"`
		Payplay       int64 `json:"payplay"`
		Paytrackmouth int64 `json:"paytrackmouth"`
		Paytrackprice int64 `json:"paytrackprice"`
	} `json:"pay"`
}

type singerItem struct {
	Id   int64  `json:"id"`
	Mid  string `json:"mid"`
	Name string `json:"name"`
}

type QQMusicAdapter struct {
	adapter.Base
	qqNumber int //登录账户的qq号，未登录时默认为0
}

// 适配器将在导入时被注册
func init() {
	if err := adapter.Register(New()); err != nil {
		panic(fmt.Errorf("failed to register adapter: %w", err))
	}
}

func New() *QQMusicAdapter {
	a := &QQMusicAdapter{
		qqNumber: 0,
	}
	metadata := adapter.Metadata{
		Id:          "qqmusic",
		Title:       "QQ音乐适配器",
		Type:        adapter.TypeOfficial,
		Version:     "0.1.1",
		Author:      "Kaguya233qwq",
		Description: "qq音乐适配器",
	}
	a.Init(metadata)
	return a
}

func (a *QQMusicAdapter) Search(keyword string, options adapter.SearchOptions) ([]model.Song, error) {
	a.Session.Headers["Referer"] = "http://m.y.qq.com"
	api := "http://c.y.qq.com/soso/fcgi-bin/search_for_qq_cp"
	params := map[string]string{
		"w":      keyword,
		"format": "json",
		"p":      fmt.Sprint(options.Page),
		"n":      fmt.Sprint(options.Limit),
	}
	resp, err := a.Session.Get(api, params)
	if err != nil {
		return []model.Song{}, fmt.Errorf("[QQMusicAdapter] %v", err)
	}
	if resp.StatusCode != 200 {
		return []model.Song{}, ErrStatusCode
	}

	var respJson searchResponseBody
	err = resp.JSON(&respJson)
	if err != nil {
		slog.Error(err.Error())
		return []model.Song{}, ErrUnmarshaling
	}

	if respJson.Code != 0 {
		return []model.Song{}, ErrSearchSong
	}

	var results []model.Song
	for _, song := range respJson.Data.Song.List {

		var singerList []model.Singer
		for _, singer := range song.Singer {
			singerList = append(singerList, model.Singer{
				ID:   singer.Mid,
				Name: singer.Name,
			})
		}

		results = append(results, model.Song{
			ID:        song.Songmid,
			Source:    a.Id(),
			Title:     song.Songname,
			Singers:   singerList,
			AlbumId:   song.Albummid,
			AlbumName: song.Albumname,
			Playable:  song.Pay.Payplay == 0, //播放无需付费即可播放
		})
	}

	return results, nil
}

type requestResult struct {
	Code int `json:"code"`
	Req0 struct {
		Code int `json:"code"`
		Data struct {
			MidURLInfo []midURLInfo `json:"midurlinfo"`
			Expiration int          `json:"expiration"`
		} `json:"data"`
	} `json:"req_0"`
}

type midURLInfo struct {
	Songmid  string `json:"songmid"`
	Filename string `json:"filename"`
	Purl     string `json:"purl"`
	Vkey     string `json:"vkey"`
}

func (a *QQMusicAdapter) PlayURL(id string) (string, error) {

	api := "https://u.y.qq.com/cgi-bin/musicu.fcg"
	guidNum := rand.Intn(9_000_000_000) + 1_000_000_000
	guid := strconv.FormatInt(int64(guidNum), 10)
	pipeline := map[string]any{
		"req_0": map[string]any{
			"module": "vkey.GetVkeyServer",
			"method": "CgiGetVkey",
			"param": map[string]any{
				"filename":  []string{fmt.Sprintf("M500%s%s.mp3", id, id)},
				"guid":      guid,
				"songmid":   []string{id},
				"songtype":  []int{0},
				"uin":       fmt.Sprint(a.qqNumber),
				"loginflag": 1,
				"platform":  "20",
			},
		},
		"comm": map[string]any{
			"uin": fmt.Sprint(a.qqNumber), "format": "json", "ct": 24, "cv": 0,
		},
		"loginUin": fmt.Sprint(a.qqNumber),
	}
	jsonData, err := json.Marshal(pipeline)
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	params := map[string]string{
		"format": "json",
		"data":   string(jsonData),
	}

	resp, err := a.Session.Get(api, params)
	if err != nil {
		slog.Error(fmt.Sprintf("[QQMusicAdapter] %v", err))
	}

	if resp.StatusCode != 200 {
		return "", ErrStatusCode
	}

	var respJson requestResult
	err = resp.JSON(&respJson)
	if err != nil {
		return "", ErrUnmarshaling
	}

	if respJson.Code != 0 || respJson.Req0.Code != 0 {
		slog.Error(fmt.Sprintf("PlayURL error response: %+v", respJson))
		return "", ErrGetPlayURL
	}

	if respJson.Req0.Data.MidURLInfo[0].Purl == "" {
		return "", ErrEmptyPlayURL
	}
	streamApi := "http://ws.stream.qqmusic.qq.com/"
	return streamApi + respJson.Req0.Data.MidURLInfo[0].Purl, nil
}

func (a *QQMusicAdapter) Lyrics(id string) (string, error) {
	api := "https://c.y.qq.com/lyric/fcgi-bin/fcg_query_lyric_new.fcg"
	params := map[string]string{
		"songmid":     id,
		"loginUin":    fmt.Sprint(a.qqNumber),
		"hostUin":     "0",
		"format":      "json",
		"inCharset":   "utf8",
		"outCharset":  "utf-8",
		"notice":      "0",
		"platform":    "yqq.json",
		"needNewCode": "0",
	}

	resp, err := a.Session.Get(api, params)
	if err != nil {
		slog.Error(fmt.Sprintf("[QQMusicAdapter] %v", err))
	}

	if resp.StatusCode != 200 {
		return "", ErrStatusCode
	}

	type Result struct {
		Code  int    `json:"code"`
		Lyric string `json:"lyric"`
	}

	var respJson Result
	err = resp.JSON(&respJson)
	if err != nil {
		return "", ErrUnmarshaling
	}

	lyric, err := base64.StdEncoding.DecodeString(respJson.Lyric)
	if err != nil {
		return "", ErrBase64Decode
	}
	return string(lyric), nil
}

type picUrlItem struct {
	PicUrl string `json:"picurl"`
}

type singer struct {
	Singerid   string `json:"singerid"`
	Singermid  string `json:"singermid"`
	Singername string `json:"singername"`
}

type albumDetailResult struct {
	Code int `json:"code"`
	Data struct {
		AlbumMid    string       `json:"album_mid"`
		AlbumName   string       `json:"album_name"`
		Desc        string       `json:"desc"`
		HeadpicList []picUrlItem `json:"headpiclist"`
		PublicTime  string       `json:"publictime"`
		SingerInfo  []singer     `json:"singerinfo"`
		SongList    []songItem   `json:"songlist"`
	} `json:"data"`
}

func (a *QQMusicAdapter) AlbumDetail(id string) (model.Album, error) {
	api := "https://c6.y.qq.com/v8/fcg-bin/musicmall.fcg"
	params := map[string]string{
		"cv":                "4747474",
		"ct":                "24",
		"format":            "json",
		"inCharset":         "utf-8",
		"outCharset":        "utf-8",
		"notice":            "0",
		"platform":          "yqq.json",
		"needNewCode":       "1",
		"uin":               fmt.Sprint(a.qqNumber),
		"g_tk_new_20200303": "1237177036",
		"g_tk":              "2059730570",
		"cmd":               "get_album_buy_page",
		"albummid":          id,
		"albumid":           "0",
	}
	timestamp := time.Now().UnixMilli()
	params["_"] = strconv.FormatInt(timestamp, 10)

	resp, err := a.Session.Get(api, params)
	if err != nil {
		slog.Error(fmt.Sprintf("[QQMusicAdapter] %v", err))
	}

	if resp.StatusCode != 200 {
		return model.Album{}, ErrStatusCode
	}

	var respJson albumDetailResult
	err = resp.JSON(&respJson)
	if err != nil {
		return model.Album{}, ErrUnmarshaling
	}

	if respJson.Code != 0 {
		slog.Error(fmt.Sprintf("PlayURL error response: %+v", respJson))
		return model.Album{}, ErrGetAlbumDetail
	}

	parsedTime, err := time.Parse("2006-01-02", respJson.Data.PublicTime)
	if err != nil {
		return model.Album{}, ErrParseTimeString
	}
	publicTimeStamp := parsedTime.Unix()

	var songList []model.Song
	for _, song := range respJson.Data.SongList {

		var singerList []model.Singer
		for _, singer := range song.Singer {
			singerList = append(singerList, model.Singer{
				ID:   singer.Mid,
				Name: singer.Name,
			})
		}
		songList = append(songList, model.Song{
			ID:        song.Songmid,
			Source:    a.Id(),
			Title:     song.Songname,
			Singers:   singerList,
			AlbumId:   song.Albummid,
			AlbumName: song.Albumname,
			Playable:  song.Pay.Payplay == 0,
		})
	}

	var singerList []model.Singer
	for _, singer := range respJson.Data.SingerInfo {
		singerList = append(singerList, model.Singer{
			ID:   singer.Singermid,
			Name: singer.Singername,
		})
	}

	return model.Album{
		ID:              respJson.Data.AlbumMid,
		Name:            respJson.Data.AlbumName,
		Description:     respJson.Data.Desc,
		PublicTimestamp: publicTimeStamp,
		CoverURL:        respJson.Data.HeadpicList[0].PicUrl,
		SongList:        songList,
		Singers:         singerList,
	}, nil
}
