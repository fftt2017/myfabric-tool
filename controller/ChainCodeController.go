package controller

import (
	"encoding/json"
	"myfabric-tool/service"
	"net/http"
	"net/url"
)

func ChainCodeListInstalled(w http.ResponseWriter, r *http.Request ){
	chaincodeList,_ :=service.ListInstalledChaincode()
	result := map[string]interface{}{
		"code":  "0",
		"msg":   "",
		"count": len(chaincodeList),
		"data":  chaincodeList,
	}
	chaincodeListJson , _ := json.Marshal(result)
	w.Write(chaincodeListJson)
}
func ChainCodeListInstantiate(w http.ResponseWriter, r *http.Request ){
	channelID := r.FormValue("channelID")
	chaincodeList, _ :=service.ListInstantiateChaincode(channelID)
	result := map[string]interface{}{
		"code":  "0",
		"msg":   "",
		"count": len(chaincodeList),
		"data":  chaincodeList,
	}
	chaincodeListJson , _ := json.Marshal(result)
	w.Write(chaincodeListJson)
}
func ChainCodeQuery(w http.ResponseWriter, r *http.Request ){
	channelID := r.FormValue("channelID")
	chaincodeName := r.FormValue("chaincodeName")
	args := r.FormValue("args")
	args,_ = url.QueryUnescape(args)
	result,err :=service.ChaincodeQuery(channelID,chaincodeName,args)
	if err!=nil {
		w.Write([]byte(err.Error()))
	}else{
		w.Write([]byte(result))
	}
}
func ChainCodeInvoke(w http.ResponseWriter, r *http.Request ){
	channelID := r.FormValue("channelID")
	chaincodeName := r.FormValue("chaincodeName")
	args := r.FormValue("args")
	args,_ = url.QueryUnescape(args)
	error := service.ChaincodeInvoke(channelID,chaincodeName,args)
	if error != nil{
		w.Write([]byte(error.Error()))
	}else{
		w.Write([]byte("success"))
	}
}
