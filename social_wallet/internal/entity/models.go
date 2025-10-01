package entity

import "time"

type Person struct {
	Pin        string `json:"pin" db:"pin"`
	IIN        string `json:"iin" db:"iin"`
	SchoolBin  string `json:"school_bin" db:"school_bin"`
	Susn       bool   `json:"susn" db:"susn"`
	CardNumber string `json:"card_number" db:"card_number"`
}

type ExternalSusnData struct {
	ID            int
	IIN           string
	SchoolBin     string
	SocialPayment bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ZKBioEvent struct {
	Pin              string `json:"pin"`
	Name             string `json:"name"`
	LastName         string `json:"lastName"`
	DeptCode         string `json:"deptCode"`
	DeptName         string `json:"deptName"`
	ReaderName       string `json:"readerName"`
	DoorName         string `json:"doorName"`
	DevSn            string `json:"devSn"`
	CapturePhotoPath string `json:"capturePhotoPath"`
	VerifyModeName   string `json:"verifyModeName"`
	EventName        string `json:"eventName"`
	EventTime        int64  `json:"eventTime"` 
}

type ZKBioPushRequest struct {
	Module       string `json:"module"`
	DataType     string `json:"dataType"`
	ModuleName   string `json:"moduleName"`
	Content      string `json:"content"` 
	PushType     string `json:"pushType"`
	PushTypeName string `json:"pushTypeName"`
}
