package myscore

import (
	"context"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// Ensure that the plugin implements the interface.
var _ framework.ScorePlugin = &MyScorePlugin{}

// MyScorePlugin is a sample Score plugin that applies some custom logic.
type MyScorePlugin struct{}

// Name returns the name of the plugin.
func (pl *MyScorePlugin) Name() string {
	return "MyScorePlugin"
}

// Score is called on each node. It must return success and an integer
// indicating the rank of the node. The higher the rank, the better the node.
func (pl *MyScorePlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	klog.Infof("yeah! yeah! yeah! from myscore\n")
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
