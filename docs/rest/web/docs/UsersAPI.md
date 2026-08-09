# \UsersAPI

All URIs are relative to *http://localhost:8001*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AuthSvcV1MeDelete**](UsersAPI.md#AuthSvcV1MeDelete) | **Delete** /auth-svc/v1/me | Delete my user
[**AuthSvcV1MeGet**](UsersAPI.md#AuthSvcV1MeGet) | **Get** /auth-svc/v1/me | Get my user
[**AuthSvcV1MeMediaDelete**](UsersAPI.md#AuthSvcV1MeMediaDelete) | **Delete** /auth-svc/v1/me/media | Delete uploaded user media
[**AuthSvcV1MeMediaPost**](UsersAPI.md#AuthSvcV1MeMediaPost) | **Post** /auth-svc/v1/me/media | Create user avatar upload media link
[**AuthSvcV1MePasswordPatch**](UsersAPI.md#AuthSvcV1MePasswordPatch) | **Patch** /auth-svc/v1/me/password | Update password
[**AuthSvcV1MePatch**](UsersAPI.md#AuthSvcV1MePatch) | **Patch** /auth-svc/v1/me | Update my user
[**AuthSvcV1MeUsernamePatch**](UsersAPI.md#AuthSvcV1MeUsernamePatch) | **Patch** /auth-svc/v1/me/username | Update my username
[**AuthSvcV1UsersGet**](UsersAPI.md#AuthSvcV1UsersGet) | **Get** /auth-svc/v1/users/ | Filter users
[**AuthSvcV1UsersUserIdGet**](UsersAPI.md#AuthSvcV1UsersUserIdGet) | **Get** /auth-svc/v1/users/{user_id} | Get user by id
[**AuthSvcV1UsersUsernameGet**](UsersAPI.md#AuthSvcV1UsersUsernameGet) | **Get** /auth-svc/v1/users/@{username} | Get user by username



## AuthSvcV1MeDelete

> AuthSvcV1MeDelete(ctx).Execute()

Delete my user



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
	r, err := apiClient.UsersAPI.AuthSvcV1MeDelete(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.AuthSvcV1MeDelete``: %v\n", err)
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

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1MeGet

> User AuthSvcV1MeGet(ctx).Include(include).Execute()

Get my user



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
	include := []string{"Include_example"} // []string | Optional related resources to include. Supported values: `email`.  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.AuthSvcV1MeGet(context.Background()).Include(include).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.AuthSvcV1MeGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1MeGet`: User
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.AuthSvcV1MeGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MeGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **include** | **[]string** | Optional related resources to include. Supported values: &#x60;email&#x60;.  | 

### Return type

[**User**](User.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1MeMediaDelete

> AuthSvcV1MeMediaDelete(ctx).DeleteUploadUserAvatar(deleteUploadUserAvatar).Execute()

Delete uploaded user media



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
	deleteUploadUserAvatar := *openapiclient.NewDeleteUploadUserAvatar(*openapiclient.NewDeleteUploadUserAvatarData("TODO", "Type_example", *openapiclient.NewDeleteUploadUserAvatarDataAttributes())) // DeleteUploadUserAvatar | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UsersAPI.AuthSvcV1MeMediaDelete(context.Background()).DeleteUploadUserAvatar(deleteUploadUserAvatar).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.AuthSvcV1MeMediaDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MeMediaDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deleteUploadUserAvatar** | [**DeleteUploadUserAvatar**](DeleteUploadUserAvatar.md) |  | 

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


## AuthSvcV1MeMediaPost

> UploadUserMediaLinks AuthSvcV1MeMediaPost(ctx).Execute()

Create user avatar upload media link



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
	resp, r, err := apiClient.UsersAPI.AuthSvcV1MeMediaPost(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.AuthSvcV1MeMediaPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1MeMediaPost`: UploadUserMediaLinks
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.AuthSvcV1MeMediaPost`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MeMediaPostRequest struct via the builder pattern


### Return type

[**UploadUserMediaLinks**](UploadUserMediaLinks.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

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
	r, err := apiClient.UsersAPI.AuthSvcV1MePasswordPatch(context.Background()).UpdatePassword(updatePassword).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.AuthSvcV1MePasswordPatch``: %v\n", err)
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

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1MePatch

> User AuthSvcV1MePatch(ctx).UpdateUser(updateUser).Execute()

Update my user



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
	updateUser := *openapiclient.NewUpdateUser(*openapiclient.NewUpdateUserData("TODO", "Type_example", *openapiclient.NewUpdateUserDataAttributes())) // UpdateUser | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.AuthSvcV1MePatch(context.Background()).UpdateUser(updateUser).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.AuthSvcV1MePatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1MePatch`: User
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.AuthSvcV1MePatch`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MePatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updateUser** | [**UpdateUser**](UpdateUser.md) |  | 

### Return type

[**User**](User.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1MeUsernamePatch

> User AuthSvcV1MeUsernamePatch(ctx).UpdateUsername(updateUsername).Execute()

Update my username



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
	updateUsername := *openapiclient.NewUpdateUsername(*openapiclient.NewUpdateUsernameData("TODO", "Type_example", *openapiclient.NewUpdateUsernameDataAttributes("Username_example"))) // UpdateUsername | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.AuthSvcV1MeUsernamePatch(context.Background()).UpdateUsername(updateUsername).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.AuthSvcV1MeUsernamePatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1MeUsernamePatch`: User
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.AuthSvcV1MeUsernamePatch`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1MeUsernamePatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updateUsername** | [**UpdateUsername**](UpdateUsername.md) |  | 

### Return type

[**User**](User.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1UsersGet

> UsersCollection AuthSvcV1UsersGet(ctx).Text(text).Page(page).Size(size).Execute()

Filter users



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
	text := "text_example" // string | Text to filter users by. Matches against `username` and `pseudonym` fields.  (optional)
	page := int32(56) // int32 | Page number (1-based). (optional)
	size := int32(56) // int32 | Max number of items per page (1-100). (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.AuthSvcV1UsersGet(context.Background()).Text(text).Page(page).Size(size).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.AuthSvcV1UsersGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1UsersGet`: UsersCollection
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.AuthSvcV1UsersGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1UsersGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **text** | **string** | Text to filter users by. Matches against &#x60;username&#x60; and &#x60;pseudonym&#x60; fields.  | 
 **page** | **int32** | Page number (1-based). | 
 **size** | **int32** | Max number of items per page (1-100). | 

### Return type

[**UsersCollection**](UsersCollection.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1UsersUserIdGet

> User AuthSvcV1UsersUserIdGet(ctx, userId).Execute()

Get user by id



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
	userId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // uuid.UUID | User id (UUID).

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.AuthSvcV1UsersUserIdGet(context.Background(), userId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.AuthSvcV1UsersUserIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1UsersUserIdGet`: User
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.AuthSvcV1UsersUserIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **uuid.UUID** | User id (UUID). | 

### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1UsersUserIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**User**](User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthSvcV1UsersUsernameGet

> User AuthSvcV1UsersUsernameGet(ctx, username).Execute()

Get user by username



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
	username := "username_example" // string | Username of the user.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.AuthSvcV1UsersUsernameGet(context.Background(), username).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.AuthSvcV1UsersUsernameGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthSvcV1UsersUsernameGet`: User
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.AuthSvcV1UsersUsernameGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**username** | **string** | Username of the user. | 

### Other Parameters

Other parameters are passed through a pointer to a apiAuthSvcV1UsersUsernameGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**User**](User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

