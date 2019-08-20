package main

import (
	"fmt"
	"log"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

func main() {
	tracer, closer := initJaeger("Formatter")
	defer closer.Close()
	// 定义"/format/"的请求函数
	http.HandleFunc("/format/", func(w http.ResponseWriter, r *http.Request) {
		spanCTX, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header)) // 从http头获取上游span信息
		formatSpan := tracer.StartSpan("formatHandler", ext.RPCServerOption(spanCTX))                   //从上游span context创建新的span
		transID := formatSpan.BaggageItem("transid")                                                    //从Baggage获取业务ID
		defer formatSpan.Finish()

		str := fmt.Sprintf("Hello, %s!", r.FormValue("helloTo")) //格式化字符串
		formatSpan.SetTag("transid", transID)                    //设置tag记录业务ID信息

		ext.HTTPUrl.Set(formatSpan, r.URL.Path)  //定义tag记录http请求的URL
		ext.HTTPMethod.Set(formatSpan, r.Method) //定义tag记录http请求的方法
		formatSpan.LogFields(                    // 定义span日志信息
			otlog.String("event", "handle format"),
			otlog.String("value", str),
		)
		formatSpan.Tracer().Inject(formatSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header)) //将span信息注入http头
		w.Write([]byte(str))
	})
	log.Fatal(http.ListenAndServe(":10001", nil)) //启动http服务，监听在10001端口
}
