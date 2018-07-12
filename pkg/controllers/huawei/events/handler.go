package events

import (
	"strings"

	"github.com/juju/errors"
	rancherKubeconfig "github.com/rancher/rancher/pkg/kubeconfig"
	rancherSettings "github.com/rancher/rancher/pkg/settings"
	typesHuaWei "github.com/rancher/types/apis/cloud.huawei.com/v3"
	typesManagement "github.com/rancher/types/apis/management.cattle.io/v3"
	typesConfig "github.com/rancher/types/config"
	"github.com/thxcode/kubernetes-event-exporter/pkg/exporters"
	k8sApisMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClientcmd "k8s.io/client-go/tools/clientcmd"
	k8sClientcmdApi "k8s.io/client-go/tools/clientcmd/api"
)

type eventLifecycle struct {
	mgrContext                *typesConfig.ScaledContext
	eventExportConfigTemplate *exporters.EventExporterConfig
	exporters                 map[string]*exporters.EventExporter
	clusterEvent              typesHuaWei.HuaWeiClusterEventInterface
	clusterEventLister        typesHuaWei.HuaWeiClusterEventLister
}

func (l *eventLifecycle) Sync(key string, cluster *typesManagement.Cluster) error {
	if cluster != nil {
		return l.startEventExporter(cluster)
	}
	return nil
}

func (l *eventLifecycle) Create(cluster *typesManagement.Cluster) (*typesManagement.Cluster, error) {
	// no-op because the sync function will take care of it
	return cluster, nil
}

func (l *eventLifecycle) Updated(cluster *typesManagement.Cluster) (*typesManagement.Cluster, error) {
	// no-op because the sync function will take care of it
	return cluster, nil
}

func (l *eventLifecycle) Remove(cluster *typesManagement.Cluster) (*typesManagement.Cluster, error) {
	if cluster != nil {
		err := l.stopEventExporter(cluster)
		if err != nil {
			return nil, err
		}
	}

	return cluster, nil
}

func (l *eventLifecycle) terminateEventExporters() {
	if len(l.exporters) != 0 {
		for _, exporter := range l.exporters {
			if exporter != nil {
				exporter.Stop()
			}
		}
	}
}

func (l *eventLifecycle) startEventExporter(cluster *typesManagement.Cluster) error {
	clusterID := cluster.Name
	clusterName := cluster.Spec.DisplayName

	isActived := true
	if cluster.DeletionTimestamp != nil {
		isActived = false
	} else {
		for _, condition := range cluster.Status.Conditions {
			if condition.Type == "Pending" && condition.Status == "Unknown" {
				isActived = false
				break
			}
		}
	}

	if _, ok := l.exporters[clusterID]; !ok && isActived {
		// ensure a system user
		clusterUser, err := l.mgrContext.UserManager.EnsureUser("system://"+clusterID, "System account for Cluster "+clusterID)
		if err != nil {
			return errors.Annotatef(err, "can't ensure user for cluster %s", clusterName)
		}
		clusterUserID := clusterUser.Name

		// take a token for user
		token, err := l.mgrContext.UserManager.EnsureToken("agent-"+clusterUserID, "", clusterUserID)
		if err != nil {
			return errors.Annotatef(err, "can't ensure token for cluster %s", clusterName)
		}

		config, err := k8sClientcmd.BuildConfigFromKubeconfigGetter("", func() (*k8sClientcmdApi.Config, error) {
			url := rancherSettings.ServerURL.Get()
			if strings.HasPrefix(url, "https://") {
				url = url[8:]
			} else if strings.HasPrefix(url, "http://") {
				url = url[7:]
			}

			kubeConfigFile, err := rancherKubeconfig.ForTokenBased(clusterName, clusterID, url, clusterUserID, token)
			if err != nil {
				return nil, err
			}

			return k8sClientcmd.Load([]byte(kubeConfigFile))
		})
		if err != nil {
			return errors.Annotatef(err, "can't get Kubernetes rest config for cluster %s", clusterName)
		}

		eventExporterConfig := *l.eventExportConfigTemplate
		eventExporterConfig.WatchingK8sConfig = config
		eventExporter, err := exporters.NewEventExporter(eventExporterConfig)
		if err != nil {
			return errors.Annotatef(err, "can't create exporter for cluster %s", clusterName)
		}

		eventExporter.Start()
		l.exporters[clusterID] = eventExporter

		eventLog, _ := l.clusterEventLister.Get(k8sApisMetaV1.NamespaceAll, clusterID)
		if eventLog == nil {
			eventLog = &typesHuaWei.HuaWeiClusterEvent{
				ObjectMeta: k8sApisMetaV1.ObjectMeta{
					Name: clusterID,
				},
				Spec: typesHuaWei.HuaWeiClusterEventSpec{
					DisplayName: clusterName,
					ClusterName: clusterID,
				},
				Status: typesHuaWei.HuaWeiClusterEventStatus{
					IsRemoved: false,
				},
			}

			_, err := l.clusterEvent.Create(eventLog)
			if err != nil {
				return errors.Annotatef(err, "can't create event log for cluster %s", clusterName)
			}
		}
	}

	return nil
}

func (l *eventLifecycle) stopEventExporter(cluster *typesManagement.Cluster) error {
	if cluster.DeletionTimestamp != nil {
		clusterID := cluster.Name

		if eventExporter, ok := l.exporters[clusterID]; ok {
			delete(l.exporters, clusterID)
			eventExporter.Stop()
		}

		eventLogName := "e-" + clusterID
		eventLog, _ := l.clusterEventLister.Get(k8sApisMetaV1.NamespaceAll, eventLogName)
		if eventLog != nil {
			eventLog.Status = typesHuaWei.HuaWeiClusterEventStatus{
				IsRemoved: true,
			}

			_, err := l.clusterEvent.Update(eventLog)
			if err != nil {
				return errors.Annotatef(err, "can't create event log for cluster %s", eventLog.Spec)
			}
		}
	}

	return nil
}

// func (l *eventLifecycle) toRESTConfig(cluster *typesManagement.Cluster) (*k8sRest.Config, error) {
// 	if cluster == nil {
// 		return nil, nil
// 	}
//
// 	if cluster.DeletionTimestamp != nil {
// 		return nil, nil
// 	}
//
// 	if cluster.Spec.Internal {
// 		return l.mgrContext.LocalConfig, nil
// 	}
//
// 	if cluster.Status.APIEndpoint == "" || cluster.Status.CACert == "" || cluster.Status.ServiceAccountToken == "" {
// 		return nil, nil
// 	}
//
// 	if !typesManagement.ClusterConditionProvisioned.IsTrue(cluster) {
// 		return nil, nil
// 	}
//
// 	u, err := url.Parse(cluster.Status.APIEndpoint)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	caBytes, err := base64.StdEncoding.DecodeString(cluster.Status.CACert)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	clusterDialer, err := l.mgrContext.Dialer.ClusterDialer(cluster.Name)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var tlsDialer typesDialer.Dialer
// 	if cluster.Status.Driver == typesManagement.ClusterDriverRKE {
// 		tlsDialer, err = nameIgnoringTLSDialer(clusterDialer, caBytes)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
//
// 	rc := &k8sRest.Config{
// 		Host:        u.String(),
// 		BearerToken: cluster.Status.ServiceAccountToken,
// 		TLSClientConfig: k8sRest.TLSClientConfig{
// 			CAData: caBytes,
// 		},
// 		Timeout: 30 * time.Second,
// 		WrapTransport: func(rt http.RoundTripper) http.RoundTripper {
// 			if ht, ok := rt.(*http.Transport); ok {
// 				ht.DialContext = nil
// 				ht.DialTLS = tlsDialer
// 				ht.Dial = clusterDialer
// 			}
// 			return rt
// 		},
// 	}
//
// 	return rc, nil
// }
//
// func nameIgnoringTLSDialer(dialer typesDialer.Dialer, caBytes []byte) (typesDialer.Dialer, error) {
// 	rkeVerify, err := verifyIgnoreDNSName(caBytes)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	tlsConfig := &tls.Config{
// 		// Use custom TLS validate that validates the cert chain, but not the server.  This should be secure because
// 		// we use a private per cluster CA always for RKE
// 		InsecureSkipVerify:    true,
// 		VerifyPeerCertificate: rkeVerify,
// 	}
//
// 	return func(network, address string) (net.Conn, error) {
// 		rawConn, err := dialer(network, address)
// 		if err != nil {
// 			return nil, err
// 		}
// 		tlsConn := tls.Client(rawConn, tlsConfig)
// 		if err := tlsConn.Handshake(); err != nil {
// 			rawConn.Close()
// 			return nil, err
// 		}
// 		return tlsConn, err
// 	}, nil
// }
//
// func verifyIgnoreDNSName(caCertsPEM []byte) (func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error, error) {
// 	rootCAs := x509.NewCertPool()
// 	if len(caCertsPEM) > 0 {
// 		caCerts, err := cert.ParseCertsPEM(caCertsPEM)
// 		if err != nil {
// 			return nil, err
// 		}
// 		for _, cert := range caCerts {
// 			rootCAs.AddCert(cert)
// 		}
// 	}
//
// 	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
// 		certs := make([]*x509.Certificate, len(rawCerts))
// 		for i, asn1Data := range rawCerts {
// 			cert, err := x509.ParseCertificate(asn1Data)
// 			if err != nil {
// 				return fmt.Errorf("failed to parse cert")
// 			}
// 			certs[i] = cert
// 		}
//
// 		opts := x509.VerifyOptions{
// 			Roots:         rootCAs,
// 			CurrentTime:   time.Now(),
// 			DNSName:       "",
// 			Intermediates: x509.NewCertPool(),
// 		}
//
// 		for i, cert := range certs {
// 			if i == 0 {
// 				continue
// 			}
// 			opts.Intermediates.AddCert(cert)
// 		}
// 		_, err := certs[0].Verify(opts)
// 		return err
// 	}, nil
// }
