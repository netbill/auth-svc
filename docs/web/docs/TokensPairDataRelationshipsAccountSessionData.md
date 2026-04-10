# TokensPairDataRelationshipsAccountSessionData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | [**uuid.UUID**](uuid.UUID.md) | account session id | 
**Type** | **string** |  | 

## Methods

### NewTokensPairDataRelationshipsAccountSessionData

`func NewTokensPairDataRelationshipsAccountSessionData(id uuid.UUID, type_ string, ) *TokensPairDataRelationshipsAccountSessionData`

NewTokensPairDataRelationshipsAccountSessionData instantiates a new TokensPairDataRelationshipsAccountSessionData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTokensPairDataRelationshipsAccountSessionDataWithDefaults

`func NewTokensPairDataRelationshipsAccountSessionDataWithDefaults() *TokensPairDataRelationshipsAccountSessionData`

NewTokensPairDataRelationshipsAccountSessionDataWithDefaults instantiates a new TokensPairDataRelationshipsAccountSessionData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *TokensPairDataRelationshipsAccountSessionData) GetId() uuid.UUID`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *TokensPairDataRelationshipsAccountSessionData) GetIdOk() (*uuid.UUID, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *TokensPairDataRelationshipsAccountSessionData) SetId(v uuid.UUID)`

SetId sets Id field to given value.


### GetType

`func (o *TokensPairDataRelationshipsAccountSessionData) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *TokensPairDataRelationshipsAccountSessionData) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *TokensPairDataRelationshipsAccountSessionData) SetType(v string)`

SetType sets Type field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


