package flextopochecker

import (
	"testing"

	"github.com/stretchr/testify/assert"

	framework "github.com/kubewharf/godel-scheduler/pkg/framework/api"
	testing_helper "github.com/kubewharf/godel-scheduler/pkg/testing-helper"
)

func TestFlexTopoChecker(t *testing.T) {
	ftc := &FlexTopoChecker{}
	podInfo := framework.NewPodInfo(testing_helper.MakePod().Label("name", "dp").Obj())
	gotCode, gotMsg := ftc.VictimSearching(nil, podInfo, nil, nil, nil)
	assert.Equal(t, framework.PreemptionSucceed, gotCode)
	assert.Empty(t, gotMsg)
}
