# UploadUserMediaLinksData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | [**uuid.UUID**](uuid.UUID.md) | user id | 
**Type** | **string** |  | 
**Attributes** | [**UploadUserMediaLinksDataAttributes**](UploadUserMediaLinksDataAttributes.md) |  | 
**Relationships** | [**UploadUserMediaLinksDataRelationships**](UploadUserMediaLinksDataRelationships.md) |  | 

## Methods

### NewUploadUserMediaLinksData

`func NewUploadUserMediaLinksData(id uuid.UUID, type_ string, attributes UploadUserMediaLinksDataAttributes, relationships UploadUserMediaLinksDataRelationships, ) *UploadUserMediaLinksData`

NewUploadUserMediaLinksData instantiates a new UploadUserMediaLinksData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUploadUserMediaLinksDataWithDefaults

`func NewUploadUserMediaLinksDataWithDefaults() *UploadUserMediaLinksData`

NewUploadUserMediaLinksDataWithDefaults instantiates a new UploadUserMediaLinksData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *UploadUserMediaLinksData) GetId() uuid.UUID`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *UploadUserMediaLinksData) GetIdOk() (*uuid.UUID, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *UploadUserMediaLinksData) SetId(v uuid.UUID)`

SetId sets Id field to given value.


### GetType

`func (o *UploadUserMediaLinksData) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *UploadUserMediaLinksData) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *UploadUserMediaLinksData) SetType(v string)`

SetType sets Type field to given value.


### GetAttributes

`func (o *UploadUserMediaLinksData) GetAttributes() UploadUserMediaLinksDataAttributes`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *UploadUserMediaLinksData) GetAttributesOk() (*UploadUserMediaLinksDataAttributes, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *UploadUserMediaLinksData) SetAttributes(v UploadUserMediaLinksDataAttributes)`

SetAttributes sets Attributes field to given value.


### GetRelationships

`func (o *UploadUserMediaLinksData) GetRelationships() UploadUserMediaLinksDataRelationships`

GetRelationships returns the Relationships field if non-nil, zero value otherwise.

### GetRelationshipsOk

`func (o *UploadUserMediaLinksData) GetRelationshipsOk() (*UploadUserMediaLinksDataRelationships, bool)`

GetRelationshipsOk returns a tuple with the Relationships field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelationships

`func (o *UploadUserMediaLinksData) SetRelationships(v UploadUserMediaLinksDataRelationships)`

SetRelationships sets Relationships field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


