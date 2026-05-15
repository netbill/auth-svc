# PaginationData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Self** | **string** | link to the current page | 
**First** | Pointer to **string** | link to the first page | [optional] 
**Last** | Pointer to **string** | link to the last page | [optional] 
**Prev** | Pointer to **string** | link to the previous page | [optional] 
**Next** | Pointer to **string** | link to the next page | [optional] 

## Methods

### NewPaginationData

`func NewPaginationData(self string, ) *PaginationData`

NewPaginationData instantiates a new PaginationData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPaginationDataWithDefaults

`func NewPaginationDataWithDefaults() *PaginationData`

NewPaginationDataWithDefaults instantiates a new PaginationData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSelf

`func (o *PaginationData) GetSelf() string`

GetSelf returns the Self field if non-nil, zero value otherwise.

### GetSelfOk

`func (o *PaginationData) GetSelfOk() (*string, bool)`

GetSelfOk returns a tuple with the Self field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelf

`func (o *PaginationData) SetSelf(v string)`

SetSelf sets Self field to given value.


### GetFirst

`func (o *PaginationData) GetFirst() string`

GetFirst returns the First field if non-nil, zero value otherwise.

### GetFirstOk

`func (o *PaginationData) GetFirstOk() (*string, bool)`

GetFirstOk returns a tuple with the First field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFirst

`func (o *PaginationData) SetFirst(v string)`

SetFirst sets First field to given value.

### HasFirst

`func (o *PaginationData) HasFirst() bool`

HasFirst returns a boolean if a field has been set.

### GetLast

`func (o *PaginationData) GetLast() string`

GetLast returns the Last field if non-nil, zero value otherwise.

### GetLastOk

`func (o *PaginationData) GetLastOk() (*string, bool)`

GetLastOk returns a tuple with the Last field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLast

`func (o *PaginationData) SetLast(v string)`

SetLast sets Last field to given value.

### HasLast

`func (o *PaginationData) HasLast() bool`

HasLast returns a boolean if a field has been set.

### GetPrev

`func (o *PaginationData) GetPrev() string`

GetPrev returns the Prev field if non-nil, zero value otherwise.

### GetPrevOk

`func (o *PaginationData) GetPrevOk() (*string, bool)`

GetPrevOk returns a tuple with the Prev field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrev

`func (o *PaginationData) SetPrev(v string)`

SetPrev sets Prev field to given value.

### HasPrev

`func (o *PaginationData) HasPrev() bool`

HasPrev returns a boolean if a field has been set.

### GetNext

`func (o *PaginationData) GetNext() string`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *PaginationData) GetNextOk() (*string, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *PaginationData) SetNext(v string)`

SetNext sets Next field to given value.

### HasNext

`func (o *PaginationData) HasNext() bool`

HasNext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


