package server

import (
	"context"

	gorillaMux "github.com/gorilla/mux"
	"github.com/juju/errors"
	rancherAuthProviders "github.com/rancher/rancher/pkg/auth/providers/publicapi"
	rancherAuthTokens "github.com/rancher/rancher/pkg/auth/tokens"
	rancherDynamicListener "github.com/rancher/rancher/pkg/dynamiclistener"
	rancherFilter "github.com/rancher/rancher/pkg/filter"
	rancherK8sProxy "github.com/rancher/rancher/pkg/k8sproxy"
	typesConfig "github.com/rancher/types/config"
	exporterApi "github.com/thxcode/kubernetes-event-exporter/pkg/api/server"
)

func Start(ctx context.Context, httpPort, httpsPort int, mgrContext *typesConfig.ScaledContext) error {
	authProvidersAPI, err := rancherAuthProviders.NewHandler(ctx, mgrContext)
	if err != nil {
		return errors.Annotate(err, "fail to new rancher auth providers API")
	}

	k8sProxy := rancherK8sProxy.New(mgrContext, mgrContext.Dialer)

	exporterAPI, err := exporterApi.New(ctx, mgrContext, k8sProxy)
	if err != nil {
		return errors.Annotate(err, "fail to new exporter API")
	}

	tokenAPI, err := rancherAuthTokens.NewAPIHandler(ctx, mgrContext)
	if err != nil {
		return errors.Annotate(err, "fail to new rancher auth tokens API")
	}

	rawAuthedAPIs := gorillaMux.NewRouter()
	rawAuthedAPIs.UseEncodedPath()
	rawAuthedAPIs.PathPrefix("/v3/identit").Handler(tokenAPI)
	rawAuthedAPIs.PathPrefix("/v3/token").Handler(tokenAPI)
	rawAuthedAPIs.PathPrefix("/v3").Handler(exporterAPI)

	authedHandler, err := rancherFilter.NewAuthenticationFilter(ctx, mgrContext, nil, rawAuthedAPIs)
	if err != nil {
		return errors.Annotate(err, "fail to create authentication handler")
	}

	root := gorillaMux.NewRouter()
	root.UseEncodedPath()

	root.PathPrefix("/v3-public").Handler(authProvidersAPI)
	root.PathPrefix("/v3").Handler(authedHandler)

	registerHealth(root)

	rancherDynamicListener.Start(ctx, mgrContext, httpPort, httpsPort, root)
	return nil
}
