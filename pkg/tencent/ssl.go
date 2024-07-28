package tencent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common"
	tencentCloudSDKError "github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common/profile"

	sslCertificate "github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/ssl/v20191205"
)

type TencentSSLCertificate struct {
	Context 					context.Context
	Credentials 			 	common.CredentialIface
	Region 						string
	CertificateID 	 			 string
	CertificateName  			 string
	CertificateResourceTypes	 []CertificateResourceType
	PublicKey					string
	PrivateKey					string
}

type CertificateResourceType struct {
	Name 		string		`mapstructure:"name"`
	Regions 	[]string	`mapstructure:"regions"`
}

type CertificateData struct {
	CertificateID	string	`mapstructure:"CertificateID"`
}

type CertificateDetail struct {
	CertificatePublicKey	string	`mapstructure:"CertificatePublicKey"`
	CertificatePrivateKey	string	`mapstructure:"CertificatePrivateKey"`
}

type CertificateNotFoundError struct {
	Message string
}

func (e *CertificateNotFoundError) Error() string {
	return e.Message
}

func (t *TencentSSLCertificate) BuildClient() (*sslCertificate.Client, error) {
	var client *sslCertificate.Client

	profile := profile.NewClientProfile()
	profile.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"

	client, err := sslCertificate.NewClient(t.Credentials, t.Region, profile)
	if err != nil {
		err := fmt.Errorf("unable to build tencent cloud ssl certificate client with error: %s", err)
		return client, err
	}

	return client, nil
}

func (t *TencentSSLCertificate) GetCertificateID(client *sslCertificate.Client) (string, error) {
	var certData []CertificateData

	// build request
	request := sslCertificate.NewDescribeCertificatesRequest()
	request.Limit = common.Uint64Ptr(1)

	request.SearchKey = &t.CertificateName
	request.CertificateStatus = common.Uint64Ptrs([]uint64{1})

	response, err := client.DescribeCertificatesWithContext(t.Context, request)
	if err != nil {
		return "", err
	}

	if *response.Response.TotalCount < 1 {
		msg := fmt.Sprintf("certificate with name or id %s not found", t.CertificateName)
		err := fmt.Errorf("%w", &CertificateNotFoundError {
			Message: msg,
		})
		return "", err
	}

	cert, err := json.Marshal(response.Response.Certificates)
	if err != nil {
		err := fmt.Errorf("invalid response while getting certificate with error: %s", err)
		return "", err
	}

	err = json.Unmarshal([]byte(cert), &certData)
	if err != nil {
		err := fmt.Errorf("unable to parse certificates response with error: %s", err)
		return "", err
	}

	return certData[0].CertificateID, nil
}

func (t *TencentSSLCertificate) GetCertificateDetail(client *sslCertificate.Client) (CertificateDetail, error)  {
	var certDetail CertificateDetail

	// build request
	request := sslCertificate.NewDescribeCertificateDetailRequest()
	request.CertificateId = &t.CertificateID

	response, err := client.DescribeCertificateDetailWithContext(t.Context, request)
	if err != nil {
		return certDetail, err
	}

	cert, err := json.Marshal(response.Response)
	if err != nil {
		err := fmt.Errorf("invalid response while getting certificate id %s detail with error: %s", t.CertificateID, err)
		return certDetail, err
	}

	err = json.Unmarshal([]byte(cert), &certDetail)
	if err != nil {
		err := fmt.Errorf("unable to parse certificate detail response with error: %s", err)
		return certDetail, err
	}

	certDetail.CertificatePublicKey = base64.StdEncoding.EncodeToString([]byte(certDetail.CertificatePublicKey))
	certDetail.CertificatePrivateKey = base64.StdEncoding.EncodeToString([]byte(certDetail.CertificatePrivateKey))

	return certDetail, nil
}

func (t *TencentSSLCertificate) GetCertificateData(client *sslCertificate.Client) (CertificateDetail, error) {
	var certDetail CertificateDetail

	// get certificate id
	if t.CertificateID == "" {
		var certificateID string
		certificateID, err := t.GetCertificateID(client)
		if err != nil {
			certNotFoundError := &CertificateNotFoundError{}
			if errors.As(err, &certNotFoundError) {
				fmt.Printf("certificate with name %s not found \n", t.CertificateName)
				fmt.Printf("will create a new certificate in tencent cloud with name %s \n", t.CertificateName)

				certificateID, err = t.CreateCertificate(client)
				if err != nil {
					return certDetail, err
				}
			} else {
				return certDetail, err
			}
		}

		t.CertificateID = certificateID
	}

	certDetail, err := t.GetCertificateDetail(client)
	if err != nil {
		return certDetail, err
	}

	return certDetail, nil
}

func (t *TencentSSLCertificate) CreateCertificate(client *sslCertificate.Client) (string, error) {
	var certData CertificateData

	publicKeyByte, err := base64.StdEncoding.DecodeString(t.PublicKey)
	if err != nil {
		err := fmt.Errorf("unable to decode public key for certificate")
		return "", err
	}
	publicKeyString := string(publicKeyByte)

	privateKeyByte, err := base64.StdEncoding.DecodeString(t.PrivateKey)
	if err != nil {
		err := fmt.Errorf("unable to decode private key for certificate")
		return "", err
	}
	privateKeyString := string(privateKeyByte)


	// build request
	repeatable := new(bool)
	*repeatable = true

	request := sslCertificate.NewUploadCertificateRequest()
	request.CertificatePublicKey = &publicKeyString
	request.CertificatePrivateKey = &privateKeyString
	request.Alias = &t.CertificateName
	request.Repeatable = repeatable

	response, err := client.UploadCertificateWithContext(t.Context, request)
	if err != nil {
		return "", err
	}

	cert, err := json.Marshal(response.Response)
	if err != nil {
		err := fmt.Errorf("invalid response while creating certificate with name %s with error: %s", t.CertificateName, err)
		return "", err
	}

	err = json.Unmarshal([]byte(cert), &certData)
	if err != nil {
		err := fmt.Errorf("unable to parse certificates response with error: %s", err)
		return "", err
	}

	return certData.CertificateID, nil
}

func (t *TencentSSLCertificate) UpdateCertificateDetail(client *sslCertificate.Client) error {
	publicKeyByte, err := base64.StdEncoding.DecodeString(t.PublicKey)
	if err != nil {
		err := fmt.Errorf("unable to decode public key for certificate")
		return err
	}
	publicKeyString := string(publicKeyByte)

	privateKeyByte, err := base64.StdEncoding.DecodeString(t.PrivateKey)
	if err != nil {
		err := fmt.Errorf("unable to decode private key for certificate")
		return err
	}
	privateKeyString := string(privateKeyByte)

	var resourceTypes []string
	for _, value := range t.CertificateResourceTypes {
		resourceTypes = append(resourceTypes, value.Name)
	}

	var resourceTypesRegions []*sslCertificate.ResourceTypeRegions
	for _, value := range t.CertificateResourceTypes {
		result := sslCertificate.ResourceTypeRegions {
			ResourceType: common.StringPtr(value.Name),
			Regions: common.StringPtrs(value.Regions),
		}

		resourceTypesRegions = append(resourceTypesRegions, &result)
	}

	// build request
	request := sslCertificate.NewUpdateCertificateInstanceRequest()
	request.OldCertificateId = common.StringPtr(t.CertificateID)
	request.CertificateId = common.StringPtr(t.CertificateID)
	request.CertificatePublicKey = common.StringPtr(publicKeyString)
	request.CertificatePrivateKey = common.StringPtr(privateKeyString)
	request.ResourceTypes = common.StringPtrs(resourceTypes)
	request.ResourceTypesRegions = resourceTypesRegions
	request.Repeatable = common.BoolPtr(true)
	request.AllowDownload = common.BoolPtr(true)
	request.ExpiringNotificationSwitch = common.Uint64Ptr(0)

	response, err := client.UpdateCertificateInstanceWithContext(t.Context, request)
	if _, ok := err.(*tencentCloudSDKError.TencentCloudSDKError); ok {
		err := fmt.Errorf("failed to update certificate %s with error: %s", t.CertificateName, err)
		return err
	}

	if err != nil {
		err := fmt.Errorf("failed to update certificate %s with error: %s", t.CertificateName, err)
		return err
	}

	cert, err := json.Marshal(response.Response)
	if err != nil {
		err := fmt.Errorf("invalid response while updating certificate with name %s with error: %s", t.CertificateName, err)
		return err
	}

	fmt.Printf("%s \n", cert)
	fmt.Println("")

	return nil
}