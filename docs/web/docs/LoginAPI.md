# \LoginAPI

All URIs are relative to *http://localhost:8001*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AuthSvcV1LoginEmailPost**](LoginAPI.md#AuthSvcV1LoginEmailPost) | **Post** /auth-svc/v1/login/email | Login by email
[**AuthSvcV1LoginGoogleCallbackGet**](LoginAPI.md#AuthSvcV1LoginGoogleCallbackGet) | **Get** /auth-svc/v1/login/google/callback | Google OAuth callback
[**AuthSvcV1LoginGooglePost**](LoginAPI.md#AuthSvcV1LoginGooglePost) | **Post** /auth-svc/v1/login/google | Start Google OAuth login
[**AuthSvcV1LoginUsernamePost**](LoginAPI.md#AuthSvcV1LoginUsernamePost) | **Post** /auth-svc/v1/login/username | Login by username



## AuthSvcV1LoginEmailPost

> TokensPair AuthSvcV1LoginEmailPost(ctx).LoginByEmail(loginByEmail).Execute()

Login by email



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
	loginByEmail := *openapiclient.NewLoginByEmail(*openapiclient.NewLoginByEmailData("Type_example", *openapiclient.NewLoginByEmailDataAttributes("example@gmail.com", "StrongP@ssw0rd!"))) // LoginByEmail | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.LoginAPI.AuthSvcV1LoginEmailPost(context.Background()).LoginByEmail(loginByEmail).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LoginAPI.AuthSvcV1LoginEmailPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1LoginEmailPost`: TokensPair
	fmt.Fprintf(os.Stdout, "Response from `LoginAPI.AuthSvcV1LoginEmailPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1LoginEmailPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **loginByEmail** | [**LoginByEmail**](LoginByEmail.md) |  | 

### Return type

[**TokensPair**](TokensPair.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1LoginGoogleCallbackGet

> TokensPair AuthSvcV1LoginGoogleCallbackGet(ctx).Code(code).Execute()

Google OAuth callback



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
	code := "code_example" // string | OAuth authorization code returned by Google

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.LoginAPI.AuthSvcV1LoginGoogleCallbackGet(context.Background()).Code(code).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LoginAPI.AuthSvcV1LoginGoogleCallbackGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1LoginGoogleCallbackGet`: TokensPair
	fmt.Fprintf(os.Stdout, "Response from `LoginAPI.AuthSvcV1LoginGoogleCallbackGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1LoginGoogleCallbackGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **code** | **string** | OAuth authorization code returned by Google | 

### Return type

[**TokensPair**](TokensPair.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1LoginGooglePost

> AuthSvcV1LoginGooglePost(ctx).Execute()

Start Google OAuth login



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
	r, err := apiClient.LoginAPI.AuthSvcV1LoginGooglePost(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LoginAPI.AuthSvcV1LoginGooglePost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1LoginGooglePostRequest struct via the builder pattern


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


## AuthSvcV1LoginUsernamePost

> TokensPair AuthSvcV1LoginUsernamePost(ctx).LoginByUsername(loginByUsername).Execute()

Login by username



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
	loginByUsername := *openapiclient.NewLoginByUsername(*openapiclient.NewLoginByUsernameData("Type_example", *openapiclient.NewLoginByUsernameDataAttributes("example", "StrongP@ssw0rd!"))) // LoginByUsername | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.LoginAPI.AuthSvcV1LoginUsernamePost(context.Background()).LoginByUsername(loginByUsername).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LoginAPI.AuthSvcV1LoginUsernamePost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1LoginUsernamePost`: TokensPair
	fmt.Fprintf(os.Stdout, "Response from `LoginAPI.AuthSvcV1LoginUsernamePost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1LoginUsernamePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **loginByUsername** | [**LoginByUsername**](LoginByUsername.md) |  | 

### Return type

[**TokensPair**](TokensPair.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

