# \QrAPI

All URIs are relative to *http://localhost:8001*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AuthSvcV1LoginQrConfirmPost**](QrAPI.md#AuthSvcV1LoginQrConfirmPost) | **Post** /auth-svc/v1/login/qr/confirm | Confirm QR token
[**AuthSvcV1LoginQrGet**](QrAPI.md#AuthSvcV1LoginQrGet) | **Get** /auth-svc/v1/login/qr | Connect to QR login session



## AuthSvcV1LoginQrConfirmPost

> AuthSvcV1LoginQrConfirmPost(ctx).QRConfirm(qRConfirm).Execute()

Confirm QR token



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	qRConfirm := *openapiclient.NewQRConfirm(*openapiclient.NewQRConfirmData("Type_example", *openapiclient.NewQRConfirmDataAttributes("TODO"))) // QRConfirm | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.QrAPI.AuthSvcV1LoginQrConfirmPost(context.Background()).QRConfirm(qRConfirm).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `QrAPI.AuthSvcV1LoginQrConfirmPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1LoginQrConfirmPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **qRConfirm** | [**QRConfirm**](QRConfirm.md) |  | 

### Return type

 (empty response body)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1LoginQrGet

> string AuthSvcV1LoginQrGet(ctx).Execute()

Connect to QR login session



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.QrAPI.AuthSvcV1LoginQrGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `QrAPI.AuthSvcV1LoginQrGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1LoginQrGet`: string
	fmt.Fprintf(os.Stdout, "Response from `QrAPI.AuthSvcV1LoginQrGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1LoginQrGetRequest struct via the builder pattern


### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/event-stream, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

