package domain

import (
	"time"
)

type AttendanceEvent struct {
	ID               int       `json:"id"`
	Pin              string    `json:"pin"`
	Name             string    `json:"name"`
	LastName         string    `json:"last_name"`
	DeptCode         string    `json:"dept_code"`
	DeptName         string    `json:"dept_name"`
	ReaderName       string    `json:"reader_name"`
	DoorName         string    `json:"door_name"`
	DeviceSN         string    `json:"device_sn"`
	CapturePhotoPath string    `json:"capture_photo_path"`
	VerifyModeName   string    `json:"verify_mode_name"`
	EventName        string    `json:"event_name"`
	EventTime        time.Time `json:"event_time"`
	RawPayload       string    `json:"raw_payload"`
	UniqueKey        string    `json:"unique_key"`
	LogID            int       `json:"log_id"`
	CreatedAt        time.Time `json:"created_at"`
}

type ZKBioPushRequest struct {
	Module       string `json:"module"`
	DataType     string `json:"dataType"`
	ModuleName   string `json:"moduleName"`
	Content      string `json:"content"`
	PushType     string `json:"pushType"`
	PushTypeName string `json:"pushTypeName"`
}

type ZKBioEventContent struct {
	AreaID           string `json:"areaId"`
	AreaName         string `json:"areaName"`
	CapturePhotoPath string `json:"capturePhotoPath"`
	CardNo           string `json:"cardNo"`
	DeptCode         string `json:"deptCode"`
	DeptID           string `json:"deptId"`
	DeptName         string `json:"deptName"`
	Description      string `json:"description"`
	DevAlias         string `json:"devAlias"`
	DevID            string `json:"devId"`
	DevSN            string `json:"devSn"`
	DoorID           string `json:"doorId"`
	DoorName         string `json:"doorName"`
	Equals           bool   `json:"equals"`
	EventAddr        int    `json:"eventAddr"`
	EventLevel       int    `json:"eventLevel"`
	EventName        string `json:"eventName"`
	EventNo          int    `json:"eventNo"`
	EventPointID     string `json:"eventPointId"`
	EventPointName   string `json:"eventPointName"`
	EventPointType   int    `json:"eventPointType"`
	EventTime        int64  `json:"eventTime"`
	LastName         string `json:"lastName"`
	LogID            int    `json:"logId"`
	MaskFlag         string `json:"maskFlag"`
	Name             string `json:"name"`
	PersPersonID     string `json:"persPersonId"`
	Pin              string `json:"pin"`
	ReaderID         string `json:"readerId"`
	ReaderName       string `json:"readerName"`
	ReaderState      int    `json:"readerState"`
	Temperature      string `json:"temperature"`
	UniqueKey        string `json:"uniqueKey"`
	VerifyModeName   string `json:"verifyModeName"`
	VerifyModeNo     int    `json:"verifyModeNo"`
}

type Person struct {
	Pin       string
	IIN       string
	SchoolBIN string
	SUSN      bool
}
