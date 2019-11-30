package controller

import (
	"encoding/json"
	"fmt"
	"html/template"
	"myfabric-tool/bindata"
	"myfabric-tool/chain/client"
	"myfabric-tool/model"
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

	channelList, _ := service.ListChannels()

	/*result := model.ChannelList{
		Code:  "0",
		Msg:   "",
		Count: 2,
		Data:  channelList,
	}*/
	result := map[string]interface{}{
		"code":  "0",
		"msg":   "6",
		"count": 2,
		"data":  channelList,
	}
	b, _ := json.Marshal(result)
	w.Write(b)
}

func ChannelGetInfo(w http.ResponseWriter, r *http.Request) {
	bytes, err := bindata.Asset("static/channelList.html") // 根据地址获取对应内容
	if err != nil {
		fmt.Println(err)
		return
	}
	t, err := template.New("tpl").Parse(string(bytes)) // 比如用于模板处理
	t.Execute(w, r.FormValue("msg"))
}

func ChannelGetInfoJson(w http.ResponseWriter, r *http.Request) {
	channel := model.Channel{
		Name: "channel1",
	}
	b, _ := json.Marshal(channel)
	w.Write(b)
}

func ChannelFetch(w http.ResponseWriter, r *http.Request) {

}

func ChannelFetchJson(w http.ResponseWriter, r *http.Request) {

}

func getChainClient() error {
	client.InitConfig()
	cc, err := client.NewClient()
	if err != nil {
		return err
	}
	fmt.Println("chainclient: %v", cc)
	return nil
}
