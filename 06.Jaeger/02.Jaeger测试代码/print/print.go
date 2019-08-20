package main

import (
	"fmt"
	"log"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func main() {

	tracer, closer := initJaeger("Printer")
	defer closer.Close()
	// 定义"/print/"的请求函数
	http.HandleFunc("/print/", func(w http.ResponseWriter, r *http.Request) {
		spanCTX, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header)) // 从http头获取上游span信息
		printSpan := tracer.StartSpan("printHandler", ext.RPCServerOption(spanCTX))                     //从上游span context创建新的
		transID := printSpan.BaggageItem("transid")                                                     //从Baggage获取业务ID
		fmt.Println(transID)
		defer printSpan.Finish()

		str := r.FormValue("helloStr") //从请求中获取form信息
		println(str)
		printSpan.SetTag("transid", transID)                                                                              //设置tag记录业务ID信息
		printSpan.LogKV("event", "handle print", "value", str)                                                            //设置span日志信息
		ext.HTTPUrl.Set(printSpan, r.URL.Path)                                                                            //设置span tag记录请求URL
		ext.HTTPMethod.Set(printSpan, r.Method)                                                                           //设置span tag记录请求方法
		printSpan.Tracer().Inject(printSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header)) //将当前span信息注入http头
		w.Write([]byte(str))
	})
	log.Fatal(http.ListenAndServe(":10002", nil)) //启动web服务器
}
