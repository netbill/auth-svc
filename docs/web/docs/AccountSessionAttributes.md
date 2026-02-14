# AccountSessionAttributes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccountId** | [**uuid.UUID**](uuid.UUID.md) | account id | 
**CreatedAt** | **time.Time** | session creation date | 
**Version** | **int32** | The version number of the account record | 
**LastUsed** | **time.Time** | last used date | 

## Methods

### NewAccountSessionAttributes

`func NewAccountSessionAttributes(accountId uuid.UUID, createdAt time.Time, version int32, lastUsed time.Time, ) *AccountSessionAttributes`

NewAccountSessionAttributes instantiates a new AccountSessionAttributes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccountSessionAttributesWithDefaults

`func NewAccountSessionAttributesWithDefaults() *AccountSessionAttributes`

NewAccountSessionAttributesWithDefaults instantiates a new AccountSessionAttributes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccountId

`func (o *AccountSessionAttributes) GetAccountId() uuid.UUID`

GetAccountId returns the AccountId field if non-nil, zero value otherwise.

### GetAccountIdOk

`func (o *AccountSessionAttributes) GetAccountIdOk() (*uuid.UUID, bool)`

GetAccountIdOk returns a tuple with the AccountId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccountId

`func (o *AccountSessionAttributes) SetAccountId(v uuid.UUID)`

SetAccountId sets AccountId field to given value.


### GetCreatedAt

`func (o *AccountSessionAttributes) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *AccountSessionAttributes) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *AccountSessionAttributes) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetVersion

`func (o *AccountSessionAttributes) GetVersion() int32`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *AccountSessionAttributes) GetVersionOk() (*int32, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *AccountSessionAttributes) SetVersion(v int32)`

SetVersion sets Version field to given value.


### GetLastUsed

`func (o *AccountSessionAttributes) GetLastUsed() time.Time`

GetLastUsed returns the LastUsed field if non-nil, zero value otherwise.

### GetLastUsedOk

`func (o *AccountSessionAttributes) GetLastUsedOk() (*time.Time, bool)`

GetLastUsedOk returns a tuple with the LastUsed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastUsed

`func (o *AccountSessionAttributes) SetLastUsed(v time.Time)`

SetLastUsed sets LastUsed field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


