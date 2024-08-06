package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/fredytarigan/Tendo/pkg/tencent"
	"github.com/fredytarigan/Tendo/pkg/tendo/config"
)

func Start(ctx context.Context, c *config.Config, kubeconfig string) error {
	tick := time.NewTicker(c.WatchInterval * time.Second)
	defer tick.Stop()

	for {

		for _, item := range c.WatchTargets {
			go func() {
				err := RunLoop(ctx, c, kubeconfig, item)
				if err != nil {
					fmt.Printf("%s \n", err)
				}
			}()
		}

		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
			continue
		}
	}
}

func RunLoop(ctx context.Context, c *config.Config, kubeconfig string, item config.WatchConfig) error {
	secret, err := GetSecret(kubeconfig, item.SecretNamespace, item.SecretName)

	if err != nil {
		return err
	}

	tencentCreds, err := tencent.BuildCredentials()

	if err != nil {
		return err
	}

	var certificateRequestTypes []tencent.CertificateResourceType
	for _, value := range item.CertificateResourceTypes {
		result := tencent.CertificateResourceType {
			Name: value.Name,
			Regions: value.Regions,
		}
		certificateRequestTypes = append(certificateRequestTypes, result)
	}


	tencentSSLCertificate := tencent.TencentSSLCertificate {
		Context: ctx,
		Credentials: tencentCreds,
		Region: item.CertificateRegion,
		CertificateID: item.CertificateID,
		CertificateName: item.CertificateName,
		CertificateResourceTypes: certificateRequestTypes,
		PublicKey: secret.PublicKey,
		PrivateKey: secret.PrivateKey,
	}

	client, err := tencentSSLCertificate.BuildClient()
	if err != nil {
		return err
	}

	cert, err := tencentSSLCertificate.GetCertificateData(client)
	if err != nil {
		return err
	}

	// create opaque secret if not exists
	CreateOpaqueSecret(kubeconfig, item.SecretNamespace, item.OpaqueSecretName, tencentSSLCertificate.CertificateID)

	// compare secret with cert
	certChanged := false

	if secret.PublicKey != cert.CertificatePublicKey || secret.PrivateKey != cert.CertificatePrivateKey {
		certChanged = true
	}

	if !certChanged {
		fmt.Printf("certificate in secret %s is up to date with certificate stored in tencent cloud \n", item.SecretName)
		fmt.Println("not doing anything for now")

		return nil
	}

	// if secret is not matched
	// we update certificates in tencent cloud
	fmt.Printf("certificate in secret %s is not matched with certificated stored in tencent cloud with name %s \n", item.SecretName, item.CertificateName)
	fmt.Printf("updating certificate in tencent cloud for certificate with name %s \n", item.CertificateName)

	err = tencentSSLCertificate.UpdateCertificateDetail(client)
	if err != nil {
		return err
	}

	// wait for 5 seconds, for deployment started
	time.Sleep(5 * time.Second)

	_, err = tencentSSLCertificate.WatchCertificateUpdateStatus(client)
	if err != nil {
		return err
	}

	// fmt.Printf("Watch certificate update result: %s", newCertID)

	return nil
}