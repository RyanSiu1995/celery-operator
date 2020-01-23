package e2e

import (
	goctx "golang.org/x/net/context"
	"testing"
	"time"

	"github.com/RyanSiu1995/celery-operator/pkg/apis"
	celeryprojectv4 "github.com/RyanSiu1995/celery-operator/pkg/apis/celeryproject/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	basicCelery := &celeryprojectv4.Celery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "basic",
			Namespace: namespace,
		},
		Spec: celeryprojectv4.CelerySpec{
			Broker: celeryprojectv4.CeleryBroker{
				Type: celeryprojectv4.RedisBroker,
			},
			Workers: []celeryprojectv4.CeleryWorker{},
		},
	}

	err = f.Client.Create(goctx.TODO(), basicCelery, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	if err != nil {
		t.Fatal(err)
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "basic-broker-deployment", 1, time.Second*5, time.Second*30)
	if err != nil {
		t.Fatal(err)
	}
}
