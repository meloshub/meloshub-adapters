package tests

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/meloshub/meloshub/adapter"
	"github.com/meloshub/meloshub/logging"

	_ "github.com/meloshub/meloshub-adapters/qqmusic"
)

func TestQQMusic(t *testing.T) {
	logging.Init(logging.Config{
		Level:     "info",
		Format:    "consle",
		AddSource: true,
	})
	slog.Info("Meloshub is initializing..")
	slog.Info("Getting available adapters")
	adapters := adapter.GetAll()
	if len(adapters) == 0 {
		slog.Warn("Adapters list is empty")
	}
	for _, v := range adapters {
		println(v.Id())
	}
	adapterQQmusic, ok := adapter.Get("qqmusic")
	if !ok {
		slog.Error("adapter 'qqmusic' is not existed")
	}
	// 基本功能测试
	// 测试搜索
	songList, err := adapterQQmusic.Search("夜的第七章", adapter.SearchOptions{
		Page:  1,
		Limit: 10,
	})
	if err != nil {
		slog.Error(err.Error())
		return
	}

	jsonBytes, err := json.MarshalIndent(songList, "", "  ")
	if err != nil {
		fmt.Println("JSON marshaling failed:", err)
		return
	}
	fmt.Println(string(jsonBytes))
	// 用一个可播放的id来测试播放
	playUrl, err := adapterQQmusic.PlayURL("004Ng8xu20eirf")
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info(fmt.Sprint("Got play url: ", playUrl))
	// 测试获取歌词
	lyrics, err := adapterQQmusic.Lyrics(songList[0].ID)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info(fmt.Sprintln("Lyrics: \n", lyrics))

}
