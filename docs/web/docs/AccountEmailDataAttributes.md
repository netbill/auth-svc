# AccountEmailDataAttributes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | **string** | The email address associated with the account | 
**Version** | **int32** | The version number of the account record | 
**Verified** | **bool** | Indicates whether the email address has been verified | 
**UpdatedAt** | **time.Time** | The date and time when the email information was last updated | 

## Methods

### NewAccountEmailDataAttributes

`func NewAccountEmailDataAttributes(email string, version int32, verified bool, updatedAt time.Time, ) *AccountEmailDataAttributes`

NewAccountEmailDataAttributes instantiates a new AccountEmailDataAttributes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccountEmailDataAttributesWithDefaults

`func NewAccountEmailDataAttributesWithDefaults() *AccountEmailDataAttributes`

NewAccountEmailDataAttributesWithDefaults instantiates a new AccountEmailDataAttributes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *AccountEmailDataAttributes) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *AccountEmailDataAttributes) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *AccountEmailDataAttributes) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetVersion

`func (o *AccountEmailDataAttributes) GetVersion() int32`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *AccountEmailDataAttributes) GetVersionOk() (*int32, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *AccountEmailDataAttributes) SetVersion(v int32)`

SetVersion sets Version field to given value.


### GetVerified

`func (o *AccountEmailDataAttributes) GetVerified() bool`

GetVerified returns the Verified field if non-nil, zero value otherwise.

### GetVerifiedOk

`func (o *AccountEmailDataAttributes) GetVerifiedOk() (*bool, bool)`

GetVerifiedOk returns a tuple with the Verified field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVerified

`func (o *AccountEmailDataAttributes) SetVerified(v bool)`

SetVerified sets Verified field to given value.


### GetUpdatedAt

`func (o *AccountEmailDataAttributes) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *AccountEmailDataAttributes) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *AccountEmailDataAttributes) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


