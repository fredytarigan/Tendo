package watcher

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/fredytarigan/Tendo/pkg/k8s"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretData struct {
	PublicKey string
	PrivateKey string
}

func GetSecret(kubeconfig string, secretNamespace string, secretName string) (SecretData, error) {
	var secretData SecretData

	client := k8s.GetKubernetesConfig(kubeconfig)

	secret, err := client.CoreV1().Secrets(secretNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		err := fmt.Errorf("secret %s not found in namespace %s", secretName, secretNamespace)
		return secretData, err

	} else if statusError, isStatus := err.(*errors.StatusError); isStatus { 
		err := fmt.Errorf("error getting secret %s", statusError.ErrStatus.Message)
		return secretData, err

	} else if err != nil {
		err := fmt.Errorf("unable to get secret %s with error: %s", secretName, err)
		return secretData, err

	} else {
		var publicKey string
		var privateKey string

		for key, value := range secret.Data {
			if key == "tls.crt" {
				encodedValue := base64.StdEncoding.EncodeToString([]byte(value))
				publicKey = encodedValue
			} else if key == "tls.key" {
				encodedValue := base64.StdEncoding.EncodeToString([]byte(value))
				privateKey = encodedValue
			} else {
				fmt.Printf("unsupported secret data")
			}
		}

		secretData.PublicKey = publicKey
		secretData.PrivateKey = privateKey

		return secretData, nil
	}
}

func CreateOpaqueSecret(kubeconfig string, secretNamespace string, secretName string, data string) error {
	client := k8s.GetKubernetesConfig(kubeconfig)

	_, err := client.CoreV1().Secrets(secretNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		// create the secret
		fmt.Printf("secret %s not found, creating a new one \n", secretName)

		secretClient := client.CoreV1().Secrets(secretNamespace)
		secret := &apiv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: secretName,
				Namespace: secretNamespace,
			},
			Type: "Opaque",
			StringData: map[string]string {
				"qcloud_cert_id": data,
			},
		}

		_, err := secretClient.Create(context.TODO(), secret, metav1.CreateOptions{})
		if err != nil {
			err := fmt.Errorf("unable to create opaque secret %s with error: %s", secretName, err)
			return err
		}
	}

	return nil
}