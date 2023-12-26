package myfilter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// MyFilterPlugin は FilterPlugin インターフェースを実装するカスタムプラグインです。
type MyFilterPlugin struct {
	// ここにプラグイン固有の状態やプロパティを追加します。
}

// コンパイル時の型チェックを行います。
var _ framework.FilterPlugin = &MyFilterPlugin{}

// Name はプラグインの名前を返します。
func (pl *MyFilterPlugin) Name() string {
	return "MyFilterPlugin"
}

func (f *MyFilterPlugin) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	// klog.Infof("hello hello hello from myfilter\n")

	// 環境変数からCPU使用率のリミット値を取得
	cpuLimitStr := os.Getenv("CPU_USAGE_LIMIT")
	if cpuLimitStr == "" {
		cpuLimitStr = "0.15" // デフォルト値
	}

	cpuLimit, err := strconv.ParseFloat(cpuLimitStr, 64)
	if err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("Invalid CPU usage limit: %v", err))
	}

	// Prometheus API URL
	prometheusURL := fmt.Sprintf("http://prometheus-kube-prometheus-prometheus.monitoring.svc.cluster.local:9090/api/v1/query?query=instance:node_cpu_utilisation:rate5m{instance=\"%s:9100\"}", nodeInfo.Node().Status.Addresses[0].Address)

	// HTTPクライアントの作成
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(prometheusURL)
	if err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("Failed to call Prometheus API: %v", err))
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("Failed to read response body: %v", err))
	}

	var result PrometheusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("Failed to unmarshal Prometheus response: %v", err))
	}

	if len(result.Data.Result) > 0 {
		// CPU使用率の値を取得して型アサーション
		cpuUtilizationStr, ok := result.Data.Result[0].Value[1].(string)
		if !ok {
			return framework.NewStatus(framework.Error, "Invalid CPU utilization format")
		}

		cpuUtilization, err := strconv.ParseFloat(cpuUtilizationStr, 64)
		if err != nil {
			return framework.NewStatus(framework.Error, fmt.Sprintf("Invalid CPU utilization value: %v", err))
		}

		if cpuUtilization > cpuLimit {
			return framework.NewStatus(framework.UnschedulableAndUnresolvable, "CPU utilization exceeds limit")
		}

		klog.Infof("Node: %s, CPU Utilization: %f, Limit: %f", nodeInfo.Node().Status.Addresses[0].Address, cpuUtilization, cpuLimit)
	} else {
		klog.Infof("no result: %s\n", nodeInfo.Node().Status.Addresses[0].Address)
	}

	return framework.NewStatus(framework.Success)
}

type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Container string `json:"container"`
				Endpoint  string `json:"endpoint"`
				Instance  string `json:"instance"`
				Job       string `json:"job"`
				Namespace string `json:"namespace"`
				Pod       string `json:"pod"`
				Service   string `json:"service"`
			} `json:"metric"`
			Value [2]interface{} `json:"value"` // 型を [2]interface{} に変更
		} `json:"result"`
	} `json:"data"`
}

// New は新しい MyFilterPlugin のインスタンスを作成します。
func New(ctx context.Context, _ runtime.Object, f framework.Handle) (framework.Plugin, error) {
	return &MyFilterPlugin{}, nil
}
