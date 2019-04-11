package main

import (
	. "./config"
	. "./db"
	. "./email"
	. "./utilities"
	. "./web"
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/alexbrainman/odbc"
	"log"
	"os"
	"strconv"
	"time"
)

//Claims_Review@Dylt.com
//TST_Recipient@DYLT.com
//TST_CC@DYLT.com
//TST_BCC@DYLT.com
//TST_Review@DYLT.com
// TODO re write the tests

var err error
var CommentsNeeded = make([]string, 0)

var buf bytes.Buffer
var logger = log.New(&buf, "logger: ", log.Lshortfile)

//var db = flag.String("db", "", "DYLT_REP")

const layoutUS = "01-02-2006"

func init() {
	NewLog(".\\log\\log.txt")
	JobConf = LoadConfigurationJob("config.json")
	// get base config for user id and pwd
	Conf = LoadConfiguration("acxpuwd")

	//*************************************************************************
	//TmwDb
	//*************************************************************************
	TmwDb, err = sql.Open("odbc", "DSN="+JobConf.DB2dsn)
	if err != nil {
		Log.Fatal(err)
	}
	Log.Println("Connecting to TMW database...")

	TmwDb.SetMaxOpenConns(10)
	//db.SetMaxIdleConns(3)
	TmwDb.SetConnMaxLifetime(30 * time.Minute)
	TmwDb.SetMaxIdleConns(0)

	err = TmwDb.Ping()
	if err != nil {
		Log.Fatal("Ping error")
	} else {
		Log.Println("Database connection to TMW is opened...")
	}

	//*************************************************************************
	// connect to the Synergize db
	//*************************************************************************
	SynDb, err = sql.Open("odbc",
		"DSN="+JobConf.SynDbDsn+";UID="+JobConf.SynUid+";PWD="+Conf.SynPwd)
	if err != nil {
		Log.Fatal(err)
	}
	Log.Println("Connecting to the Synergize database...")
	SynDb.SetMaxOpenConns(10)
	//db.SetMaxIdleConns(3)
	SynDb.SetConnMaxLifetime(30 * time.Minute)
	SynDb.SetMaxIdleConns(0)

	err = SynDb.Ping()
	if err != nil {
		Log.Fatal("Ping error ", err)
	} else {
		Log.Println("The Synergize database connection is opened...")
	}
	//*************************************************************************
	//svc_sys_dash  or sa with pwd Syn3rg1ze
	// connect to the Synergize db
	//*************************************************************************
	SynTariffDb, err = sql.Open("odbc",
		"DSN="+JobConf.SynDbDsn+";UID="+JobConf.SynUid+";PWD="+Conf.SynPwd)
	if err != nil {
		Log.Fatal(err)
	}
	Log.Println("Connecting to the Synergize database...")
	SynTariffDb.SetMaxOpenConns(10)
	//db.SetMaxIdleConns(3)
	SynTariffDb.SetConnMaxLifetime(30 * time.Minute)
	SynTariffDb.SetMaxIdleConns(0)

	err = SynTariffDb.Ping()
	if err != nil {
		Log.Fatal("Ping error ", err)
	} else {
		Log.Println("The Synergize database connection is opened...")
	}

}

func main() {
	start := time.Now()
	fmt.Println("Start Date: " + start.String())

	// this will clear the temp image files from the tmp_images directory
	ClearTmpFiles()

	//create your file with desired read/write permissions
	f, err := os.OpenFile("log_file", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		Log.Println(err)
	}
	//defer to close when you're done with it
	defer f.Close()
	//set output of logs to f  this points the "log" commad to output to the f which is defined above
	Log.SetOutput(f)
	Log.Println("Starting mail program.")

	// parse the parameters and if empty set the dates
	flag.Parse()

	Log.Println("Start Flag Date: " + *StartDate)
	Log.Println("End Flag Date: " + *EndDate)
	if *StartDate == "" {
		*StartDate = time.Now().Format(layoutUS)
		//*startDate = "3/11/2019"
		Log.Println("Start Date from flag " + *StartDate)
	} else {
		Log.Println("Start Date from flag e " + *StartDate)
	}
	if *EndDate == "" {
		now := int(time.Now().Weekday())
		lastSun := int(time.Sunday)
		dateDiff := now - lastSun
		addDate := time.Now().AddDate(0, 0, -dateDiff)
		*EndDate = addDate.Format(layoutUS)
		Log.Println("End Date from flag " + *EndDate)
	} else {
		Log.Println("End Date from flag e " + *EndDate)
	}

	//*********************************************
	// Test code here.  remove it for prod
	//*startDate = "3/11/2019"
	//*endDate = "4/07/2019"
	//*********************************************

	// get the files off the O drive
	CopyODrive()
	Log.Println("copyODrive")

	// connect to the db to get the DB config showing the checklist items to be processed
	GetConfig()
	Log.Println("Got DB config")
	if err != nil {
		panic(err)
	}

	// ***************************************************************
	// Iterate over the list of checklist types
	// ***************************************************************
	for i := 0; i < len(Mc); i++ {
		Log.Println("  --------------------------------------------  ")
		Log.Println("  MailConfig")
		Log.Println("  --------------------------------------------  ")
		Log.Println("  --------------------------------------------  ")

		// Parse out the file type.  It is stored as a comma delited string in the db.
		DocsNeeded = DocsNeeded[:0]
		if Mc[i].Docsneeded != "" {
			DocsNeeded = ParseString(Mc[i].Docsneeded)
			Log.Println("Docs Needed line 1")
			Log.Println(DocsNeeded)
		}
		Log.Println("parseString")

		//GetOFiles(i)
		//Log.Println("Got to GetOFiles")

		// get a list of claims for the checklist item type.
		GetClaims(strconv.Itoa(Mc[i].Id))
		Log.Println("Got Claims")

		// ***************************************************************
		//Iterate over the list of claims and process them
		// ***************************************************************
		for j := 0; j < len(Clm); j++ {
			Log.Println("  --------------------------------------------  ")
			Log.Println(" Claim ID " + strconv.Itoa(Clm[j].Claim_id))
			Log.Println("  --------------------------------------------  ")
			Log.Println("  --------------------------------------------  ")
			LocalFile = LocalFile[:0]
			Files = Files[:0]
			CommentsNeeded = CommentsNeeded[:0]
			//isValid := true
			Log.Println(Clm[j])
			Clm[j].FileError = false

			Log.Println("Setting isValid = true")
			// Get all the files needed to attach to the email.
			// return an isValid = false if any file is not present

			GetFiles(i, j)
			Log.Println("isValid state after GetFiles: " + strconv.FormatBool(Clm[j].FileError))

			// calc days diff.
			if Mc[i].Id == 101 {
				Log.Println("Start Date: " + Clm[j].POD_Date.String() + "  End Date: " + Clm[j].Claim_Date.String())
				diffDay := BusinessDays(Clm[j].POD_Date, Clm[j].Claim_Date)
				Log.Println(" Days difference: " + strconv.Itoa(diffDay))
				if diffDay < 5 {
					Log.Println("diffDay is less than 5")
					Clm[j].IsValid = false
					Clm[j].Comments = "diffDay is less than 5"
				}
			}

			// get notes
			Log.Println("Get Notes")
			if Mc[i].Id == 101 || Mc[i].Id == 146 || Mc[i].Id == 99 {
				Clm[j].Note = FormatNote(j, Mc[i].Id)
				Log.Println("101 or 146 or 99 - Parsed Email Config")
				//Log.Println("J is: %d", j)
			} else {
				gotNote := GetNote(strconv.Itoa(Clm[j].Detail_line_id), j)
				if Clm[j].IsValid && !gotNote {
					Log.Println("gotNote is false as well as inValid")
					Clm[j].Comments = "gotNote is false as well as inValid"
					Clm[j].IsValid = false
				} else if Clm[j].IsValid && !gotNote {
					Log.Println("gotNote is false.  Setting isValid to false")
					Log.Println("FileError is: " + strconv.FormatBool(Clm[j].FileError))
					Clm[j].Comments = "gotNote is false.  Setting isValid to false"
					Clm[j].IsValid = false
				} else {
					Log.Println("Used Notes")
					Clm[j].IsValid = true
				}
			}

			// Log the claim
			Log.Println(Clm[j].LcId, strconv.Itoa(Mc[i].Id), Mc[i].EmailFrom, Mc[i].EmailTo, strconv.Itoa(Clm[j].Detail_line_id), strconv.Itoa(Clm[j].Claim_id),
				Clm[j].Bill_number, Clm[j].Contact_name, Clm[j].Assigned_to, Clm[j].Analyst, Clm[j].Claim_Date, Clm[j].TraceNo, Clm[j].AmtClamed,
				Clm[j].Address1, Clm[j].Address2, Clm[j].City, Clm[j].Provence, Clm[j].PostalCode,
				Clm[j].Note, Clm[j].POD_Signed, Clm[j].POD_Date, Clm[j].DaysDiff, strconv.Itoa(Clm[j].Multiple))

			// check the fields to be sure all valid data is there.  If nto then send it to the reviewClaims inBox
			if Clm[j].IsValid {
				ValidateFields(i, j)
			}
			Log.Println("isValid status from validateFields = " + strconv.FormatBool(Clm[j].IsValid))
			Log.Println("Claim file error status = " + strconv.FormatBool(Clm[j].FileError))
			//if Clm[j].FileError {
			//	isValid = false
			//}

			// check to see if this claim and list item have ever been sent before.
			// if so then flag it for review.
			CheckClaimLog(j)

			//for phase one send these to review automatically.
			if Mc[i].Id == 168 || Mc[i].Id == 174 {
				Clm[j].FileError = true
			}

			// send email
			resultErr := SendEmail(strconv.Itoa(Mc[i].Id), Mc[i].EmailFrom, Mc[i].EmailTo, j, Clm[j].IsValid)
			if resultErr {
				Log.Println("ERROR with send email.  " + strconv.Itoa(Clm[j].Claim_id) + Clm[j].Bill_number)
			}

			// Log the result in the logClaim which puts it into the DYLT_SETTLEMENT_LTR_LOG table
			LogClaim(j)
		}
		// reset Claim slice back to 0 elements
		Clm = Clm[:0]
	}

	// clean up tmp files
	SendMailStatus()
	ClearTmpFiles()
	// ***************************************************************
	// done with process.
	// ***************************************************************
}
