package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request){
	//解析模板文件
	t, _ := template.ParseFiles("static/index.html")
	//执行模板
	t.Execute(w, r.FormValue("msg"))
}

func List(w http.ResponseWriter, r *http.Request) {
	bytes, err := Asset("static/index.html")    // 根据地址获取对应内容
	if err != nil {
		fmt.Println(err)
		return
	}
	t, err := template.New("tpl").Parse(string(bytes))      // 比如用于模板处理
	t.Execute(w, r.FormValue("msg"))
}
