package myscore

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// Ensure that the plugin implements the interface.
var _ framework.ScorePlugin = &MyScorePlugin{}

// MyScorePlugin is a sample Score plugin that applies some custom logic.
type MyScorePlugin struct{}

type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Value [2]interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// Name returns the name of the plugin.
func (pl *MyScorePlugin) Name() string {
	return "MyScorePlugin"
}

// Score is called on each node. It must return success and an integer
// indicating the rank of the node. The higher the rank, the better the node.
func (pl *MyScorePlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	klog.Infof("yeah! yeah! yeah! from myscore\n")

	// Prometheus クエリの構築
	baseURL := "http://prometheus-kube-prometheus-prometheus.monitoring.svc.cluster.local:9090/api/v1/query"
	queryParams := "sum(rate(request_total{app=\"frontend\",dst_service=\"adservice\", namespace=\"saiki\"}[60m]))"

	// クエリ文字列のエンコード
	encodedQuery := url.QueryEscape(queryParams)
	queryURL := baseURL + "?query=" + encodedQuery

	// HTTPリクエストの実行
	resp, err := http.Get(queryURL)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}
	defer resp.Body.Close()

	// HTTPステータスコードのチェック
	if resp.StatusCode != http.StatusOK {
		return 0, framework.NewStatus(framework.Error, "Received non-OK response from Prometheus: "+resp.Status)
	}

	// レスポンスの内容を読み込むg
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}
	// 応答の内容をログに出力
	klog.Info("Prometheus response: ", string(body))

	var result PrometheusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}

	// Result 配列が空でないことを確認
	if len(result.Data.Result) == 0 {
		klog.Info("No data in Prometheus response")
		return 0, framework.NewStatus(framework.Success)
	}

	// 変化率を取り出し、ログに出力
	changeRate := result.Data.Result[0].Value[1]
	klog.Infof("Change rate: %v\n", changeRate)

	return 50, framework.NewStatus(framework.Success)
}

// ScoreExtensions of the Score plugin.
func (pl *MyScorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// New initializes a new plugin and returns it.
func New(ctx context.Context, _ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &MyScorePlugin{}, nil
}
