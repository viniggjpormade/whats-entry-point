package main

type WaReceivedResponse struct {
	Object string            `json:"object"`
	Entry  []WaReceivedEntry `json:"entry"`
}
type WaReceivedEntry struct {
	Changes []WaReceivedChanges `json:"changes"`
	Id      string              `json:"id"`
}
type WaReceivedChanges struct {
	Value WaReceivedValue `json:"value"`
	Field string          `json:"field"`
}
type WaReceivedMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberId      string `json:"phone_number_id"`
}
type WaReceivedStatus struct {
	Id          string         `json:"id"`
	Status      string         `json:"status"`
	Timestamp   string         `json:"timestamp"`
	RecipientId string         `json:"recipient_id"`
	Errors      []WaErrorField `json:"errors,omitempty"`
}
type WaErrorField struct {
	Code      int               `json:"code"`
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	ErrorData *WaErrorDataField `json:"error_data,omitempty"`
}
type WaErrorDataField struct {
	Details string `json:"details"`
}
type WaMediaField struct {
	Id       string `json:"id"`
	MimeType string `json:"mime_type"`
	Caption  string `json:"caption,omitempty"`
}
type WaReceivedTextField struct {
	Body string `json:"body"`
}
type WaReceivedProfileField struct {
	Name string `json:"name"`
}
type WaReceivedMessage struct {
	From        string               `json:"from"`
	FromUserId  string               `json:"from_user_id"`
	Id          string               `json:"id"`
	TimeStamp   string               `json:"timestamp"`
	MessageType string               `json:"type"`
	Text        *WaReceivedTextField `json:"text,omitempty"`
	Image       *WaMediaField        `json:"image,omitempty"`
	Audio       *WaMediaField        `json:"audio,omitempty"`
	Video       *WaMediaField        `json:"video,omitempty"`
	Document    *WaMediaField        `json:"document,omitempty"`
}
type WaReceivedContact struct {
	Profile WaReceivedProfileField `json:"profile"`
	UserId  string                 `json:"user_id"`
	WaId    string                 `json:"wa_id"`
}

type WaReceivedValue struct {
	Contacts         []WaReceivedContact `json:"contacts"`
	Messages         []WaReceivedMessage `json:"messages"`
	Statuses         []WaReceivedStatus  `json:"statuses"`
	MessagingProduct string              `json:"messaging_product"`
	Metadata         WaReceivedMetadata  `json:"metadata"`
}
