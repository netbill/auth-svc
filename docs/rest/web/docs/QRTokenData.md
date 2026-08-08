# QRTokenData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** |  | 
**Attributes** | [**QRTokenDataAttributes**](QRTokenDataAttributes.md) |  | 

## Methods

### NewQRTokenData

`func NewQRTokenData(type_ string, attributes QRTokenDataAttributes, ) *QRTokenData`

NewQRTokenData instantiates a new QRTokenData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewQRTokenDataWithDefaults

`func NewQRTokenDataWithDefaults() *QRTokenData`

NewQRTokenDataWithDefaults instantiates a new QRTokenData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *QRTokenData) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *QRTokenData) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *QRTokenData) SetType(v string)`

SetType sets Type field to given value.


### GetAttributes

`func (o *QRTokenData) GetAttributes() QRTokenDataAttributes`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *QRTokenData) GetAttributesOk() (*QRTokenDataAttributes, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *QRTokenData) SetAttributes(v QRTokenDataAttributes)`

SetAttributes sets Attributes field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


