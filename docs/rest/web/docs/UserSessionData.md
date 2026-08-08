# UserSessionData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | [**uuid.UUID**](uuid.UUID.md) | session id | 
**Type** | **string** |  | 
**Attributes** | [**UserSessionAttributes**](UserSessionAttributes.md) |  | 

## Methods

### NewUserSessionData

`func NewUserSessionData(id uuid.UUID, type_ string, attributes UserSessionAttributes, ) *UserSessionData`

NewUserSessionData instantiates a new UserSessionData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserSessionDataWithDefaults

`func NewUserSessionDataWithDefaults() *UserSessionData`

NewUserSessionDataWithDefaults instantiates a new UserSessionData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *UserSessionData) GetId() uuid.UUID`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *UserSessionData) GetIdOk() (*uuid.UUID, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *UserSessionData) SetId(v uuid.UUID)`

SetId sets Id field to given value.


### GetType

`func (o *UserSessionData) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *UserSessionData) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *UserSessionData) SetType(v string)`

SetType sets Type field to given value.


### GetAttributes

`func (o *UserSessionData) GetAttributes() UserSessionAttributes`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *UserSessionData) GetAttributesOk() (*UserSessionAttributes, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *UserSessionData) SetAttributes(v UserSessionAttributes)`

SetAttributes sets Attributes field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


