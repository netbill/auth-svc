# AccountData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | [**uuid.UUID**](uuid.UUID.md) | account ID | 
**Type** | **string** |  | 
**Attributes** | [**AccountDataAttributes**](AccountDataAttributes.md) |  | 

## Methods

### NewAccountData

`func NewAccountData(id uuid.UUID, type_ string, attributes AccountDataAttributes, ) *AccountData`

NewAccountData instantiates a new AccountData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccountDataWithDefaults

`func NewAccountDataWithDefaults() *AccountData`

NewAccountDataWithDefaults instantiates a new AccountData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AccountData) GetId() uuid.UUID`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AccountData) GetIdOk() (*uuid.UUID, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AccountData) SetId(v uuid.UUID)`

SetId sets Id field to given value.


### GetType

`func (o *AccountData) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *AccountData) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *AccountData) SetType(v string)`

SetType sets Type field to given value.


### GetAttributes

`func (o *AccountData) GetAttributes() AccountDataAttributes`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *AccountData) GetAttributesOk() (*AccountDataAttributes, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *AccountData) SetAttributes(v AccountDataAttributes)`

SetAttributes sets Attributes field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


