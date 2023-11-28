package myprefilter

import (
	"context"
	"io/ioutil"
	"net/http"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// MyPreFilterPlugin は PreFilterPlugin インターフェースを実装するカスタムプラグインです。
type MyPreFilterPlugin struct{}

// インターフェースの実装を確実にするために、コンパイル時の型チェックを行います。
var _ framework.PreFilterPlugin = &MyPreFilterPlugin{}

// Name はプラグインの名前を返します。
func (pl *MyPreFilterPlugin) Name() string {
	return "MyPreFilterPlugin"
}

// PreFilter はプラグインのメインロジックです。
func (pl *MyPreFilterPlugin) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) (*framework.PreFilterResult, *framework.Status) {
	klog.Infof("yah! yah! yah! from myprefilter\n")
	// ポッドのラベルからサービス名を取得する
	sn := p.Labels["app"]
	if sn != "" {
		klog.Infof("PreFilter called for service: %s", sn)
	} else {
		klog.Info("PreFilter called for a pod without an 'app' label")
	}
	// ここに必要なロジックを実装します。
	// この例では、特に PreFilterResult を変更する必要はないので nil を返します。

	// api通信
	// Prometheus サービスへのクエリを実行
	queryURL := "http://prometheus-kube-prometheus-prometheus.monitoring.svc.cluster.local:9090/api/v1/query?query=up"
	resp, err := http.Get(queryURL)
	if err != nil {
		// HTTP エラーをハンドル
		return nil, framework.NewStatus(framework.Error, err.Error())
	}
	defer resp.Body.Close()

	// レスポンスの内容を読み込む
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// 読み込みエラーをハンドル
		return nil, framework.NewStatus(framework.Error, err.Error())
	}

	// 応答をログに出力
	klog.Info(string(body))

	return nil, framework.NewStatus(framework.Success, "")
}

// PreFilterExtensions はプラグインが PreFilterExtensions を実装している場合に返します。
func (pl *MyPreFilterPlugin) PreFilterExtensions() framework.PreFilterExtensions {
	// この例では、PreFilterExtensions は実装されていませんが、必要に応じて実装を追加できます。
	return nil
}

// New は新しいプラグインのインスタンスを作成します。
func New(ctx context.Context, configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
	return &MyPreFilterPlugin{}, nil
}
