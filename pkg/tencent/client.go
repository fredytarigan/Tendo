package tencent

import (
	"fmt"
	"os"

	"github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common"
)

func BuildCredentials() (common.CredentialIface, error) {
	var creds common.CredentialIface

	secretID := os.Getenv("TENCENTCLOUD_SECRET_ID")
	secretKey := os.Getenv("TENCENTCLOUD_SECRET_KEY")

	if secretID != "" && secretKey != "" {
		creds = common.NewCredential(
			secretID,
			secretKey,
		)

		return creds, nil
	}

	providerChain := common.DefaultProviderChain()
	creds, err := providerChain.GetCredential()

	if err != nil {
		err := fmt.Errorf("unable to build credentials for tencent cloud client with error: %s", err)
		return creds, err
	}

	return creds, nil
}

