package tencent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fredytarigan/Tendo/pkg/tendo/logger"
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

type CertifiateUpdateStatus struct {
	TotalCount			int								`json:"TotalCount"`
	DeployRecordLists 	[]CertificateDeployRecord		 `json:"DeployRecordList"`
	RequestID 			string 							`json:"RequestId"`
}

type CertificateDeployRecord struct {
	ID 				int			`json:"Id"`
	CertID 			string 		`json:"CertId"`
	OldCertID 		string 		`json:"OldCertId"`
	ResourceTypes 	[]string 	`json:"ResourceTypes"`
	Status 			int 		`json:"Status"`
	CreateTime 		string 		`json:"CreateTime"`
	UpdateTime 		string 		`json:"UpdateTime"`
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
				logger.Logger.Info(fmt.Sprintf("certificate with name %s not found", t.CertificateName))
				logger.Logger.Info(fmt.Sprintf("will create a new certificate in tencent cloud with name %s", t.CertificateName))

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

	_, err = json.Marshal(response.Response)
	if err != nil {
		err := fmt.Errorf("invalid response while updating certificate with name %s with error: %s", t.CertificateName, err)
		return err
	}

	return nil
}

func (t *TencentSSLCertificate) WatchCertificateUpdateStatus(client *sslCertificate.Client) (string, error) {
	updateFinished := false

	for !updateFinished {
		certDeployRecordList, err := t.DescribeCertificateUpdateStatus(client)
		if err != nil {
			return "", err
		}

		var wg sync.WaitGroup
		status := []int{}

		for _, item := range certDeployRecordList {

			logger.Logger.Info("Checking for certificate deployment status")

			wg.Add(1)

			go func(item CertificateDeployRecord) {
				defer wg.Done()

				status = append(status, item.Status)
			}(item)
		}

		wg.Wait()

		if allStatusIsDone(status) {
			logger.Logger.Info("All deployment is completed")

			for _, item := range certDeployRecordList {
				_, err := t.DeleteCertificate(client, item.OldCertID)
				if err != nil {
					fmt.Printf("%s", err)
					continue
				}

				_, err = t.ModifyCertificateName(client, item.CertID, t.CertificateName)
				if err != nil {
					fmt.Printf("%s", err)
					continue
				}
			}
			
			break
		}

		logger.Logger.Info("Not all deployment is finished, so we are waiting for all deployment to completed")
		time.Sleep(5 * time.Second)
	}

	return "", nil
}

func (t *TencentSSLCertificate) DescribeCertificateUpdateStatus(client *sslCertificate.Client) ([]CertificateDeployRecord, error) {
	var certificateUpdateStatus CertifiateUpdateStatus
	var certificateDeployRecord []CertificateDeployRecord

	request := sslCertificate.NewDescribeHostUpdateRecordRequest()
	request.OldCertificateId = common.StringPtr(t.CertificateID)

	response, err := client.DescribeHostUpdateRecordWithContext(t.Context, request)
	if _, ok := err.(*tencentCloudSDKError.TencentCloudSDKError); ok {
		err := fmt.Errorf("failed to get certificate %s update with error: %s", t.CertificateID, err)
		return certificateDeployRecord, err
	}

	if err != nil {
		err := fmt.Errorf("failed to get certificate %s update with error: %s", t.CertificateID, err)
		return certificateDeployRecord, err
	}

	cert, err := json.Marshal(response.Response)
	if err != nil {
		err :=  fmt.Errorf("invalid response while getting certificate update with name %s with error: %s", t.CertificateID, err)
		return certificateDeployRecord, err
	}

	err = json.Unmarshal([]byte(cert), &certificateUpdateStatus)
	if err != nil {
		err := fmt.Errorf("unable to parse response with error :%s", err)
		return certificateDeployRecord, err
	}

	if certificateUpdateStatus.TotalCount > 0 {
		for _, status := range certificateUpdateStatus.DeployRecordLists {
			fmt.Printf("Status: %v", status)
			fmt.Println("")
		}

		return certificateUpdateStatus.DeployRecordLists, nil
	} 

	err = fmt.Errorf("certificate update status is not found")
	return certificateDeployRecord, err
}

func allStatusIsDone(a []int) bool {
	for i := 1; i < len(a); i++ {
		if a[i] != 1 {
			return false
		}
	}

	return true
}

func (t *TencentSSLCertificate) DeleteCertificate(client *sslCertificate.Client, certID string) (bool, error) {
	request := sslCertificate.NewDeleteCertificateRequest()
	request.CertificateId = common.StringPtr(certID)

	_, err := client.DeleteCertificateWithContext(t.Context, request)
	if _, ok := err.(*tencentCloudSDKError.TencentCloudSDKError); ok {
		err := fmt.Errorf("failed to remove certificate %s with error: %s", certID, err)
		return false, err
	}

	if err != nil {
		err := fmt.Errorf("failed to remove certificate %s with error: %s", certID, err)
		return false, err
	}

	return true, nil
}

func (t *TencentSSLCertificate) ModifyCertificateName(client *sslCertificate.Client, certID string, name string) (bool, error) {
	request := sslCertificate.NewModifyCertificateAliasRequest()
	request.CertificateId = common.StringPtr(certID)
	request.Alias = common.StringPtr(name)

	_, err := client.ModifyCertificateAliasWithContext(t.Context, request)
	if _, ok := err.(*tencentCloudSDKError.TencentCloudSDKError); ok {
		err := fmt.Errorf("failed to change certificate name in id %s with error: %s", certID, err)
		return false, err
	}

	if err != nil {
		err := fmt.Errorf("failed to change certificate name in id %s with error: %s", certID, err)
		return false, err
	}

	return true, nil
}