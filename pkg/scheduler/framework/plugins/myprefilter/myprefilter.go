package myprefilter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// MyPreFilterPlugin は PreFilterPlugin インターフェースを実装するカスタムプラグインです。
type MyPreFilterPlugin struct {
	ServiceScores map[string]int64
}

// インターフェースの実装を確実にするために、コンパイル時の型チェックを行います。
var _ framework.PreFilterPlugin = &MyPreFilterPlugin{}

// Name はプラグインの名前を返します。
func (pl *MyPreFilterPlugin) Name() string {
	return "MyPreFilterPlugin"
}

// PrometheusResponse は Prometheus API のレスポンス構造を表します。
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				App string `json:"app"` // レスポンスに合わせて `Service` から `App` に変更
			} `json:"metric"`
			Value [2]interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// PreFilter はプラグインのメインロジックです。
func (pl *MyPreFilterPlugin) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) (*framework.PreFilterResult, *framework.Status) {
	klog.Infof("yah! yah! yah! from myprefilter\n")
	// ポッドのラベルからサービス名を取得する
	sn := p.Labels["app"]
	if sn == "" {
		klog.Info("MyPreFilter called for a pod without an 'app' label")
		return nil, framework.NewStatus(framework.Error, "Pod does not have an 'app' label")
	}

	// Prometheus クエリの構築
	baseURL := "http://prometheus-kube-prometheus-prometheus.monitoring.svc.cluster.local:9090/api/v1/query"
	queryParams := fmt.Sprintf("topk(5, sum(rate(request_total{app=\"%s\", namespace=\"saiki\", dst_service!=\"redis-cart\", dst_service!=\"otel-collector-collector\", dst_service!=\"jaeger\", dst_service!=\"prometheus-kube-prometheus-prometheus\", dst_service!=\"\"}[60m])) by (dst_service) OR sum(rate(request_total{dst_service=\"%s\", namespace=\"saiki\", app!=\"redis-cart\", app!=\"otelcollector\", app!=\"jaeger\", app!=\"\"}[60m])) by (app))", sn, sn)

	// クエリ文字列のエンコード
	encodedQuery := url.QueryEscape(queryParams)
	queryURL := baseURL + "?query=" + encodedQuery

	// クエリ文字列のログ出力
	klog.Info("Prometheus query: ", queryURL)

	// HTTPリクエストの実行
	resp, err := http.Get(queryURL)
	if err != nil {
		return nil, framework.NewStatus(framework.Error, err.Error())
	}
	defer resp.Body.Close()

	// HTTPステータスコードのチェック
	if resp.StatusCode != http.StatusOK {
		return nil, framework.NewStatus(framework.Error, "Received non-OK response from Prometheus: "+resp.Status)
	}

	// レスポンスの内容を読み込む
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, framework.NewStatus(framework.Error, err.Error())
	}

	// レスポンスの内容をログに出力
	klog.Info("Prometheus response: ", string(body))

	// Prometheusレスポンスの解析
	var result PrometheusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, framework.NewStatus(framework.Error, err.Error())
	}

	// サービススコアのマップを作成
	pl.ServiceScores = make(map[string]int64)
	score := int64(50)
	for i, res := range result.Data.Result {
		if i >= 5 {
			break
		}
		// `Service` ではなく `App` を使用
		pl.ServiceScores[res.Metric.App] = score
		score -= 10
	}

	if len(pl.ServiceScores) == 0 {
		return nil, framework.NewStatus(framework.Error, "No services found in Prometheus response")
	}

	// マップの内容をログに出力
	for service, score := range pl.ServiceScores {
		klog.Infof("Service: %s, Score: %d", service, score)
	}

	return nil, framework.NewStatus(framework.Success, "")
}

// PreFilterExtensions はプラグインが PreFilterExtensions を実装している場合に返します。
func (pl *MyPreFilterPlugin) PreFilterExtensions() framework.PreFilterExtensions {
	// この例では、PreFilterExtensions は実装されていませんが、必要に応じて実装を追加できます。
	return nil
}

// New は新しいプラグインのインスタンスを作成します。
func New(ctx context.Context, configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
	return &MyPreFilterPlugin{ServiceScores: make(map[string]int64)}, nil
}
