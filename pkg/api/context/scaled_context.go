package context

import (
	"context"

	normanStoreProxy "github.com/rancher/norman/store/proxy"
	rancherAuthProviders "github.com/rancher/rancher/pkg/auth/providers/common"
	rancherDialer "github.com/rancher/rancher/pkg/dialer"
	rancherK8scheck "github.com/rancher/rancher/pkg/k8scheck"
	rancherRBAC "github.com/rancher/rancher/pkg/rbac"
	typesConfig "github.com/rancher/types/config"
	k8sRest "k8s.io/client-go/rest"
)

func BuildScaledContext(ctx context.Context, rancherBackendK8sConfig *k8sRest.Config, httpsPort int) (*typesConfig.ScaledContext, error) {
	mgrCtx, err := typesConfig.NewScaledContext(*rancherBackendK8sConfig)
	if err != nil {
		return nil, err
	}
	mgrCtx.LocalConfig = rancherBackendK8sConfig

	if err := rancherK8scheck.Wait(ctx, *rancherBackendK8sConfig); err != nil {
		return nil, err
	}

	simpleClientGetter, err := normanStoreProxy.NewClientGetterFromConfig(*rancherBackendK8sConfig)
	if err != nil {
		return nil, err
	}
	mgrCtx.ClientGetter = simpleClientGetter
	mgrCtx.AccessControl = rancherRBAC.NewAccessControl(mgrCtx.RBAC)

	dialerFactory, err := rancherDialer.NewFactory(mgrCtx)
	if err != nil {
		return nil, err
	}
	mgrCtx.Dialer = dialerFactory

	userManager, err := rancherAuthProviders.NewUserManager(mgrCtx)
	if err != nil {
		return nil, err
	}
	mgrCtx.UserManager = userManager

	return mgrCtx, nil
}
