package utils

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// 请求次数
	requestsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "myapp_requests_processed_total",
			Help: "The total number of processed requests",
		},
		[]string{"service", "version", "os"}, // 添加了三个标签：service, version, os
	)

	// 创建一个 Gauge 类型的指标，用于记录后台在捕获用户支付信息的时候开启的 Goroutine 数量
	paymentGoroutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "payment_goroutines",
			Help: "Number of Goroutines spawned during payment processing",
		},
	)

	// 支付响应时间
	paymentResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_response_time",
			Help:    "Distribution of time taken for payment processing",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "version", "os"}, // 添加了三个标签：service、version 和 os
	)

	// 创建一个带标签的 Summary 指标 统计 任务执行时间
	taskExecutionTime = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "cart_idle_time",
			Help:       "Distribution of idle time for shopping carts",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"service", "version", "os"}, // 添加了三个标签：service、version 和 os
	)

	// 创建一个带标签的 Counter 指标
	cartAndPaymentInteractionCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "interaction_counter",
			Help: "Count of interactions between shopping cart and payment module",
		},
		[]string{"service", "version", "os"}, // 添加了三个标签：service、version 和 os
	)
)

func CounterRequestProcess(service, version, os string) {
	// 增加指标
	requestsProcessed.WithLabelValues(service, version, os).Inc()
}

func PaymentGoroutinesInc() {
	paymentGoroutines.Inc()
}

func PaymentGoroutinesDec() {
	paymentGoroutines.Dec()
}

func RecordPaymentResponseTime(service, version, os string, duration float64) {
	// 根据标签值记录 histogram 数据
	paymentResponseTime.WithLabelValues(service, version, os).Observe(duration)
}

func TaskExecutionTime(service, version, os string, duration float64) {
	// 根据标签值记录 summary 数据
	taskExecutionTime.WithLabelValues(service, version, os).Observe(duration)
}

func Counterinteraction(service, version, os string) {
	cartAndPaymentInteractionCounter.WithLabelValues(service, version, os).Inc()
}

// RegisterMetrics 将所有指标注册到 Prometheus
func RegisterMetrics() {
	// 注册请求次数指标
	prometheus.MustRegister(requestsProcessed)

	// 注册支付成功率指标
	prometheus.MustRegister(paymentGoroutines)

	// 注册支付响应时间指标
	prometheus.MustRegister(paymentResponseTime)

	// 注册购物车空闲时间指标
	prometheus.MustRegister(taskExecutionTime)

	// 注册购物车和支付模块交互次数指标
	prometheus.MustRegister(cartAndPaymentInteractionCounter)
	fmt.Println("所有指标注册成功")
}

func PrometheusBoot(port int) {
	// 在 "/metrics" 路径上注册一个处理器，用于 Prometheus 的数据抓取
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		// 构造监听地址和端口，启动 HTTP 服务
		// 0.0.0.0 表示服务器将接受来自任何 IP 地址的连接，因此可以通过任何可用的 IP 地址和端口来访问该服务器。
		// 监听请求并返回静态内容，使用默认的 HTTP 处理逻辑来处理请求足够了，所以第二个参数nil
		err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(port), nil)
		// 如果启动失败，记录致命错误并退出
		if err != nil {
			fmt.Println("start fail")
		}
	}()

	// 注册指标
	RegisterMetrics()

	// 记录日志信息，表明监控服务已启动
	fmt.Println("监控启动，端口为：" + strconv.Itoa(port))
}
