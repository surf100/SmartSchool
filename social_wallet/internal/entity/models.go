package entity

import "time"

type AccessMapping struct {
	ReaderName string
	Pin        string
}

type Student struct {
	Pin       string
	IIN       string
	SchoolBin string
	SetSocPay bool
}

type AccessEvent struct {
	Pin        string
	ReaderName string
	EventTime  time.Time
	RawJSON    string
}
type TerminalPayload struct {
	Module     string `json:"module"`
	DataType   string `json:"dataType"`
	ModuleName string `json:"moduleName"`
	Data       string `json:"content"`
	PushType   string `json:"pushType,omitempty"`
}

type TransactionRequest struct {
	IIN         string `json:"iin" validate:"required,len=12"`
	Date        string `json:"date" validate:"required,datetime=2006-01-02 15:04:05"`
	SchoolBIN   string `json:"school_bin" validate:"required,len=12"`
	SetSocPay   string `json:"set_socpay"`
	ResetSocPay string `json:"reset_socpay"`
}

type Person struct {
	Pin               string `json:"pin" bson:"pin" db:"pin"`
	DeptName          string `json:"deptname" bson:"deptname" db:"deptname"`
	Name              string `json:"name" bson:"name" db:"name"`
	LastName          string `json:"lastname" bson:"lastname" db:"lastname"`
	Gender            string `json:"gender" bson:"gender" db:"gender"`
	Birthday          string `json:"birthday" bson:"birthday" db:"birthday"`
	MobilePhone       string `json:"mobilephone" bson:"mobilephone" db:"mobilephone"`
	Email             string `json:"email" bson:"email" db:"email"`
	CertType          string `json:"certtype" bson:"certtype" db:"certtype"`
	CertNumber        string `json:"certnumber" bson:"certnumber" db:"certnumber"`
	PhotoPath         string `json:"photopath" bson:"photopath" db:"photopath"`
	VislightPhotoPath string `json:"vislightphotopath" bson:"vislightphotopath" db:"vislightphotopath"`
	AccLevelIds       string `json:"acclevelids" bson:"acclevelids" db:"acclevelids"`
	ParentIin         string `json:"parentiin" bson:"parentiin" db:"parentiin"`
	InSchool          bool   `json:"inschool" bson:"inschool" db:"inschool"`
	Susn              bool   `json:"susn" bson:"susn" db:"susn"`
	SetSocPay         bool   `json:"set_socpay" bson:"set_socpay" db:"set_socpay"` 
}
