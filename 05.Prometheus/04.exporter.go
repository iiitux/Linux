package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 定义性能指标
var (
	httpStatus = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_status",
		Help: "Http response status.",
	},
		[]string{"method", "url", "code"},
	)
)

// 注册性能指标
func init() {
	prometheus.MustRegister(httpStatus)
}

// http请求默认处理函数
func defaultFunc(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/err":
		httpStatus.WithLabelValues(r.Method, r.URL.Path, "404").Inc()
		w.WriteHeader(404)
	case "/errsu":
		httpStatus.WithLabelValues(r.Method, r.URL.Path, "503").Inc()
		w.WriteHeader(503)
	default:
		fmt.Fprintf(w, "Request URL: %s;\nRequest Method: %s;\nResponse StatusCode: %s\n", r.URL.Path, r.Method, "200")
		httpStatus.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
	}
}

func main() {
	http.HandleFunc("/", defaultFunc)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":10000", nil))
}
