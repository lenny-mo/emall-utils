package tracer

import (
	"errors"
	"fmt"
	"io"

	"github.com/opentracing/opentracing-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
)

var (
	Tracer opentracing.Tracer
	Closer io.Closer
)

func InitTracer(serviceName, agentHostPort string) error {
	// 检查是否已经有Tracer和Closer，如果有则返回nil，表示已经初始化
	if Tracer != nil && Closer != nil {
		return nil
	}

	// 定义Jaeger的配置项
	cfg := jaegerConfig.Configuration{
		// 设置采样器类型为常量采样，参数为1
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		// 设置Reporter的配置，其中LogSpans为false，不记录Spans的日志，LocalAgentHostPort指定本地Agent的地址和端口
		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans:           false,
			LocalAgentHostPort: agentHostPort,
		},
		// 设置服务名
		ServiceName: serviceName,
	}

	// 使用配置项创建新的Tracer
	_tracer, _closer, err := cfg.NewTracer()
	if err != nil {
		// 如果创建Tracer出错，则打印错误信息并返回错误
		fmt.Println("Init GlobalJaegerTracer failed, err : ", err)
		return err
	}

	// 将新创建的Tracer和Closer赋值给全局变量Tracer和Closer
	Tracer = _tracer
	Closer = _closer

	// 返回nil，表示初始化成功
	return nil
}

func getParentSpan(operationName, traceId string, startIfNoParent bool) (span opentracing.Span, err error) {
	// 检查 Tracer 是否为空
	if Tracer == nil {
		err = errors.New("jaeger tracing error : Tracer is nil")
		fmt.Println(err)
		return
	}

	// 从传入的 traceId 中尝试提取父级 span 的上下文信息
	parentSpanCtx, err := Tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier{"UBER-TRACE-ID": traceId})
	if err != nil {
		// 如果无法提取父级 span 上下文信息，并且允许创建新的 span，则创建新的根 span
		if startIfNoParent {
			span = Tracer.StartSpan(operationName)
		}
	} else {
		// 如果能够成功提取到父级 span 上下文信息，则以此为基础创建一个子 span
		span = Tracer.StartSpan(operationName, opentracing.ChildOf(parentSpanCtx))
	}

	// 重置错误为 nil，表示没有发生错误
	err = nil
	return
}

func StartSpan(operationName, parentSpanTraceId string, startIfNoParent bool) (span opentracing.Span, spanTraceId string, err error) {
	// 调用获取父级 Span 的函数，并得到返回的 span 及错误信息
	plainParentSpan, err := getParentSpan(operationName, parentSpanTraceId, startIfNoParent)

	// 如果出现错误或者无法获取父级 Span，则打印信息并直接返回
	if err != nil || plainParentSpan == nil {
		fmt.Println("No span return")
		return
	}

	// 创建一个空的 map，用作载荷承载者
	m := map[string]string{}
	carrier := opentracing.TextMapCarrier(m)

	// 将当前 Span 的上下文信息注入到 carrier 中
	err = Tracer.Inject(plainParentSpan.Context(), opentracing.TextMap, carrier)

	// 如果注入过程出现错误，打印错误信息并直接返回
	if err != nil {
		fmt.Println("jaeger tracing inject error : ", err)
		return
	}

	// 从 carrier 中获取 "uber-trace-id" 的值，并赋给 spanTraceId
	spanTraceId = carrier["uber-trace-id"]
	span = plainParentSpan

	// 返回 span 及 spanTraceId
	return
}

func FinishSpan(span opentracing.Span) {
	if span != nil {
		span.Finish()
	}
}

func SpanSetTag(span opentracing.Span, tagname string, tagvalue interface{}) {
	if span != nil {
		span.SetTag(tagname, tagvalue)
	}
}
