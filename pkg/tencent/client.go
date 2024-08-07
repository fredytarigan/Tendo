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

	// if we run in tke, use DefaultTkeOIDCRoleArnProvider
	if os.Getenv("TKE_REGION") != "" && os.Getenv("TKE_PROVIDER_ID") != "" && os.Getenv("TKE_WEB_IDENTITY_TOKEN_FILE") != "" && os.Getenv("TKE_ROLE_ARN") != "" {
		provider, err := common.DefaultTkeOIDCRoleArnProvider()

		if err == nil {
			creds, err := provider.GetCredential()

			if err == nil {
				return creds, nil
			}
		}
	}

	providerChain := common.NewProviderChain(
		[]common.Provider{
			common.DefaultEnvProvider(),
			common.DefaultProfileProvider(),
			common.DefaultCvmRoleProvider(),
		},
	)
	creds, err := providerChain.GetCredential()

	if err != nil {
		err := fmt.Errorf("unable to build credentials for tencent cloud client with error: %s", err)
		return creds, err
	}

	return creds, nil
}

