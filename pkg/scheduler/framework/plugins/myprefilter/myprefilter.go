package myprefilter

import (
	"context"

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

// PreFilter はプラグインのメインロジックです。
func (pl *MyPreFilterPlugin) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) (*framework.PreFilterResult, *framework.Status) {
	klog.Infof("yah! yah! yah! from myprefilter\n")

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
