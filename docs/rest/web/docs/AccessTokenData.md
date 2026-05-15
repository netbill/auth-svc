# AccessTokenData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** |  | 
**Attributes** | [**AccessTokenDataAttributes**](AccessTokenDataAttributes.md) |  | 

## Methods

### NewAccessTokenData

`func NewAccessTokenData(type_ string, attributes AccessTokenDataAttributes, ) *AccessTokenData`

NewAccessTokenData instantiates a new AccessTokenData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessTokenDataWithDefaults

`func NewAccessTokenDataWithDefaults() *AccessTokenData`

NewAccessTokenDataWithDefaults instantiates a new AccessTokenData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *AccessTokenData) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *AccessTokenData) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *AccessTokenData) SetType(v string)`

SetType sets Type field to given value.


### GetAttributes

`func (o *AccessTokenData) GetAttributes() AccessTokenDataAttributes`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *AccessTokenData) GetAttributesOk() (*AccessTokenDataAttributes, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *AccessTokenData) SetAttributes(v AccessTokenDataAttributes)`

SetAttributes sets Attributes field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


