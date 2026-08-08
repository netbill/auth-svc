# QRTokenDataAttributes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**QrToken** | [**uuid.UUID**](uuid.UUID.md) | Token to render as a QR code. Send it back via POST /auth-svc/v1/login/qr/confirm to complete the login.  | 

## Methods

### NewQRTokenDataAttributes

`func NewQRTokenDataAttributes(qrToken uuid.UUID, ) *QRTokenDataAttributes`

NewQRTokenDataAttributes instantiates a new QRTokenDataAttributes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewQRTokenDataAttributesWithDefaults

`func NewQRTokenDataAttributesWithDefaults() *QRTokenDataAttributes`

NewQRTokenDataAttributesWithDefaults instantiates a new QRTokenDataAttributes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetQrToken

`func (o *QRTokenDataAttributes) GetQrToken() uuid.UUID`

GetQrToken returns the QrToken field if non-nil, zero value otherwise.

### GetQrTokenOk

`func (o *QRTokenDataAttributes) GetQrTokenOk() (*uuid.UUID, bool)`

GetQrTokenOk returns a tuple with the QrToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQrToken

`func (o *QRTokenDataAttributes) SetQrToken(v uuid.UUID)`

SetQrToken sets QrToken field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


