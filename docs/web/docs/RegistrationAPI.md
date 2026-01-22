# \RegistrationAPI

All URIs are relative to *http://localhost:8001*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AuthSvcV1RegistrationAdminPost**](RegistrationAPI.md#AuthSvcV1RegistrationAdminPost) | **Post** /auth-svc/v1/registration/admin | Register a new admin account
[**AuthSvcV1RegistrationPost**](RegistrationAPI.md#AuthSvcV1RegistrationPost) | **Post** /auth-svc/v1/registration/ | Register a new account



## AuthSvcV1RegistrationAdminPost

> Account AuthSvcV1RegistrationAdminPost(ctx).RegistrationAdmin(registrationAdmin).Execute()

Register a new admin account



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
	registrationAdmin := *openapiclient.NewRegistrationAdmin(*openapiclient.NewRegistrationAdminData("Type_example", *openapiclient.NewRegistrationAdminDataAttributes("example1312@gmail.com", "adminUser", "StrongP@ssw0rd!", "admin"))) // RegistrationAdmin | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegistrationAPI.AuthSvcV1RegistrationAdminPost(context.Background()).RegistrationAdmin(registrationAdmin).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistrationAPI.AuthSvcV1RegistrationAdminPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1RegistrationAdminPost`: Account
	fmt.Fprintf(os.Stdout, "Response from `RegistrationAPI.AuthSvcV1RegistrationAdminPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1RegistrationAdminPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **registrationAdmin** | [**RegistrationAdmin**](RegistrationAdmin.md) |  | 

### Return type

[**Account**](Account.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1RegistrationPost

> AuthSvcV1RegistrationPost(ctx).Registration(registration).Execute()

Register a new account



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
	registration := *openapiclient.NewRegistration(*openapiclient.NewRegistrationData("Type_example", *openapiclient.NewRegistrationDataAttributes("example@gmail.com", "user123", "StrongP@ssw0rd!"))) // Registration | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RegistrationAPI.AuthSvcV1RegistrationPost(context.Background()).Registration(registration).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegistrationAPI.AuthSvcV1RegistrationPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1RegistrationPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **registration** | [**Registration**](Registration.md) |  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

