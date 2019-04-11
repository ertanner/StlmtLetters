package db

import (
	. "../config"
	. "../utilities"
	. "../web"
	"strconv"
	"strings"
)

const layoutUS = "01-02-2006"

func GetCustTariff(i, j int) bool {
	Log.Println("GetCustTariff")

	// generate the string of comma delimited values to put into the frtBill values
	stmt := `select top 1 In_DocID, coalesce(In_DocName, 'nil')as DocName, In_DocTypeID, In_DocFileExt, In_DocLocation, ClientID, In_DocCreated  
				from CONTRACTDOCS.dbo.Main
				where ClientID = ` + Clm[j].ClaimantID + `
				and In_DocTypeID = 13
				and In_DocCreated < '` + Clm[j].Claim_Date.Format(layoutUS) + `'`

	Log.Println("Syn Cust Tariff Stmt = " + stmt)
	rows, err := SynTariffDb.Query(stmt)
	fileFound := false
	if err != nil {
		Log.Println(err)
	} else {
		defer rows.Close()

		for rows.Next() {
			f := new(File)
			err = rows.Scan(&f.DocId, &f.DocName, &f.DocTypeID, &f.DocExt, &f.DocLocation, &f.ClientId, &f.DocCreated)
			if err != nil {
				Log.Println(err)
			}
			Files = append(Files, f)
			fileName := f.DocId + "." + f.DocExt
			Log.Println("FileNameTariff = " + fileName)
			if fileName != "" {
				// strings.TrimLeft( f.DocLocation,"1\\CONTRACTDOCS")
				str := strings.Replace(f.DocLocation, "1\\CONTRACTDOCS", "", -1)
				Log.Println("Trimmed string " + string(str))
				Log.Println("Y:" + string(str) + "\\" + fileName)
				CopyFile("Y:"+string(str)+"\\"+fileName, ".\\tmp_Images\\"+fileName)
				LocalFile = append(LocalFile, ".\\tmp_Images\\"+fileName)
				if !Clm[j].FileError {
					fileFound = true
				}
			} else {
				Log.Println("fileName is empty")
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
		additionFiles := ParseString(Clm[j].Comments)
		for z := 0; z < len(additionFiles); z++ {
			Log.Println("NMFC Item " + additionFiles[z] + ".pdf")
			LocalFile = append(LocalFile, ".\\tmp_images\\NMFC Item "+additionFiles[z]+".pdf")
			filesNeeded = append(filesNeeded, "nmfc")
		}
	}

	// get 110Tarrif and NMFTA if needed
	if len(DocsNeeded) > 0 {
		for k := 0; k < len(DocsNeeded); k++ {
			if DocsNeeded[k] == "110Tariff" {
				Log.Println("Gettign 110Tariff")
				LocalFile = append(LocalFile, ".\\tmp_images\\110Tariff.pdf")
				filesNeeded = append(filesNeeded, "110Tariff")
			}
			if DocsNeeded[k] == "nmfta" {
				Log.Println("Getting NMFTA files")
				LocalFile = append(LocalFile, ".\\tmp_images\\NMFTA_Item_300105.pdf")
				LocalFile = append(LocalFile, ".\\tmp_images\\NMFTA_Item_300135.pdf")
				filesNeeded = append(filesNeeded, "nmfta")
			}
		}
	}

	//Get the tif from the Synergize
	fileToGet := make([]string, 0)
	for k := 0; k < len(DocsNeeded); k++ {
		Log.Println(DocsNeeded[k])
		if DocsNeeded[k] == "bol" || DocsNeeded[k] == "dr" {
			fileToGet = append(fileToGet, DocsNeeded[k])
		}
		if Clm[j].ClaimantID == Clm[j].RateClientId {
			if DocsNeeded[k] == "custTariff" {
				isValid := GetCustTariff(i, j)
				if isValid {
					Log.Println("found custTariff")
					filesNeeded = append(filesNeeded, "custTariff")
				} else {
					Log.Println("could not find cust tariff")
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
				Log.Println("fileStr1" + fileStr)
			}
			if fileToGet[l] == "dr" {
				fileStr = fileStr + "25, "
				Log.Println("fileStr2" + fileStr)
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
			//Log.Println("syn stmt = " + stmt)
			rows, err := SynDb.Query(stmt)
			if err != nil {
				Log.Println(err)
			}
			defer rows.Close()
			for rows.Next() {
				f := new(File)
				err = rows.Scan(&f.DocCreated, &f.DocName, &f.FBNumber, &f.DocExt, &f.DocTypeID, &f.DocType, &f.DocLocation)
				if err != nil {
					Log.Println(err)
				}
				Files = append(Files, f)
				fileName := f.DocName + "." + f.DocExt
				Log.Println("FileName3 = " + fileName)

				//add to list of files copied.
				if f.DocType == "bol" {
					filesNeeded = append(filesNeeded, "bol")
				}
				if f.DocType == "dr" {
					filesNeeded = append(filesNeeded, "dr")
				}

				// get the token
				token := GetToken()

				//Get the file

				GetPdfSyn(fileName, f.DocLocation, token, f.DocType, f.FBNumber)
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
			Log.Println("file was not found")

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
						Log.Println("file type was found" + filesNeeded[i] + "  " + DocsNeeded[j])
						break
					}
				}
			}
			if !foundFile {
				Clm[j].FileError = true
				Log.Println("No file type was found" + filesNeeded[i])
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
		Log.Fatal(err)
	}
	defer rows.Close()

	mc := make([]*MailConfig, 0)
	for rows.Next() {
		cl := new(MailConfig)
		err := rows.Scan(&cl.Id, &cl.Idname, &cl.Docsneeded, &cl.EmailFrom, &cl.EmailTo, &cl.Subject, &cl.Greeting, &cl.Body, &cl.Closing, &cl.Note)
		if err != nil {
			Log.Fatal(err)
		}
		mc = append(mc, cl)
	}
	Log.Println(mc)
	Mc = mc
}
func GetClaims(listId string) {
	Log.Println(listId)
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
	where lc.UPDATED_WHEN between  '` + *StartDate + `' and '` + *EndDate + `' 
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
		Log.Println(err)
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
			Log.Fatal(err)
		}
		Clm = append(Clm, claim)
		//Log.Println(claim)
	}
	err = rows.Err()
	if err != nil {
		Log.Fatal(err)
	}
}
func GetNote(dlid string, j int) bool {

	stmt := `SELECT coalesce(cast(THE_NOTE as varchar(32000)), 'nil')  FROM NOTES N WHERE PROG_TABLE = 'TLORDER'  AND NOTE_TYPE = '3'  AND ID_KEY = '` + dlid + `' Fetch first row only`
	//Log.Println(stmt)
	rows, err := TmwDb.Query(stmt)
	if err != nil {
		Log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		claim := new(Claim)
		err := rows.Scan(&claim.Note)
		if err != nil {
			Log.Fatal(err)
		}
		if claim.Note == "nil" {
			Log.Println("GetNote  - " + claim.Note)
			Clm[j].Note = ""
			return false
		} else {
			s := strings.Replace(claim.Note, "\n", "<br>", -1)
			//badChar := IsLetter(s)
			s = StripCtlAndExtFromUTF8(s)
			Clm[j].Note = s

			return true
		}
	}
	err = rows.Err()
	if err != nil {
		Log.Println(err)
		return false
	}
	// check for illegal characters

	return true
}
func LogClaim(j int) {
	Log.Println("logClaim")
	Log.Println(Clm[j].Claim_id)

	str := `INSERT INTO TMWIN.DYLT_SETLMNT_LTR_LOG(CLAIMID, LCID, DETAIL_LINE_ID, BILL_NUMBER, CLAIM_STATUS, ANALYST, CONTACT_NAME, 
			CLAIMANT_EMAIL, UPDATEDWHEN, CLAIM_DATE, RATECLIENTID, CLAIMANTID, TRACENO, AMTCLAMED, COMPANY, ADDRESS1, ADDRESS2, CITY, 
			PROVENCE, POSTALCODE, ASSIGNED_TO, ITEM_DESCR, ITEM_IS_REQUIRED, POD_SIGNED, POD_DATE, DAYSDIFF, COMMENTS, 
			SALESREP, MULTIPLE, FILEERROR, ISVALID)
			VALUES('` + strconv.Itoa(Clm[j].Claim_id) + `', '` + strconv.Itoa(Clm[j].LcId) + `', '` + strconv.Itoa(Clm[j].Detail_line_id) + `', '` + Clm[j].Bill_number + `', '` +
		Clm[j].Claim_status + `', '` + Clm[j].Analyst + `', '` + Clm[j].Contact_name + `', '` + Clm[j].Claimant_email + `', '` +
		Clm[j].UpdatedWhen.Format(layoutUS) + `', '` + Clm[j].Claim_Date.Format(layoutUS) + `', '` + Clm[j].RateClientId + `', '` + Clm[j].ClaimantID + `', '` + Clm[j].TraceNo + `', ` +
		strconv.FormatFloat(Clm[j].AmtClamed, 'E', -1, 64) +
		`, '` + Clm[j].Company +
		`', '` + Clm[j].Address1 + `', '` + Clm[j].Address2 + `', '` + Clm[j].City + `', '` + Clm[j].Provence + `', '` + Clm[j].PostalCode +
		`', '` + Clm[j].Assigned_to + `', '` + Clm[j].Item_descr + `', '` + Clm[j].Item_is_required + `', '` + Clm[j].POD_Signed +
		`', '` + Clm[j].POD_Date.Format(layoutUS) + `', ` + strconv.Itoa(Clm[j].DaysDiff) + `, '` + Clm[j].Comments + `', '` +
		Clm[j].SalesRep + `', ` + strconv.Itoa(Clm[j].Multiple) + `, '` + strconv.FormatBool(Clm[j].FileError) + `', '` + strconv.FormatBool(Clm[j].IsValid) + `')`

	Log.Println(" Log Claim Instert string: " + str)

	_, err := TmwDb.Exec(str)
	if err != nil {
		Log.Println(err)
	}
}
func CheckClaimLog(j int) {

	str := "select count(ClaimId) from TMWIN.DYLT_SETLMNT_LTR_LOG where CLAIMID = '" + strconv.Itoa(Clm[j].Claim_id) + "' and LCID = '" + strconv.Itoa(Clm[j].LcId) + "' with ur"
	var count int
	Log.Println(str)
	rows, err := TmwDb.Query(str)
	if err != nil {
		Log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			Log.Println(err)
		}
		//Log.Println(claim)
		if count > 0 {
			Log.Println("Found record in DB already.")
			Clm[j].IsValid = false
			Clm[j].Comments = "Found record in DB already."
		}
	}
	err = rows.Err()
	if err != nil {
		Log.Fatal(err)
	}
}
