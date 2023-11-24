package myprefilter

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// MyPreFilterPlugin は PreFilterPlugin インターフェースを実装するカスタムプラグインです。
type MyPreFilterPlugin struct{}

// Name はプラグインの名前を返します。
func (pl *MyPreFilterPlugin) Name() string {
	klog.Infof("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&\n")
	return "MyPreFilterPlugin"
}

// PreFilter はプラグインのメインロジックです。
func (pl *MyPreFilterPlugin) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) *framework.Status {
	klog.Infof("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
	klog.Infof("\nPreFilter called for pod: %s", p.Name)
	// ここに必要なロジックを実装します。
	return framework.NewStatus(framework.Success, "")
}

func (pl *MyPreFilterPlugin) PreFilterExtensions() framework.PreFilterExtensions {
	return nil // または実際の拡張機能を提供する場合はそれを返す
}

// New は新しいプラグインのインスタンスを作成します。
func New(ctx context.Context, _ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &MyPreFilterPlugin{}, nil
}
