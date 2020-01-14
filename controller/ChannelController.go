package controller

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"myfabric-tool/bindata"
	"myfabric-tool/service"
	"net/http"
)

func ChannelList(w http.ResponseWriter, r *http.Request) {
	//解析模板文件
	t, _ := template.ParseFiles("static/channelList.html")
	//执行模板
	t.Execute(w, r.FormValue("msg"))
}

func ChannelListJson(w http.ResponseWriter, r *http.Request) {
	// channelList := []*model.Channel{
	// 	&model.Channel{
	// 		Name: "channel11",
	// 	},
	// 	&model.Channel{
	// 		Name: "channel21",
	// 	},
	// }

	channelList, err := service.ListChannels()

	if err!=nil {
		log.Println(err)
	}

	/*result := model.ChannelList{
		Code:  "0",
		Msg:   "",
		Count: 2,
		Data:  channelList,
	}*/
	result := map[string]interface{}{
		"code":  "0",
		"msg":   "6",
		"count": len(channelList),
		"data":  channelList,
	}
	b, _ := json.Marshal(result)
	w.Write(b)
}

func ChannelGetInfo(w http.ResponseWriter, r *http.Request) {
	bytes, err := bindata.Asset("static/channel-view.html") // 根据地址获取对应内容
	if err != nil {
		fmt.Println(err)
		return
	}
	t, err := template.New("tpl").Parse(string(bytes)) // 比如用于模板处理
	channelId := r.FormValue("channelId")
	t.Execute(w, channelId)
}

func ChannelGetInfoJson(w http.ResponseWriter, r *http.Request) {
	channelId := r.FormValue("channelId")
	channel,err := service.GetChannelInfo(channelId)
	if err != nil {
		fmt.Println(err)
		return
	}
	channel_info, _ := json.Marshal(channel)
	w.Write(channel_info)
}

func ChannelFetch(w http.ResponseWriter, r *http.Request) {

}

func ChannelFetchJson(w http.ResponseWriter, r *http.Request) {
	channelId := r.FormValue("channelId")
	blockIndex := r.FormValue("blockIndex")
	if blockIndex == "" {
		blockIndex = "newest"
	}
	fetchResult,err := service.ChannelFetch(channelId,blockIndex)
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Write([]byte(service.WrapFetchResult(fetchResult)))
}

