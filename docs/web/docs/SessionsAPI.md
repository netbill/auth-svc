# \SessionsAPI

All URIs are relative to *http://localhost:8001*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AuthSvcV1MeLogoutPost**](SessionsAPI.md#AuthSvcV1MeLogoutPost) | **Post** /auth-svc/v1/me/logout | Logout
[**AuthSvcV1MeSessionsDelete**](SessionsAPI.md#AuthSvcV1MeSessionsDelete) | **Delete** /auth-svc/v1/me/sessions | Delete my sessions
[**AuthSvcV1MeSessionsGet**](SessionsAPI.md#AuthSvcV1MeSessionsGet) | **Get** /auth-svc/v1/me/sessions | Get my sessions
[**AuthSvcV1MeSessionsSessionIdDelete**](SessionsAPI.md#AuthSvcV1MeSessionsSessionIdDelete) | **Delete** /auth-svc/v1/me/sessions/{session_id} | Delete my session
[**AuthSvcV1MeSessionsSessionIdGet**](SessionsAPI.md#AuthSvcV1MeSessionsSessionIdGet) | **Get** /auth-svc/v1/me/sessions/{session_id} | Get my session
[**AuthSvcV1RefreshPost**](SessionsAPI.md#AuthSvcV1RefreshPost) | **Post** /auth-svc/v1/refresh | Refresh session



## AuthSvcV1MeLogoutPost

> AuthSvcV1MeLogoutPost(ctx).Execute()

Logout



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
	r, err := apiClient.SessionsAPI.AuthSvcV1MeLogoutPost(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SessionsAPI.AuthSvcV1MeLogoutPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MeLogoutPostRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1MeSessionsDelete

> AuthSvcV1MeSessionsDelete(ctx).Execute()

Delete my sessions



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
	r, err := apiClient.SessionsAPI.AuthSvcV1MeSessionsDelete(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SessionsAPI.AuthSvcV1MeSessionsDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MeSessionsDeleteRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1MeSessionsGet

> AccountSessionsCollection AuthSvcV1MeSessionsGet(ctx).PageLimit(pageLimit).PageOffset(pageOffset).FilterActive(filterActive).SortLastUsed(sortLastUsed).Execute()

Get my sessions



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
	pageLimit := int32(56) // int32 | Max number of items to return (optional)
	pageOffset := int32(56) // int32 | Number of items to skip (optional)
	filterActive := true // bool | Filter sessions by active status. - `true` — only active (non-deleted) sessions - `false` — only deleted sessions - omit — all sessions (default)  (optional)
	sortLastUsed := "sortLastUsed_example" // string | Sort sessions by last used date. - `desc` — most recently used first (default) - `asc` — least recently used first  (optional) (default to "desc")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SessionsAPI.AuthSvcV1MeSessionsGet(context.Background()).PageLimit(pageLimit).PageOffset(pageOffset).FilterActive(filterActive).SortLastUsed(sortLastUsed).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SessionsAPI.AuthSvcV1MeSessionsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1MeSessionsGet`: AccountSessionsCollection
	fmt.Fprintf(os.Stdout, "Response from `SessionsAPI.AuthSvcV1MeSessionsGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MeSessionsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **pageLimit** | **int32** | Max number of items to return | 
 **pageOffset** | **int32** | Number of items to skip | 
 **filterActive** | **bool** | Filter sessions by active status. - &#x60;true&#x60; — only active (non-deleted) sessions - &#x60;false&#x60; — only deleted sessions - omit — all sessions (default)  | 
 **sortLastUsed** | **string** | Sort sessions by last used date. - &#x60;desc&#x60; — most recently used first (default) - &#x60;asc&#x60; — least recently used first  | [default to &quot;desc&quot;]

### Return type

[**AccountSessionsCollection**](AccountSessionsCollection.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1MeSessionsSessionIdDelete

> AuthSvcV1MeSessionsSessionIdDelete(ctx, sessionId).Execute()

Delete my session



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
	sessionId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // uuid.UUID | Session ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SessionsAPI.AuthSvcV1MeSessionsSessionIdDelete(context.Background(), sessionId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SessionsAPI.AuthSvcV1MeSessionsSessionIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sessionId** | **uuid.UUID** | Session ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MeSessionsSessionIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1MeSessionsSessionIdGet

> AccountSession AuthSvcV1MeSessionsSessionIdGet(ctx, sessionId).Execute()

Get my session



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
	sessionId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // uuid.UUID | Session ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SessionsAPI.AuthSvcV1MeSessionsSessionIdGet(context.Background(), sessionId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SessionsAPI.AuthSvcV1MeSessionsSessionIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1MeSessionsSessionIdGet`: AccountSession
	fmt.Fprintf(os.Stdout, "Response from `SessionsAPI.AuthSvcV1MeSessionsSessionIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sessionId** | **uuid.UUID** | Session ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MeSessionsSessionIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**AccountSession**](AccountSession.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1RefreshPost

> TokensPair AuthSvcV1RefreshPost(ctx).RefreshSession(refreshSession).Execute()

Refresh session



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
	refreshSession := *openapiclient.NewRefreshSession(*openapiclient.NewRefreshSessionData("Type_example", *openapiclient.NewRefreshSessionDataAttributes("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."))) // RefreshSession | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SessionsAPI.AuthSvcV1RefreshPost(context.Background()).RefreshSession(refreshSession).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SessionsAPI.AuthSvcV1RefreshPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1RefreshPost`: TokensPair
	fmt.Fprintf(os.Stdout, "Response from `SessionsAPI.AuthSvcV1RefreshPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1RefreshPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **refreshSession** | [**RefreshSession**](RefreshSession.md) |  | 

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

