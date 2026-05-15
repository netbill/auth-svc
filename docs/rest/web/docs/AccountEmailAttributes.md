# AccountEmailAttributes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | **string** | The email address associated with the account | 
**Version** | **int32** | The version number of the account record | 
**Verified** | **bool** | Indicates whether the email address has been verified | 
**UpdatedAt** | **time.Time** | The date and time when the email information was last updated | 
**CreatedAt** | **time.Time** | The date and time when the email information was created | 

## Methods

### NewAccountEmailAttributes

`func NewAccountEmailAttributes(email string, version int32, verified bool, updatedAt time.Time, createdAt time.Time, ) *AccountEmailAttributes`

NewAccountEmailAttributes instantiates a new AccountEmailAttributes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccountEmailAttributesWithDefaults

`func NewAccountEmailAttributesWithDefaults() *AccountEmailAttributes`

NewAccountEmailAttributesWithDefaults instantiates a new AccountEmailAttributes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *AccountEmailAttributes) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *AccountEmailAttributes) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *AccountEmailAttributes) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetVersion

`func (o *AccountEmailAttributes) GetVersion() int32`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *AccountEmailAttributes) GetVersionOk() (*int32, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *AccountEmailAttributes) SetVersion(v int32)`

SetVersion sets Version field to given value.


### GetVerified

`func (o *AccountEmailAttributes) GetVerified() bool`

GetVerified returns the Verified field if non-nil, zero value otherwise.

### GetVerifiedOk

`func (o *AccountEmailAttributes) GetVerifiedOk() (*bool, bool)`

GetVerifiedOk returns a tuple with the Verified field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVerified

`func (o *AccountEmailAttributes) SetVerified(v bool)`

SetVerified sets Verified field to given value.


### GetUpdatedAt

`func (o *AccountEmailAttributes) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *AccountEmailAttributes) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *AccountEmailAttributes) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetCreatedAt

`func (o *AccountEmailAttributes) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *AccountEmailAttributes) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *AccountEmailAttributes) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


