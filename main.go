package main

import (
	"html/template"
	"log"
	. "myfabric-tool/bindata"
	"myfabric-tool/config"
	. "myfabric-tool/controller"
	"net/http"
)

func main() {

	if err := config.LoadConfig("config-first-network.yaml"); err != nil {
		log.Fatalf("load config failed: %s", err)
	}

	//to run 'go-bindata-assetfs  -o "bindata/bindata.go" -pkg bindata static/...' in terminal
	ConfigRouter()
	//client.RefreshDefaultClient("org0","node1","org1","node1","org1","user")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("start server failed, reson is : ", err)
	}
}
func ConfigRouter() {
	http.Handle("/", http.FileServer(AssetFS()))
	http.HandleFunc("/index", Index)

	http.HandleFunc("/channel/list", ChannelList)
	http.HandleFunc("/channel/listJson", ChannelListJson)
	http.HandleFunc("/channel/getInfo", ChannelGetInfo)
	http.HandleFunc("/channel/getInfoJson", ChannelGetInfoJson)
	http.HandleFunc("/channel/fetch", ChannelFetch)
	http.HandleFunc("/channel/fetchJson", ChannelFetchJson)
	http.HandleFunc("/getOrderPeerConfig",OrderPeerConfigJson)
	http.HandleFunc("/switchNode",SwitchNode)
	http.HandleFunc("/chaincode/listInstalled",ChainCodeListInstalled)
	http.HandleFunc("/chaincode/listInstantiate",ChainCodeListInstantiate)
	http.HandleFunc("/chaincode/query",ChainCodeQuery)
	http.HandleFunc("/chaincode/invoke",ChainCodeInvoke)
}
func Index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/index.html")
	//执行模板
	t.Execute(w, nil)
}
