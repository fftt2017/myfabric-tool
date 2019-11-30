package main

import (
	. "myfabric-tool/bindata"
	. "myfabric-tool/controller"
	"html/template"
	"log"
	"net/http"
)

func main() {
	//to run 'go-bindata-assetfs  -o "bindata/bindata.go" -pkg bindata static/...' in terminal
	ConfigRouter()
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("start server failed, reson is : ", err)
	}
}
func ConfigRouter(){
	http.Handle("/", http.FileServer(AssetFS()))
	http.HandleFunc("/index", Index)

	http.HandleFunc("/channel/list",ChannelList)
	http.HandleFunc("/channel/listJson",ChannelListJson)
	http.HandleFunc("/channel/getInfo",ChannelGetInfo)
	http.HandleFunc("/channel/getInfoJson",ChannelGetInfoJson)
	http.HandleFunc("/channel/fetch", ChannelFetch)
	http.HandleFunc("/channel/fetchJson", ChannelFetchJson)
}
func Index(w http.ResponseWriter, r *http.Request){
	t, _ := template.ParseFiles("static/index.html")
	//执行模板
	t.Execute(w,nil)
}
