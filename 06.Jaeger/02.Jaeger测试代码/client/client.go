package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func main() {
	if len(os.Args) != 2 {
		panic("Error: Expecting one argument!")
	} //判断输入是否正确
	tracer, closer := initJaeger("client") // 初始化tracer实例
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer) //将tracer设置全局

	input := os.Args[1] //获取终端输入信息

	span := tracer.StartSpan("hello")       //启动span
	span.SetTag("HelloTo", input)           // 设置标签
	span.SetBaggageItem("transid", "xxxxx") // TransationID模拟跟踪业务
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span) //将spanContext信息存入ctx

	helloStr := formatString(ctx, input) // 调用formatString函数
	printHello(ctx, helloStr)
}

// formatString函数，发起http Get请求
func formatString(ctx context.Context, str string) string {
	span, _ := opentracing.StartSpanFromContext(ctx, "formatString") // 从上下文获取上游span信息，并创建新的span
	fmt.Println(span.BaggageItem("transid"))
	defer span.Finish()
	v := url.Values{}
	v.Set("helloTo", str)
	url := "http://localhost:10001/format?" + v.Encode() //生成完整的HTTP Get请求URL
	req, err := http.NewRequest("GET", url, nil)         //定义http request实例
	if err != nil {
		panic(err.Error())
	}
	ext.HTTPUrl.Set(span, url)                                                                                //设置tag记录URL请求
	ext.HTTPMethod.Set(span, "GET")                                                                           //设置tag记录请求方法
	span.LogKV("event", "formatString", "value", str)                                                         //设置日志
	span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header)) //将span信息注入http header传递给下游web服务器
	resp, err := httpClientDo(req)                                                                            //执行http请求
	if err != nil {
		panic(err.Error())
	}

	respStr := string(resp)

	return respStr
}

// printHello发起http请求，打印字符串
func printHello(ctx context.Context, str string) {
	span, _ := opentracing.StartSpanFromContext(ctx, "printHello") // 从上下文获取上游span信息，并创建新的span
	defer span.Finish()
	fmt.Println(span.BaggageItem("transid"))

	v := url.Values{}
	v.Set("helloStr", str)
	url := "http://localhost:10002/print?" + v.Encode() //生成完整的HTTP Get请求URL
	req, err := http.NewRequest("GET", url, nil)        //定义http request实例
	if err != nil {
		panic(err.Error())
	}

	ext.SpanKindRPCClient.Set(span)                                                                           //设置span为client类型
	ext.HTTPUrl.Set(span, url)                                                                                //设置tag记录URL请求
	ext.HTTPMethod.Set(span, "GET")                                                                           //设置tag记录请求方法
	span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header)) //将span信息注入http header传递给下游web服务器
	_, err = httpClientDo(req)                                                                                //执行http请求
	if err != nil {
		panic(err.Error())
	}
}
