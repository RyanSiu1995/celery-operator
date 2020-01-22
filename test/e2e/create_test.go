package e2e

import (
	"testing"
	"time"

	"github.com/RyanSiu1995/celery-operator/pkg/apis"
	celeryprojectv4 "github.com/RyanSiu1995/celery-operator/pkg/apis/celeryproject/v4"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
)

func TestCreateOperator(t *testing.T) {
	cleanupTimeout, _ := time.ParseDuration("60s")
	cleanupRetryInterval, _ := time.ParseDuration("5s")

	celery := &celeryprojectv4.Celery{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, celery)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	err = ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}

	// get namespace
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for memcached-operator to be ready
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "celery-operator", 1, time.Second*5, time.Second*30)
	if err != nil {
		t.Fatal(err)
	}
}
