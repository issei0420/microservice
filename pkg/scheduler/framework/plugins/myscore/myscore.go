package myscore

import (
	"context"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Ensure that the plugin implements the interface.
var _ framework.ScorePlugin = &MyScorePlugin{}

// MyScorePlugin is a sample Score plugin that applies some custom logic.
type MyScorePlugin struct{}

// Name returns the name of the plugin.
func (pl *MyScorePlugin) Name() string {
	return "MyScorePlugin"
}

type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				App        string `json:"app"`
				DstService string `json:"dst_service"` // DstServiceフィールドを追加
			} `json:"metric"`
			Value [2]interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// Score is called on each node. It must return success and an integer
// indicating the rank of the node. The higher the rank, the better the node.
func (pl *MyScorePlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	// klog.Infof("yeah! yeah! yeah! from myscore\n")

	// Pod のサービス名を取得
	sn := pod.Labels["app"]
	if sn == "" {
		klog.Info("Pod does not have an 'app' label")
		return 0, framework.NewStatus(framework.Success)
	}

	// Prometheus クエリの構築
	baseURL := "http://prometheus-kube-prometheus-prometheus.monitoring.svc.cluster.local:9090/api/v1/query"
	queryParams := fmt.Sprintf("topk(5, sum(rate(request_total{app=\"%s\", namespace=\"saiki\", dst_service!=\"redis-cart\", dst_service!=\"otel-collector-collector\", dst_service!=\"jaeger\", dst_service!=\"prometheus-kube-prometheus-prometheus\", dst_service!=\"\"}[30m])) by (dst_service) OR sum(rate(request_total{dst_service=\"%s\", namespace=\"saiki\", app!=\"redis-cart\", app!=\"otelcollector\", app!=\"jaeger\", app!=\"\"}[30m])) by (app))", sn, sn)

	// クエリ文字列のエンコード
	encodedQuery := url.QueryEscape(queryParams)
	queryURL := baseURL + "?query=" + encodedQuery

	// クエリ文字列のログ出力
	// klog.Info("service now on scheduling: ", sn)

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

	// レスポンスの内容を読み込む
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}

	// レスポンスの内容をログに出力
	// klog.Info("Prometheus response: ", string(body))

	// Prometheusレスポンスの解析
	var result PrometheusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}

	// レスポンスが空の場合、nodeScoreを0として返す
	if len(result.Data.Result) == 0 {
		klog.Info("No results found in Prometheus response, returning nodeScore 0")
		return 0, framework.NewStatus(framework.Success)
	}

	// サービススコアのマップを作成
	serviceScores := make(map[string]int64)
	score := int64(30)
	for i, res := range result.Data.Result {
		if i >= 3 {
			break
		}
		// `app` ラベルが存在するかチェックし、なければ `dst_service` ラベルを使用
		serviceKey := res.Metric.App
		if serviceKey == "" {
			serviceKey = res.Metric.DstService
		}

		serviceScores[serviceKey] = score
		score -= 10
	}

	// if len(serviceScores) == 0 {
	// 	klog.Info("No service found in Prometheus response")
	// 	return 0, framework.NewStatus(framework.Success)
	// }

	// マップの内容をログに出力
	for service, score := range serviceScores {
		klog.Infof("Service: %s, Score: %d", service, score)
	}

	// calcScore呼び出し
	nodeScore := calcScore(nodeName, serviceScores)

	return nodeScore, framework.NewStatus(framework.Success)
}

func calcScore(nodeName string, serviceScores map[string]int64) int64 {
	// InClusterConfigはクラスタ内で実行されているPod用にKubernetes APIサーバーへの接続設定を取得します
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting cluster config: %v\n", err)
		os.Exit(1)
	}

	// クライアントセットの作成
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating clientset: %v\n", err)
		os.Exit(1)
	}

	// 指定されたノード上のPodをリストアップ
	pods, err := clientset.CoreV1().Pods("saiki").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing pods on node %s: %v\n", nodeName, err)
		os.Exit(1)
	}

	// totalScoreの初期化
	var totalScore int64 = 0

	// リストアップされたPodの情報を出力し、スコアを計算
	for _, pod := range pods.Items {
		appLabel := pod.Labels["app"]
		fmt.Printf("NodeName: %s    App Label: %s\n", nodeName, appLabel)

		// serviceScoresマップからスコアを取得し、totalScoreに加算
		if score, exists := serviceScores[appLabel]; exists {
			totalScore += score
		}
	}

	return totalScore
}

// ScoreExtensions of the Score plugin.
func (pl *MyScorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// New initializes a new plugin and returns it.
func New(ctx context.Context, _ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &MyScorePlugin{}, nil
}
