package config

import (
	"database/sql"
	"flag"
	"time"
)

type MailConfig struct {
	Id         int
	Idname     string
	Docsneeded string
	EmailFrom  string
	EmailTo    string
	Subject    string
	Greeting   string
	Body       string
	Closing    string
	Note       string
}
type Claim struct {
	LcId             int
	Claim_id         int
	Detail_line_id   int
	Bill_number      string
	Claim_status     string
	Analyst          string
	Contact_name     string
	Claimant_email   string
	UpdatedWhen      time.Time
	Claim_Date       time.Time
	RateClientId     string
	ClaimantID       string
	TraceNo          string
	AmtClamed        float64
	Company          string
	Address1         string
	Address2         string
	City             string
	Provence         string
	PostalCode       string
	Assigned_to      string
	Item_descr       string
	Item_is_required string
	POD_Signed       string
	POD_Date         time.Time
	DaysDiff         int
	Comments         string
	Note             string
	SalesRep         string
	Multiple         int
	FileError        bool
	IsValid          bool
}
type File struct {
	FBNumber    string
	BillToCode  string
	ClientId    string
	DocName     string
	DocLocation string
	DocId       string
	DocExt      string
	DocCreated  string
	DocTypeID   int
	DocType     string
}
type State struct {
	Claim_id         string
	Detail_line_id   string
	Bill_number      string
	Claim_status     string
	Email_from       string
	Email_to         string
	Analyst          string
	Contact_name     string
	Claimant_email   string
	Claim_Date       time.Time
	TraceNo          string
	AmtClamed        float64
	Company          string
	Address1         string
	Address2         string
	City             string
	Provence         string
	PostalCode       string
	Assigned_to      string
	Item_descr       string
	Item_is_required string
	POD_Signed       string
	POD_Date         time.Time
	DaysDiff         int
	Comments         string
	Note             string
	SalesRep         string
	Multiple         string
	State            string
}
type Config struct {
	MailPwd string `json:"MailPwd"`
	SynPwd  string `json:"SynPwd"`
	PdfPwd  string `json:"PdfPwd"`
}
type JobConfig struct {
	SynUid   string `json:"SynUid"`
	DB2dsn   string `json:"DB2dsn"`
	SynDbDsn string `json:"SynDbDsn"`
	HostMail string `json:"HostMail"`
	MailUid  string `json:"MailUid"`
	MailPort int    `json:"MailPort"`
	PdfUrl   string `json:"PdfUrl"`
	PdfUid   string `json:"PdfUid"`
}

var StartDate = flag.String("s", "", "start date")
var EndDate = flag.String("e", "", "end date")

var Clm = make([]*Claim, 0)
var Mc = make([]*MailConfig, 0)
var Files = make([]*File, 0)
var Conf Config
var JobConf JobConfig
var DocsNeeded = make([]string, 0)
var OFiles = make([]string, 0)
var MailStatus = make([]*State, 0)

var TmwDb *sql.DB
var SynDb *sql.DB
var SynTariffDb *sql.DB
