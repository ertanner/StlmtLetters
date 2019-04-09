package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/alexbrainman/odbc"
	"github.com/vjeantet/jodaTime"
	"gopkg.in/gomail.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//Claims_Review@Dylt.com
//TST_Recipient@DYLT.com
//TST_CC@DYLT.com
//TST_BCC@DYLT.com
//TST_Review@DYLT.com

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
	SSPwd   string `json:"SSPwd"`
}

var err error
var Conf Config
var Clm = make([]*Claim, 0)
var Mc = make([]*MailConfig, 0)
var Files = make([]*File, 0)
var LocalFile = make([]string, 0)
var OFiles = make([]string, 0)
var DocsNeeded = make([]string, 0)
var CommentsNeeded = make([]string, 0)
var MailStatus = make([]*State, 0)

var buf bytes.Buffer
var logger = log.New(&buf, "logger: ", log.Lshortfile)

var startDate = flag.String("s", "", "start date")
var endDate = flag.String("e", "", "end date")

//var db = flag.String("db", "", "DYLT_REP")

const layoutUS = "01-02-2006"

var TmwDb *sql.DB
var SynDb *sql.DB
var SynTariffDb *sql.DB

func init() {
	//*************************************************************************
	//TmwDb
	//*************************************************************************
	TmwDb, err = sql.Open("odbc", "DSN=DYLT_PRJ")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connecting to TMW database...")

	TmwDb.SetMaxOpenConns(10)
	//db.SetMaxIdleConns(3)
	TmwDb.SetConnMaxLifetime(30 * time.Minute)
	TmwDb.SetMaxIdleConns(0)

	err = TmwDb.Ping()
	if err != nil {
		log.Println("Ping error")
	} else {
		log.Println("Database connection to TMW is opened...")
	}

	//*************************************************************************
	// connect to the Synergize db
	//*************************************************************************
	SynDb, err = sql.Open("odbc",
		"DSN=SYNERGIZE;UID=sa;PWD=Syn3rg1ze")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connecting to the Synergize database...")
	SynDb.SetMaxOpenConns(10)
	//db.SetMaxIdleConns(3)
	SynDb.SetConnMaxLifetime(30 * time.Minute)
	SynDb.SetMaxIdleConns(0)

	err = SynDb.Ping()
	if err != nil {
		log.Println("Ping error ", err)
	} else {
		log.Println("The Synergize database connection is opened...")
	}
	//*************************************************************************
	//svc_sys_dash  or sa with pwd Syn3rg1ze
	// connect to the Synergize db
	//*************************************************************************
	SynTariffDb, err = sql.Open("odbc",
		"DSN=SYNERGIZE;UID=sa;PWD=Syn3rg1ze")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connecting to the Synergize database...")
	SynTariffDb.SetMaxOpenConns(10)
	//db.SetMaxIdleConns(3)
	SynTariffDb.SetConnMaxLifetime(30 * time.Minute)
	SynTariffDb.SetMaxIdleConns(0)

	err = SynTariffDb.Ping()
	if err != nil {
		log.Println("Ping error ", err)
	} else {
		log.Println("The Synergize database connection is opened...")
	}

}

func main() {
	start := time.Now()
	fmt.Println("Start Date: " + start.String())

	// this will clear the temp image files from the tmp_images directory
	clearTmpFiles()

	//create your file with desired read/write permissions
	f, err := os.OpenFile("log_file", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println(err)
	}
	//defer to close when you're done with it
	defer f.Close()
	//set output of logs to f  this points the "log" commad to output to the f which is defined above
	log.SetOutput(f)
	log.Println("Starting mail program.")

	// parse the parameters and if empty set the dates
	flag.Parse()

	log.Println("Start Flag Date: " + *startDate)
	log.Println("End Flag Date: " + *endDate)
	if *startDate == "" {
		*startDate = time.Now().Format(layoutUS)
		//*startDate = "3/11/2019"
		log.Println("Start Date from flag " + *startDate)
	} else {
		log.Println("Start Date from flag e " + *startDate)
	}
	if *endDate == "" {
		now := int(time.Now().Weekday())
		lastSun := int(time.Sunday)
		dateDiff := now - lastSun
		addDate := time.Now().AddDate(0, 0, -dateDiff)
		*endDate = addDate.Format(layoutUS)
		log.Println("End Date from flag " + *endDate)
	} else {
		log.Println("End Date from flag e " + *endDate)
	}

	//*********************************************
	// Test code here.  remove it for prod
	//*startDate = "3/11/2019"
	//*endDate = "4/07/2019"
	//*********************************************

	// get the files off the O drive
	copyODrive()
	log.Println("copyODrive")

	// get base config for user id and pwd
	Conf = LoadConfiguration("acxpuwd")

	// connect to the db to get the DB config showing the checklist items to be processed
	GetConfig()
	log.Println("Got DB config")
	if err != nil {
		panic(err)
	}

	// ***************************************************************
	// Iterate over the list of checklist types
	// ***************************************************************
	for i := 0; i < len(Mc); i++ {
		log.Println("  --------------------------------------------  ")
		log.Println("  MailConfig")
		log.Println("  --------------------------------------------  ")
		log.Println("  --------------------------------------------  ")

		// Parse out the file type.  It is stored as a comma delited string in the db.
		DocsNeeded = DocsNeeded[:0]
		if Mc[i].Docsneeded != "" {
			DocsNeeded = parseString(Mc[i].Docsneeded)
			log.Println("Docs Needed line 1")
			log.Println(DocsNeeded)
		}
		log.Println("parseString")

		//GetOFiles(i)
		//log.Println("Got to GetOFiles")

		// get a list of claims for the checklist item type.
		GetClaims(strconv.Itoa(Mc[i].Id))
		log.Println("Got Claims")

		// ***************************************************************
		//Iterate over the list of claims and process them
		// ***************************************************************
		for j := 0; j < len(Clm); j++ {
			log.Println("  --------------------------------------------  ")
			log.Println(" Claim ID " + strconv.Itoa(Clm[j].Claim_id))
			log.Println("  --------------------------------------------  ")
			log.Println("  --------------------------------------------  ")
			LocalFile = LocalFile[:0]
			Files = Files[:0]
			CommentsNeeded = CommentsNeeded[:0]
			isValid := true
			log.Println("Setting isValid = true")
			// Get all the files needed to attach to the email.
			// return an isValid = false if any file is not present
			GetFiles(i, j)
			log.Println("isValid state after GetFiles: " + strconv.FormatBool(Clm[j].FileError))

			// calc days diff.
			if Mc[i].Id == 101 {
				log.Println("Start Date: " + Clm[j].POD_Date.String() + "  End Date: " + Clm[j].Claim_Date.String())
				diffDay := businessDays(Clm[j].POD_Date, Clm[j].Claim_Date)
				log.Println(" Days difference: " + strconv.Itoa(diffDay))
			}

			// get notes
			log.Println("Get Notes")
			if Mc[i].Id == 101 || Mc[i].Id == 146 || Mc[i].Id == 99 {
				Clm[j].Note = formatNote(j, Mc[i].Id)
				log.Println("101 or 146 or 99 - Parsed Email Config")
				//log.Println("J is: %d", j)
			} else {
				gotNote := GetNote(strconv.Itoa(Clm[j].Detail_line_id), j)
				if !isValid && !gotNote {
					log.Println("gotNote is false as well as inValid")
				} else if isValid && !gotNote {
					log.Println("gotNote is false.  Setting isValid to false")
					log.Println("FileError is: " + strconv.FormatBool(Clm[j].FileError))
					isValid = false
				} else {
					log.Println("Used Notes")
					//log.Println(Clm[j].Note)
				}
			}

			// log the claim
			log.Println(Clm[j].LcId, strconv.Itoa(Mc[i].Id), Mc[i].EmailFrom, Mc[i].EmailTo, strconv.Itoa(Clm[j].Detail_line_id), strconv.Itoa(Clm[j].Claim_id),
				Clm[j].Bill_number, Clm[j].Contact_name, Clm[j].Assigned_to, Clm[j].Analyst, Clm[j].Claim_Date, Clm[j].TraceNo, Clm[j].AmtClamed,
				Clm[j].Address1, Clm[j].Address2, Clm[j].City, Clm[j].Provence, Clm[j].PostalCode,
				Clm[j].Note, Clm[j].POD_Signed, Clm[j].POD_Date, Clm[j].DaysDiff, strconv.Itoa(Clm[j].Multiple))
			clm := Clm[j]
			mcon := Mc[i]

			// check the fields to be sure all valid data is there.  If nto then send it to the reviewClaims inBox
			if isValid {
				isValid = validateFields(*clm, *mcon)
			}
			log.Println("isValid status from validateFields = " + strconv.FormatBool(isValid))
			log.Println("Claim file error status = " + strconv.FormatBool(Clm[j].FileError))
			if Clm[j].FileError {
				isValid = false
			}
			// send email
			resultErr := sendEmail(strconv.Itoa(Mc[i].Id), Mc[i].EmailFrom, Mc[i].EmailTo, j, isValid)
			if resultErr {
				log.Println("ERROR with send email.  " + strconv.Itoa(Clm[j].Claim_id) + Clm[j].Bill_number)
			}

		}
		// reset Claim slice back to 0 elements
		Clm = Clm[:0]
	}

	// clean up tmp files
	sendMailStatus()
	clearTmpFiles()
	// ***************************************************************
	// done with process.
	// ***************************************************************
}
func GetCustTariff(i, j int) bool {
	log.Println("GetCustTariff")

	// generate the string of comma delimited values to put into the frtBill values
	stmt := `select top 1 In_DocID, coalesce(In_DocName, 'nil')as DocName, In_DocTypeID, In_DocFileExt, In_DocLocation, ClientID, In_DocCreated  
				from CONTRACTDOCS.dbo.Main
				where ClientID = ` + Clm[j].ClaimantID + `
				and In_DocTypeID = 13
				and In_DocCreated < '` + Clm[j].Claim_Date.Format(layoutUS) + `'`

	log.Println("Syn Cust Tariff Stmt = " + stmt)
	rows, err := SynTariffDb.Query(stmt)
	fileFound := false
	if err != nil {
		log.Println(err)
	} else {
		defer rows.Close()

		for rows.Next() {
			f := new(File)
			err = rows.Scan(&f.DocId, &f.DocName, &f.DocTypeID, &f.DocExt, &f.DocLocation, &f.ClientId, &f.DocCreated)
			if err != nil {
				log.Println(err)
			}
			Files = append(Files, f)
			fileName := f.DocId + "." + f.DocExt
			log.Println("FileNameTariff = " + fileName)
			if fileName != "" {
				// strings.TrimLeft( f.DocLocation,"1\\CONTRACTDOCS")
				str := strings.Replace(f.DocLocation, "1\\CONTRACTDOCS", "", -1)
				log.Println("Trimmed string " + string(str))
				log.Println("Y:" + string(str) + "\\" + fileName)
				copyFile("Y:"+string(str)+"\\"+fileName, ".\\tmp_Images\\"+fileName)
				LocalFile = append(LocalFile, ".\\tmp_Images\\"+fileName)
				if !Clm[j].FileError {
					fileFound = true
				}
			} else {
				log.Println("fileName is empty")
				fileFound = false
				Clm[j].FileError = true
			}
		}

	}
	return fileFound
}
func GetFiles(i, j int) {
	// Get Files fucntion

	filesNeeded := make([]string, 0)

	// Attach the files from the comments field if they are present.
	// These are NMFC items and were donwloaded with the O drive
	// if the comments field is not null then get the additonal files
	if Mc[i].Id == 169 && Clm[j].Comments != "" {
		additionFiles := parseString(Clm[j].Comments)
		for z := 0; z < len(additionFiles); z++ {
			log.Println("NMFC Item " + additionFiles[z] + ".pdf")
			LocalFile = append(LocalFile, ".\\tmp_images\\NMFC Item "+additionFiles[z]+".pdf")
			filesNeeded = append(filesNeeded, "nmfc")
		}
	}

	// get 110Tarrif and NMFTA if needed
	if len(DocsNeeded) > 0 {
		for k := 0; k < len(DocsNeeded); k++ {
			if DocsNeeded[k] == "110Tariff" {
				log.Println("Gettign 110Tariff")
				LocalFile = append(LocalFile, ".\\tmp_images\\110Tariff.pdf")
				filesNeeded = append(filesNeeded, "110Tariff")
			}
			if DocsNeeded[k] == "nmfta" {
				log.Println("Getting NMFTA files")
				LocalFile = append(LocalFile, ".\\tmp_images\\NMFTA_Item_300105.pdf")
				LocalFile = append(LocalFile, ".\\tmp_images\\NMFTA_Item_300135.pdf")
				filesNeeded = append(filesNeeded, "nmfta")
			}
		}
	}

	//Get the tif from the Synergize
	fileToGet := make([]string, 0)
	for k := 0; k < len(DocsNeeded); k++ {
		log.Println(DocsNeeded[k])
		if DocsNeeded[k] == "bol" || DocsNeeded[k] == "dr" {
			fileToGet = append(fileToGet, DocsNeeded[k])
		}
		if Clm[j].ClaimantID == Clm[j].RateClientId {
			if DocsNeeded[k] == "custTariff" {
				isValid := GetCustTariff(i, j)
				if isValid {
					log.Println("found custTariff")
					filesNeeded = append(filesNeeded, "custTariff")
				} else {
					log.Println("could not find cust tariff")
					Clm[j].FileError = true
				}
			}
		}
	}
	if len(fileToGet) > 0 {
		fileStr := ""
		for l := 0; l < len(fileToGet); l++ {
			if fileToGet[l] == "bol" {
				fileStr = fileStr + "16, "
				log.Println("fileStr1" + fileStr)
			}
			if fileToGet[l] == "dr" {
				fileStr = fileStr + "25, "
				log.Println("fileStr2" + fileStr)
			}
		}
		if fileStr != "" {
			fileStr = strings.TrimRight(fileStr, ", ")
			// generate the string of comma delimited values to put into the frtBill values
			stmt := `with cte as (
				select c.FBNumber, c.Bill_To_Code, c.In_DocFamilyID, m.In_DocLocation, m.In_DocID, m.In_DocFileExt, m.In_DocCreated, m.In_DocTypeID, 
				case 
				when m.In_DocTypeID = 25 then 'dr' 
				when m.In_DocTypeID = 16 then 'bol' 
				end as DocType 
				 from DELIVERYDOCS.dbo.Child C
				 inner join DELIVERYDOCS.dbo.Main M on M.In_DocID = C.In_DocFamilyID
				 where FBNumber = '` + Clm[j].Bill_number + `'
				 and M.DeliveryDate is not null
				 and m.In_DocTypeID in (` + fileStr + `)
				 )
				select max(In_DocCreated) as DocCreated, max(In_DocID) as DocName, FBNumber, in_docFileExt as DocExt, In_DocTypeID as DocTypeID, DocType, In_DocLocation as DocLcation
				from cte
				group by  FBNumber, In_DocTypeID, in_docFileExt, DocType, In_DocLocation`
			//log.Println("syn stmt = " + stmt)
			rows, err := SynDb.Query(stmt)
			if err != nil {
				log.Println(err)
			}
			defer rows.Close()
			for rows.Next() {
				f := new(File)
				err = rows.Scan(&f.DocCreated, &f.DocName, &f.FBNumber, &f.DocExt, &f.DocTypeID, &f.DocType, &f.DocLocation)
				if err != nil {
					log.Println(err)
				}
				Files = append(Files, f)
				fileName := f.DocName + "." + f.DocExt
				log.Println("FileName3 = " + fileName)

				//add to list of files copied.
				if f.DocType == "bol" {
					filesNeeded = append(filesNeeded, "bol")
				}
				if f.DocType == "dr" {
					filesNeeded = append(filesNeeded, "dr")
				}

				// get the token
				token := getToken()

				//Get the file

				getPdfSyn(fileName, f.DocLocation, token, f.DocType, f.FBNumber)
			}
			rows.Close()
		}
	}

	for m := 0; m < len(fileToGet); m++ {
		found := true
		for n := 0; n < len(Files); n++ {
			if Files[n].DocType == fileToGet[m] {
				if found {
					found = true
				}
			} else {
				found = false
			}
		}
		if !found && Clm[j].FileError {
			log.Println("file was not found")

			Clm[j].FileError = found
		}
	}

	// make sure all the files were retreived.  If any are missing then raise the error.
	if len(DocsNeeded) != len(filesNeeded) {
		for i := range filesNeeded {
			foundFile := false
			for j := range DocsNeeded {
				if filesNeeded[i] == DocsNeeded[j] {
					//found
					if foundFile {
						foundFile = true
						log.Println("file type was found" + filesNeeded[i] + "  " + DocsNeeded[j])
						break
					}
				}
			}
			if !foundFile {
				Clm[j].FileError = true
				log.Println("No file type was found" + filesNeeded[i])
				break
			}

		}
	}
}
func GetConfig() {
	//	fetch mail config data from database

	stmt := `select distinct ID, IDNAME, DOCSNEEDED, EMAIL_FROM, EMAIL_TO, SUBJECT, GREETING, BODY, CLOSING, NOTE 
			 from DYLT_SETTLEMENT_LETTERS 
			 where END_DATE is null
			 with ur`
	rows, err := TmwDb.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	mc := make([]*MailConfig, 0)
	for rows.Next() {
		cl := new(MailConfig)
		err := rows.Scan(&cl.Id, &cl.Idname, &cl.Docsneeded, &cl.EmailFrom, &cl.EmailTo, &cl.Subject, &cl.Greeting, &cl.Body, &cl.Closing, &cl.Note)
		if err != nil {
			log.Fatal(err)
		}
		mc = append(mc, cl)
	}
	log.Println(mc)
	Mc = mc
}
func GetClaims(listId string) {
	log.Println(listId)
	// fetch mail config data from database
	stmt := `select lc.List_id, C.CLAIM_ID, TR.RATE_CLIENT_ID, c.claimant, t.DETAIL_LINE_ID, t.BILL_NUMBER, 
    coalesce(C.CLAIM_STATUS, 'nil'), coalesce(C.user3, 'nil'), coalesce(c.CONTACT_NAME, 'nil'), 
    coalesce(c.CLAIMANT_EMAIL, 'nil'), coalesce(LC.UPDATED_WHEN, 'nil'),
	coalesce(c.CLAIM_DATE, 'nil'), coalesce(c.TRACE_NO, 'nil'), 
    coalesce(c.AMT_CLAIMED, 0.0),    
	coalesce(cl.NAME, 'nil'), 
	coalesce(cl.ADDRESS_1, 'nil'), coalesce(cl.ADDRESS_2, 'nil'), coalesce(cl.CITY, 'nil'), 
	coalesce(cl.PROVINCE, 'nil'), 
    coalesce(cl.POSTAL_CODE, 'nil'), 
    coalesce(LC.ASSIGNED_TO, 'nil'), 
    coalesce(LI.ITEM_DESCR, 'nil'), 
    coalesce(LI.ITEM_IS_REQUIRED, 'nil'), 
    coalesce(Lc.COMMENT, 'nil'),
	coalesce(d.POD_SIGNED_BY, 'nil'), 
	coalesce(d.POD_SIGNED_ON, '01/01/2000'),
	coalesce(cl.SALES_REP, 'nil'),
	count(sl.ID) as Multiple
	from CLAIM C 
	join LIST_CHECKIN LC on lc.LIST_CODE = c.CLAIM_ID
	join LIST_ITEM LI on lc.LIST_ID = li.ITEM_ID
	join TLORDER T on C.ORDER_ID = t.DETAIL_LINE_ID
	left join CLIENT CL on C.CLAIMANT = CL.CLIENT_ID
	left join DYLT_SETTLEMENT_LETTERS sl on lc.LIST_ID = sl.ID and sl.END_DATE is null
	left join POD D on d.DLID = t.DETAIL_LINE_ID and d.TX_TYPE = 'Drop'
	left outer join tlorder_rates TR on TR.dlid = t.DETAIL_LINE_ID 
	where lc.UPDATED_WHEN between  '` + *startDate + `' and '` + *endDate + `' 
	  and lc.List_id = ` + listId + ` 
	and C.CLAIM_STATUS in ('CLOSED', 'OPEN')
	and lc.IS_COMPLETE = 'True'
	group by lc.List_id, C.CLAIM_ID, TR.RATE_CLIENT_ID, c.claimant, t.DETAIL_LINE_ID, t.BILL_NUMBER,
	coalesce(C.CLAIM_STATUS, 'nil'), coalesce(C.user3, 'nil'), coalesce(c.CONTACT_NAME, 'nil'), 
	coalesce(c.CLAIMANT_EMAIL, 'nil'), coalesce(LC.UPDATED_WHEN, 'nil'),
	coalesce(c.CLAIM_DATE, 'nil'), coalesce(c.TRACE_NO, 'nil'), 
	coalesce(c.AMT_CLAIMED, 0.0),
	coalesce(cl.NAME, 'nil'), 
	coalesce(cl.ADDRESS_1, 'nil'), coalesce(cl.ADDRESS_2, 'nil'), coalesce(cl.CITY, 'nil'), 
	coalesce(cl.PROVINCE, 'nil'), 
	coalesce(cl.POSTAL_CODE, 'nil'),
	coalesce(LC.ASSIGNED_TO, 'nil'), coalesce(LI.ITEM_DESCR, 'nil'), coalesce(LI.ITEM_IS_REQUIRED, 'nil'), coalesce(Lc.COMMENT, 'nil'), 
	coalesce(d.POD_SIGNED_BY, 'nil'), coalesce(d.POD_SIGNED_ON, '01/01/2000'),coalesce(cl.SALES_REP, 'nil')
	with ur`

	rows, err := TmwDb.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		claim := new(Claim)
		claim.FileError = false
		err := rows.Scan(&claim.LcId, &claim.Claim_id, &claim.RateClientId, &claim.ClaimantID, &claim.Detail_line_id, &claim.Bill_number,
			&claim.Claim_status, &claim.Analyst, &claim.Contact_name, &claim.Claimant_email, &claim.UpdatedWhen,
			&claim.Claim_Date, &claim.TraceNo,
			&claim.AmtClamed,
			&claim.Company,
			&claim.Address1, &claim.Address2, &claim.City,
			&claim.Provence, &claim.PostalCode,
			&claim.Assigned_to, &claim.Item_descr,
			&claim.Item_is_required, &claim.Comments, &claim.POD_Signed, &claim.POD_Date, &claim.SalesRep, &claim.Multiple)
		if err != nil {
			log.Fatal(err)
		}
		Clm = append(Clm, claim)
		//log.Println(claim)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}
func GetNote(dlid string, j int) bool {

	stmt := `SELECT coalesce(cast(THE_NOTE as varchar(32000)), 'nil')  FROM NOTES N WHERE PROG_TABLE = 'TLORDER'  AND NOTE_TYPE = '3'  AND ID_KEY = '` + dlid + `' Fetch first row only`
	//log.Println(stmt)
	rows, err := TmwDb.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		claim := new(Claim)
		err := rows.Scan(&claim.Note)
		if err != nil {
			log.Fatal(err)
		}
		if claim.Note == "nil" {
			log.Println("GetNote  - " + claim.Note)
			Clm[j].Note = ""
			return false
		} else {
			s := strings.Replace(claim.Note, "\n", "<br>", -1)
			Clm[j].Note = s
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return true
}

func getPdfSyn(fileName string, paths string, token string, docType string, fb string) {
	log.Println(fb)
	//	log.Println(pathtofile)
	log.Println("getPdfSyn call to synergize fileName:" + fileName)
	name := strings.TrimRight(strings.SplitAfter(fileName, ".")[0], ".")

	fileName = ".\\tmp_images\\" + name + ".pdf"
	//log.Println("Src file = " + pathtofile+"\\"+fileName)
	log.Println(fileName)
	//	copyFile(pathtofile+"\\"+fileName, ".\\tmp_images\\"+fileName)

	// get the pdf
	url := "https://api.dylt.com/image/" + fb + "/" + docType + "/pdf?userName=AUTORGH&password=alwayskeepasmile"
	log.Println(url)

	request, _ := http.NewRequest("GET", url, nil)
	//request.Header.Set("Content-Type", "application/pdf")
	request.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	response, err := client.Do(request)
	defer response.Body.Close()
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
	}
	if response.StatusCode != http.StatusOK {
		log.Println("Status error: %v", response.StatusCode)
	} else {
		img, err := os.Create(fileName)
		if err != nil {
			log.Println(err)
		}
		defer img.Close()

		//data, _ := ioutil.ReadAll(response.Body)
		b, _ := io.Copy(img, response.Body)
		if err != nil {
			log.Println(err)
		}
		log.Println("File Size: ", b)
		img.Close()
		LocalFile = append(LocalFile, fileName)
	}
}

// Copy a file
func copyFile(src, dest string) {
	// Open original file
	originalFile, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer originalFile.Close()

	// Create new file
	newFile, err := os.Create(dest)
	if err != nil {
		log.Fatal(err)
	}
	defer newFile.Close()

	// Copy the bytes to destination from source
	bytesWritten, err := io.Copy(newFile, originalFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Copied %d bytes.", bytesWritten)

	// Commit the file contents
	// Flushes memory to disk
	err = newFile.Sync()
	if err != nil {
		log.Fatal(err)
	}
}
func parseString(docsNeeded string) []string {
	//
	log.Println("Incomming string to parseString: " + docsNeeded)
	//reset the slice to o length
	needed := make([]string, 0)

	r := csv.NewReader(strings.NewReader(docsNeeded))
	//var docs = DocsNeeded
	//fmt.Println(reflect.TypeOf(docs))
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
		}
		needed = record

		log.Println("record1 " + record[0])
		log.Println(DocsNeeded)
		//fmt.Println(reflect.TypeOf(record))

	}
	//DocsNeeded = append(DocsNeeded, csvData)
	log.Println("System Docs Needed.")
	log.Println(needed)
	return needed
}

func sendEmail(id, fromEmail, toEmail string, claimNo int, isValid bool) bool {

	d := gomail.NewDialer("DLEXCH01.daylight.ads", 587, "etanner", Conf.MailPwd)

	S, err := d.Dial()
	if err != nil {
		log.Fatal(err)
	}
	//set up the email
	m := gomail.NewMessage()

	// set the toEmail if error
	toEmail = "TST_Recipient@DYLT.com" //"claimsError@dylt.com"

	var subject = ""
	log.Println("isValid status in sendmail: " + strconv.FormatBool(isValid))
	log.Println("Claim error status in sendmail: " + strconv.FormatBool(Clm[claimNo].FileError))
	if !isValid || Clm[claimNo].FileError {
		//Error !
		log.Println("isValid state: " + strconv.FormatBool(isValid))
		log.Println("FileErrir state: " + strconv.FormatBool(Clm[claimNo].FileError))
		//errorStr = "Error"
		toEmail = "TST_Review@DYLT.com" //"claimsError@dylt.com"
		logEmail(id, fromEmail, toEmail, strconv.Itoa(Clm[claimNo].Detail_line_id), strconv.Itoa(Clm[claimNo].Claim_id),
			Clm[claimNo].Bill_number, Clm[claimNo].Contact_name, Clm[claimNo].Assigned_to, Clm[claimNo].Note, Clm[claimNo].Multiple, "Error")
		log.Println("Error with formating.  Sending to error email box.")
		m.SetHeader("From", fromEmail)
		m.SetHeader("To", toEmail) //, "claims@madegoods.com") //c.CLAIMANT_EMAIL
		subject = "ERROR !!! - Settlement Letter, Claim ID " + strconv.Itoa(Clm[claimNo].Claim_id) + ", for Pro " + Clm[claimNo].Bill_number + " - Error !!!"
	} else {
		// No Error
		logEmail(id, fromEmail, toEmail, strconv.Itoa(Clm[claimNo].Detail_line_id), strconv.Itoa(Clm[claimNo].Claim_id),
			Clm[claimNo].Bill_number, Clm[claimNo].Contact_name, Clm[claimNo].Assigned_to, Clm[claimNo].Note, Clm[claimNo].Multiple, "Success")
		log.Println("Formatting is correct.  Sending out email.")
		m.SetHeader("From", fromEmail)
		m.SetHeader("To", toEmail)                        //, "claims@madegoods.com") //Clm.CLAIMANT_EMAIL
		m.SetAddressHeader("Cc", "TST_CC@DYLT.com", "")   // Claims@dylt.com
		m.SetAddressHeader("Bcc", "TST_BCC@DYLT.com", "") // Clm[claimNo].SalesRep+ "@dylt.com", Clm[claimNo].SalesRep )
		subject = "Settlement Letter, Claim ID " + strconv.Itoa(Clm[claimNo].Claim_id) + ", for Pro " + Clm[claimNo].Bill_number
	}

	// set up the email with the data from above
	//var subject = ""
	//if errorStr > "" {
	//	subject = "ERROR !!! - Settlement Letter, Claim ID " + strconv.Itoa(Clm[claimNo].Claim_id) + ", for Pro " + Clm[claimNo].Bill_number + " - Error !!!"
	//} else {
	//	subject = "Settlement Letter, Claim ID " + strconv.Itoa(Clm[claimNo].Claim_id) + ", for Pro " + Clm[claimNo].Bill_number
	//}

	m.SetHeader("Subject", subject)

	a := strings.TrimPrefix(OFiles[0], ".\\tmp_images\\")
	log.Println("File Name Send Email " + a)

	if len(LocalFile) > 0 {
		//m.Embed(LocalFile[0])
		log.Println("Attaching LocalFile")
		for i := 0; i < len(LocalFile); i++ {
			log.Println(LocalFile[i])
			m.Attach(LocalFile[i])
		}
	}
	//if len(DocsNeeded) > 0 {
	//	log.Println("Attaching DocsNeeded")
	//	for l := 0; l < len(DocsNeeded); l++ {
	//		log.Println(DocsNeeded[l])
	//		m.Attach(DocsNeeded[l])
	//	}
	//}

	headerStr := getHeadder(claimNo)
	//log.Println(headerStr)
	footerStr := getFooter(claimNo)
	//log.Println(footerStr)

	m.Embed(".\\tmp_images\\image.tif")

	var bodyStr = `<img src="cid:image.tif" alt="My image" /><br><br><br>` + headerStr + Clm[claimNo].Note + footerStr
	//log.Println("note from email - " + Clm[claimNo].Note)

	m.SetBody("text/html", bodyStr)

	// send the email.
	if err := gomail.Send(S, m); err != nil {
		log.Printf("Could not send email to %q: %v", Clm[claimNo].Contact_name, err)
		return true
	}

	// wait 15 sec for the mailserver.  There is a limit to the # of email per min going out.
	timer1 := time.NewTimer(15 * time.Second)
	<-timer1.C

	m.Reset()
	return false
}

func getHeadder(claimNo int) string {
	str := ""
	str = "" + Clm[claimNo].Contact_name + "<br>"
	str = str + Clm[claimNo].Company + "<br>"
	if Clm[claimNo].Address1 != "" || Clm[claimNo].Address1 != "nil" {
		str = str + Clm[claimNo].Address1 + " "
	} else {
		log.Println("Address1 is null")
	}
	if Clm[claimNo].City != "" || Clm[claimNo].City != "nil" {
		str = str + Clm[claimNo].City + ", "
	} else {
		log.Println("City is null")
	}
	if Clm[claimNo].Provence != "" || Clm[claimNo].Provence != "nil" {
		str = str + Clm[claimNo].Provence + " "
	} else {
		log.Println("State is null")
	}
	if Clm[claimNo].PostalCode != "" || Clm[claimNo].PostalCode != "nil" {
		str = str + Clm[claimNo].PostalCode + "<br>"
	} else {
		log.Println("State is null")
	}
	str = str + "Claimant: " + Clm[claimNo].Claimant_email + "<br>"
	str = str + "Sales: " + Clm[claimNo].SalesRep + "<br><br><br>"

	str = str + "Your Claim #: " + Clm[claimNo].TraceNo + "<br>"
	str = str + "Our Claim #: " + strconv.Itoa(Clm[claimNo].Claim_id) + "<br>"
	str = str + "Freight Bill#: " + Clm[claimNo].Bill_number + "<br>"
	str = str + "Claim Amount: $" + strconv.FormatFloat(Clm[claimNo].AmtClamed, 'f', 2, 64) + "<br><br><br>"

	//if Clm[claimNo].Contact_name == ""{
	//	str = str + "Dear Claimant, <br><br>"
	//}else{
	//	str = str + "Dear " + Clm[claimNo].Contact_name + ",<br><br>"
	//}
	return str
}
func getFooter(claimNo int) string {
	str := "<br><br>Sincerely,<br>" +
		"         " + Clm[claimNo].Analyst + " <br>" +
		"         Claims Analyst " +
		//"<br><br>" +
		//"Attachment: Delivery Receipt for the same Freight Bill #" + Clm[claimNo].Bill_number + " from Synergize " +
		"<br><br><br><br>"

	str = str + "Note:</b>  Rebuttals must be submitted in writing with additional information supporting your claim.  "
	str = str + "Please reference claim # " + strconv.Itoa(Clm[claimNo].Claim_id) + "  and fax to 888-845-9251 or email to claims@dylt.com.<br>"

	str = str + "<br><br><br><hr><br><center>Daylight Transport LLC 1501 Hughes Way, Suite 200, Long Beach, California 90810 ~ 800-468-9999 ~ 888-845-9251 Fax ~ www.dylt.com<center>"

	return str
}
func logEmail(id, fromEmail, toEmail, dlid, claimID, billNo, contact, assignedTo, note string, multiple int, state string) {
	// write the error
	ms := new(State)
	s := strconv.Itoa(multiple)
	ms.Claim_id = id
	ms.Email_from = fromEmail
	ms.Email_to = toEmail
	ms.Detail_line_id = dlid
	ms.Claim_id = claimID
	ms.Bill_number = billNo
	ms.Contact_name = contact
	ms.Assigned_to = assignedTo
	ms.Note = note
	ms.Multiple = s
	ms.State = state

	MailStatus = append(MailStatus, ms)

}
func formatNote(i, id int) string {
	//
	var filename = ""
	var s = ""
	if id == 101 {
		filename = "101.txt"
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Print(err)
		}
		s = string(b)
		s = strings.Replace(s, "<name>", Clm[i].Contact_name, -1)
		s = strings.Replace(s, "<comapany>", Clm[i].Company, -1)
		s = strings.Replace(s, "<address>", Clm[i].Address1, -1)
		s = strings.Replace(s, "<city>", Clm[i].City, -1)
		s = strings.Replace(s, "<state>", Clm[i].Provence, -1)
		s = strings.Replace(s, "<zip>", Clm[i].PostalCode, -1)
		s = strings.Replace(s, "<your_claim>", Clm[i].TraceNo, -1)
		s = strings.Replace(s, "<claim>", strconv.Itoa(Clm[i].Claim_id), -1)
		s = strings.Replace(s, "<bill>", Clm[i].Bill_number, -1)
		s = strings.Replace(s, "<amount	>", strconv.FormatFloat(Clm[i].AmtClamed, 'E', 2, 64), -1)
		s = strings.Replace(s, "<date_shipped>", Clm[i].Claim_Date.String(), -1)
		s = strings.Replace(s, "<del_date>", Clm[i].POD_Date.String(), -1)
		s = strings.Replace(s, "<received_date>", Clm[i].Claim_Date.String(), -1)
		s = strings.Replace(s, "<day_dif>", strconv.Itoa(Clm[i].DaysDiff), -1)
		s = strings.Replace(s, "\n", "<br>", -1)
		log.Println("Replaced strings - 101")
	}
	if id == 146 {
		filename = "146.txt"
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Print(err)
		}
		s = string(b)
		s = strings.Replace(s, "<name>", Clm[i].Contact_name, -1)
		s = strings.Replace(s, "<comapany>", Clm[i].Company, -1)
		s = strings.Replace(s, "<address>", Clm[i].Address1, -1)
		s = strings.Replace(s, "<city>", Clm[i].City, -1)
		s = strings.Replace(s, "<state>", Clm[i].Provence, -1)
		s = strings.Replace(s, "<zip>", Clm[i].PostalCode, -1)
		s = strings.Replace(s, "<your_claim>", Clm[i].TraceNo, -1)
		s = strings.Replace(s, "<claim>", strconv.Itoa(Clm[i].Claim_id), -1)
		s = strings.Replace(s, "<bill>", Clm[i].Bill_number, -1)
		s = strings.Replace(s, "<amount	>", strconv.FormatFloat(Clm[i].AmtClamed, 'E', 2, 64), -1)
		s = strings.Replace(s, "<date_shipped>", Clm[i].Claim_Date.String(), -1)
		s = strings.Replace(s, "<del_date>", Clm[i].POD_Date.String(), -1)
		s = strings.Replace(s, "<received_date>", Clm[i].Claim_Date.String(), -1)
		s = strings.Replace(s, "<day_dif>", strconv.Itoa(Clm[i].DaysDiff), -1)
		s = strings.Replace(s, "\n", "<br>", -1)
		log.Println("Replaced strings - 146")
	}
	if id == 99 {
		filename = "99.txt"
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Print(err)
		}
		s = strings.Replace(string(b), "\n", "<br>", -1)
		//s = string(b)
	}
	return s
}
func sendMailStatus() {
	//
	d := gomail.NewDialer("DLEXCH01.daylight.ads", 587, "etanner", Conf.MailPwd)

	S, err := d.Dial()
	if err != nil {
		log.Fatal(err)
	}
	//send the email
	m := gomail.NewMessage()
	m.SetHeader("From", "etanner@dylt.com")
	m.SetHeader("To", "etanner@dylt.com") //, "claims@madegoods.com") //c.CLAIMANT_EMAIL
	m.SetAddressHeader("Cc", "etanner@dylt.com", "")
	date := jodaTime.Format("YYYY.MM.dd", time.Now())

	var str = "Automated Settlement Letter batch job for " + date
	m.SetHeader("Subject", str)
	m.Attach("log_file")

	var bodyStr = "<br><br>"

	for _, i := range MailStatus {
		bodyStr = bodyStr + i.State + " - " + i.Claim_id + " assigned to " + i.Analyst + " was processed. <br> \n"
	}
	m.SetBody("text/html", bodyStr)
	//log.Println("Body String:  "+bodyStr)
	// send the email.
	if err := gomail.Send(S, m); err != nil {
		log.Printf("Could not send email %v", err)
	}

}
func businessDays(from time.Time, to time.Time) int {
	//start format is '01/06/2019'
	//date := 0
	log.Println("start date ") // + start)
	log.Println("end date ")   // + end)

	totalDays := float32(to.Sub(from) / (24 * time.Hour))
	weekDays := float32(from.Weekday()) - float32(to.Weekday())
	businessDays := int(1 + (totalDays*5-weekDays*2)/7)
	if to.Weekday() == time.Saturday {
		businessDays--
	}
	if from.Weekday() == time.Sunday {
		businessDays--
	}

	return businessDays
}
func LoadConfiguration(file string) Config {
	var config Config
	pwd, _ := os.Getwd()
	configFile, err := os.Open(pwd + "\\" + file)
	defer configFile.Close()
	if err != nil {
		log.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}
func clearTmpFiles() {
	// remove all temp files, if any
	directory := "./tmp_Images"
	log.Println(directory)
	dirRead, _ := os.Open(directory)
	dirFiles, _ := dirRead.Readdir(0)

	for index := range dirFiles {
		fileHere := dirFiles[index]
		nameHere := fileHere.Name()
		fullPath := directory + "\\" + nameHere
		log.Println(fullPath)
		os.Remove(fullPath)
	}
	//os.Remove("log_file")
	date := jodaTime.Format("YYYY_MM_dd_HH_mm_ss", time.Now())
	os.Rename("log_file", "log_file-"+date)
}
func getToken() string {
	//Consumer Key: x5Vxusddiy2pYqwpZytwxqkG0lW7Z6a5
	//Consumer Secret: ThzO25vxF0RDuA2U
	body := strings.NewReader(`client_secret=P0AGMIlIAFC1vEqn&grant_type=client_credentials&client_id=QLVqxgQk85apoXB8AFAeSOYTv4RR53lh`)
	req, err := http.NewRequest("POST", "https://api.dylt.com/oauth/client_credential/accesstoken?grant_type=client_credentials", body)
	if err != nil {
		// handle err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	token, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	var data map[string]string
	json.Unmarshal(token, &data)
	//log.Println(data)
	//log.Println(data["access_token"])
	return data["access_token"]
}
func validateFields(claim Claim, mailConfig MailConfig) bool {
	log.Println("Got to validateFields")
	var isValid bool

	if claim.Multiple > 1 || strconv.Itoa(claim.Claim_id) == "" || strconv.Itoa(claim.Claim_id) == "nil" || claim.Bill_number == "" || claim.Bill_number == "nil" ||
		claim.Contact_name == "" || claim.Contact_name == "nil" || claim.Company == "" || claim.Company == "nil" || //len(OFiles[0]) == 0 ||
		claim.Note == "" || claim.Note == "nil" || claim.Analyst == "" || claim.Analyst == "nil" || claim.Address1 == "" || claim.Address1 == "nil" ||
		claim.City == "" || claim.City == "nil" || claim.Provence == "" || claim.Provence == "nil" || claim.PostalCode == "" || claim.PostalCode == "nil" {

		if claim.Address1 == "" {
			log.Println("Error with claim.Address1 = " + claim.Address1)
		}
		if claim.Address1 == "nil" {
			log.Println("Error with claim.Address1 nil = " + claim.Address1)
		}
		if claim.City == "" {
			log.Println("Error with claim.City = " + claim.City)
		}
		if claim.City == "nil" {
			log.Println("Error with claim.City nil = " + claim.City)
		}
		if claim.Provence == "" {
			log.Println("Error with claim.Provence = " + claim.Provence)
		}
		if claim.Provence == "nil" {
			log.Println("Error with claim.Provence nil = " + claim.Provence)
		}
		if claim.PostalCode == "" {
			log.Println("Error with claim.PostalCode = " + claim.PostalCode)
		}
		if claim.PostalCode == "nil" {
			log.Println("Error with claim.PostalCode nil = " + claim.PostalCode)
		}
		if claim.Multiple > 1 {
			log.Println("Error with claim.Multiple > 1")
		}
		if strconv.Itoa(claim.Claim_id) == "" {
			log.Println("Error with claim.Claim_id = " + strconv.Itoa(claim.Claim_id))
		}
		if strconv.Itoa(claim.Claim_id) == "nil" {
			log.Println("Error with claim.Claim_id nil = " + strconv.Itoa(claim.Claim_id))
		}
		if claim.Bill_number == "" {
			log.Println("Error with claim.Bill_Number = " + claim.Bill_number)
		}
		if claim.Bill_number == "nil" {
			log.Println("Error with claim.Bill_Number nil =  " + claim.Bill_number)
		}
		if claim.Contact_name == "" {
			log.Println("Error with claim.Contact_name = " + claim.Contact_name)
		}
		if claim.Contact_name == "nil" {
			log.Println("Error with claim.Contact_name = nil" + claim.Contact_name)
		}
		if claim.Assigned_to == "" {
			log.Println("Error with claim.Company = " + claim.Company)
		}
		if claim.Assigned_to == "nil" {
			log.Println("Error with claim.Company nil = " + claim.Company)
		}
		if claim.Note == "" {
			log.Println("Error with claim.Note = " + claim.Note)
		}
		if claim.Note == "nil" {
			log.Println("Error with claim.Note = 'nil' " + claim.Note)
		}
		if claim.Analyst == "" {
			log.Println("Error with claim.Analyst = " + claim.Analyst)
		}
		if claim.Analyst == "nil" {
			log.Println("Error with claim.Analyst nil = " + claim.Analyst)
		}

		isValid = false
		logEmail(strconv.Itoa(mailConfig.Id), mailConfig.EmailFrom, mailConfig.EmailTo, strconv.Itoa(claim.Detail_line_id),
			strconv.Itoa(claim.Claim_id), claim.Bill_number, claim.Contact_name, claim.Assigned_to, claim.Note,
			claim.Multiple, "Error")
		log.Println("Error with formating.  Sending to error email box.")
	} else {
		logEmail(strconv.Itoa(mailConfig.Id), mailConfig.EmailFrom, mailConfig.EmailTo, strconv.Itoa(claim.Detail_line_id),
			strconv.Itoa(claim.Claim_id), claim.Bill_number, claim.Contact_name, claim.Assigned_to, claim.Note,
			claim.Multiple, "Success")
		isValid = true
		log.Println("Formatting is correct.  Sending out email.")
	}

	return isValid
}
func copyODrive() {
	//
	pathtofiles := "O:\\DEPARTMENTS\\Claims\\_PDF\\"
	files, err := ioutil.ReadDir(pathtofiles)
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}
	for _, file := range files {
		copyFile(pathtofiles+"\\"+file.Name(), ".\\tmp_images\\"+file.Name())
		OFiles = append(OFiles, ".\\tmp_images\\"+file.Name())
	}

}
