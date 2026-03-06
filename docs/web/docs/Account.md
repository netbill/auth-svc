# Account

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Data** | [**AccountData**](AccountData.md) |  | 
**Included** | Pointer to [**[]AccountEmail**](AccountEmail.md) |  | [optional] 

## Methods

### NewAccount

`func NewAccount(data AccountData, ) *Account`

NewAccount instantiates a new Account object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccountWithDefaults

`func NewAccountWithDefaults() *Account`

NewAccountWithDefaults instantiates a new Account object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetData

`func (o *Account) GetData() AccountData`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *Account) GetDataOk() (*AccountData, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *Account) SetData(v AccountData)`

SetData sets Data field to given value.


### GetIncluded

`func (o *Account) GetIncluded() []AccountEmail`

GetIncluded returns the Included field if non-nil, zero value otherwise.

### GetIncludedOk

`func (o *Account) GetIncludedOk() (*[]AccountEmail, bool)`

GetIncludedOk returns a tuple with the Included field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIncluded

`func (o *Account) SetIncluded(v []AccountEmail)`

SetIncluded sets Included field to given value.

### HasIncluded

`func (o *Account) HasIncluded() bool`

HasIncluded returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


