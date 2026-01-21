# \AccountsAPI

All URIs are relative to *http://localhost:8001*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AuthSvcV1MeDelete**](AccountsAPI.md#AuthSvcV1MeDelete) | **Delete** /auth-svc/v1/me | Delete my account
[**AuthSvcV1MeEmailGet**](AccountsAPI.md#AuthSvcV1MeEmailGet) | **Get** /auth-svc/v1/me/email | Get my email data
[**AuthSvcV1MeGet**](AccountsAPI.md#AuthSvcV1MeGet) | **Get** /auth-svc/v1/me | Get my account
[**AuthSvcV1MePasswordPatch**](AccountsAPI.md#AuthSvcV1MePasswordPatch) | **Patch** /auth-svc/v1/me/password | Update password
[**AuthSvcV1MeUsernamePatch**](AccountsAPI.md#AuthSvcV1MeUsernamePatch) | **Patch** /auth-svc/v1/me/username | Update username



## AuthSvcV1MeDelete

> AuthSvcV1MeDelete(ctx).Execute()

Delete my account



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
	r, err := apiClient.AccountsAPI.AuthSvcV1MeDelete(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccountsAPI.AuthSvcV1MeDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MeDeleteRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1MeEmailGet

> AccountEmail AuthSvcV1MeEmailGet(ctx).Execute()

Get my email data



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
	resp, r, err := apiClient.AccountsAPI.AuthSvcV1MeEmailGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccountsAPI.AuthSvcV1MeEmailGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1MeEmailGet`: AccountEmail
	fmt.Fprintf(os.Stdout, "Response from `AccountsAPI.AuthSvcV1MeEmailGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MeEmailGetRequest struct via the builder pattern


### Return type

[**AccountEmail**](AccountEmail.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1MeGet

> Account AuthSvcV1MeGet(ctx).Execute()

Get my account



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
	resp, r, err := apiClient.AccountsAPI.AuthSvcV1MeGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccountsAPI.AuthSvcV1MeGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1MeGet`: Account
	fmt.Fprintf(os.Stdout, "Response from `AccountsAPI.AuthSvcV1MeGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MeGetRequest struct via the builder pattern


### Return type

[**Account**](Account.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1MePasswordPatch

> AuthSvcV1MePasswordPatch(ctx).UpdatePassword(updatePassword).Execute()

Update password



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
	updatePassword := *openapiclient.NewUpdatePassword(*openapiclient.NewUpdatePasswordData("Type_example", *openapiclient.NewUpdatePasswordDataAttributes("OldP@ssw0rd!", "StrongP@ssw0rd!"))) // UpdatePassword | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.AccountsAPI.AuthSvcV1MePasswordPatch(context.Background()).UpdatePassword(updatePassword).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccountsAPI.AuthSvcV1MePasswordPatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MePasswordPatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updatePassword** | [**UpdatePassword**](UpdatePassword.md) |  | 

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


## AuthSvcV1MeUsernamePatch

> Account AuthSvcV1MeUsernamePatch(ctx).UpdateUsername(updateUsername).Execute()

Update username



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
	updateUsername := *openapiclient.NewUpdateUsername(*openapiclient.NewUpdateUsernameData("Type_example", *openapiclient.NewUpdateUsernameDataAttributes("NewUsername_example"))) // UpdateUsername | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AccountsAPI.AuthSvcV1MeUsernamePatch(context.Background()).UpdateUsername(updateUsername).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccountsAPI.AuthSvcV1MeUsernamePatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1MeUsernamePatch`: Account
	fmt.Fprintf(os.Stdout, "Response from `AccountsAPI.AuthSvcV1MeUsernamePatch`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MeUsernamePatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updateUsername** | [**UpdateUsername**](UpdateUsername.md) |  | 

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

