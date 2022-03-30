package controllers

import (
	"context"
	"testing"

	api "github.com/kuadrant/authorino/api/v1beta1"
	"github.com/kuadrant/authorino/pkg/cache"
	mock_cache "github.com/kuadrant/authorino/pkg/cache/mocks"
	"github.com/kuadrant/authorino/pkg/log"

	"github.com/golang/mock/gomock"
	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newStatusUpdateAuthConfig(authConfigLabels map[string]string) api.AuthConfig {
	return api.AuthConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AuthConfig",
			APIVersion: "authorino.kuadrant.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "auth-config-1",
			Namespace: "authorino",
			Labels:    authConfigLabels,
		},
		Spec: api.AuthConfigSpec{
			Hosts: []string{"echo-api"},
		},
		Status: api.AuthConfigStatus{
			Ready: false,
		},
	}
}

func newStatusUpdaterReconciler(client client.WithWatch, c cache.Cache) *AuthConfigStatusUpdater {
	return &AuthConfigStatusUpdater{
		Client: client,
		Logger: log.WithName("test").WithName("authconfigstatusupdater"),
		Cache:  c,
	}
}

func TestAuthConfigStatusUpdater_Reconcile(t *testing.T) {
	mockctrl := gomock.NewController(t)
	defer mockctrl.Finish()
	cache := mock_cache.NewMockCache(mockctrl)
	cache.EXPECT().FindKeys("authorino/auth-config-1").Return([]string{"echo-api"})
	authConfig := newStatusUpdateAuthConfig(map[string]string{})
	resourceName := types.NamespacedName{Namespace: authConfig.Namespace, Name: authConfig.Name}
	client := newTestK8sClient(&authConfig)
	reconciler := newStatusUpdaterReconciler(client, cache)

	result, err := reconciler.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: resourceName})

	assert.Equal(t, result, ctrl.Result{})
	assert.NilError(t, err)

	authConfigCheck := api.AuthConfig{}
	_ = client.Get(context.TODO(), resourceName, &authConfigCheck)
	assert.Check(t, authConfigCheck.Status.Ready)
}

func TestAuthConfigStatusUpdater_MissingWatchedAuthConfigLabels(t *testing.T) {
	mockctrl := gomock.NewController(t)
	defer mockctrl.Finish()
	cache := mock_cache.NewMockCache(mockctrl)
	cache.EXPECT().FindKeys("authorino/auth-config-1").Return([]string{"echo-api"})
	authConfig := newTestAuthConfig(map[string]string{"authorino.kuadrant.io/managed-by": "authorino"})
	resourceName := types.NamespacedName{Namespace: authConfig.Namespace, Name: authConfig.Name}
	client := newTestK8sClient(&authConfig)
	reconciler := newStatusUpdaterReconciler(client, cache)

	result, err := reconciler.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: resourceName})

	assert.Equal(t, result, ctrl.Result{})
	assert.NilError(t, err)

	authConfigCheck := api.AuthConfig{}
	_ = client.Get(context.TODO(), resourceName, &authConfigCheck)
	assert.Check(t, authConfigCheck.Status.Ready)
}

func TestAuthConfigStatusUpdater_MatchingAuthConfigLabels(t *testing.T) {
	mockctrl := gomock.NewController(t)
	defer mockctrl.Finish()
	cache := mock_cache.NewMockCache(mockctrl)
	cache.EXPECT().FindKeys("authorino/auth-config-1").Return([]string{"echo-api"})
	authConfig := newTestAuthConfig(map[string]string{"authorino.kuadrant.io/managed-by": "authorino"})
	resourceName := types.NamespacedName{Namespace: authConfig.Namespace, Name: authConfig.Name}
	client := newTestK8sClient(&authConfig)
	reconciler := newStatusUpdaterReconciler(client, cache)
	reconciler.LabelSelector = ToLabelSelector("authorino.kuadrant.io/managed-by=authorino")

	result, err := reconciler.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: resourceName})

	assert.Equal(t, result, ctrl.Result{})
	assert.NilError(t, err)

	authConfigCheck := api.AuthConfig{}
	_ = client.Get(context.TODO(), resourceName, &authConfigCheck)
	assert.Check(t, authConfigCheck.Status.Ready)
}

func TestAuthConfigStatusUpdater_UnmatchingAuthConfigLabels(t *testing.T) {
	mockctrl := gomock.NewController(t)
	defer mockctrl.Finish()
	cache := mock_cache.NewMockCache(mockctrl)
	authConfig := newTestAuthConfig(map[string]string{"authorino.kuadrant.io/managed-by": "other"})
	resourceName := types.NamespacedName{Namespace: authConfig.Namespace, Name: authConfig.Name}
	client := newTestK8sClient(&authConfig)
	reconciler := newStatusUpdaterReconciler(client, cache)
	reconciler.LabelSelector = ToLabelSelector("authorino.kuadrant.io/managed-by=authorino")

	result, err := reconciler.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: resourceName})

	assert.Equal(t, result, ctrl.Result{})
	assert.NilError(t, err)

	authConfigCheck := api.AuthConfig{}
	_ = client.Get(context.TODO(), resourceName, &authConfigCheck)
	assert.Check(t, !authConfigCheck.Status.Ready)
}

func TestAuthConfigStatusUpdater_NotReady(t *testing.T) {
	mockctrl := gomock.NewController(t)
	defer mockctrl.Finish()
	cache := mock_cache.NewMockCache(mockctrl)
	cache.EXPECT().FindKeys("authorino/auth-config-1").Return([]string{})
	authConfig := newStatusUpdateAuthConfig(map[string]string{})
	resourceName := types.NamespacedName{Namespace: authConfig.Namespace, Name: authConfig.Name}
	client := newTestK8sClient(&authConfig)
	reconciler := newStatusUpdaterReconciler(client, cache)

	result, err := reconciler.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: resourceName})

	assert.Check(t, result.Requeue)
	assert.NilError(t, err)

	authConfigCheck := api.AuthConfig{}
	_ = client.Get(context.TODO(), resourceName, &authConfigCheck)
	assert.Check(t, !authConfigCheck.Status.Ready)
}
