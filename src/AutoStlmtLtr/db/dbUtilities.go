package db

import (
	. "../config"
	. "../email"
	. "../utilities"
	"io/ioutil"
	"strconv"
	"strings"
)

func ValidateFields(i, j int) {
	Log.Println("Got to validateFields")

	if Clm[j].Multiple > 1 || strconv.Itoa(Clm[j].Claim_id) == "" || strconv.Itoa(Clm[j].Claim_id) == "nil" || Clm[j].Bill_number == "" || Clm[j].Bill_number == "nil" ||
		Clm[j].Contact_name == "" || Clm[j].Contact_name == "nil" || Clm[j].Company == "" || Clm[j].Company == "nil" || //len(OFiles[0]) == 0 ||
		Clm[j].Note == "" || Clm[j].Note == "nil" || Clm[j].Analyst == "" || Clm[j].Analyst == "nil" || Clm[j].Address1 == "" || Clm[j].Address1 == "nil" ||
		Clm[j].City == "" || Clm[j].City == "nil" || Clm[j].Provence == "" || Clm[j].Provence == "nil" || Clm[j].PostalCode == "" ||
		Clm[j].PostalCode == "nil" || Clm[j].DaysDiff <= 5 {

		if Clm[j].Address1 == "" {
			Log.Println("Error with claim.Address1 = " + Clm[j].Address1)
			Clm[j].Comments = "Error with claim.Address1 = " + Clm[j].Address1
		}
		if Clm[j].Address1 == "nil" {
			Log.Println("Error with claim.Address1 nil = " + Clm[j].Address1)
		}
		if Clm[j].City == "" {
			Log.Println("Error with claim.City = " + Clm[j].City)
		}
		if Clm[j].City == "nil" {
			Log.Println("Error with claim.City nil = " + Clm[j].City)
		}
		if Clm[j].Provence == "" {
			Log.Println("Error with claim.Provence = " + Clm[j].Provence)
		}
		if Clm[j].Provence == "nil" {
			Log.Println("Error with claim.Provence nil = " + Clm[j].Provence)
		}
		if Clm[j].PostalCode == "" {
			Log.Println("Error with claim.PostalCode = " + Clm[j].PostalCode)
		}
		if Clm[j].PostalCode == "nil" {
			Log.Println("Error with claim.PostalCode nil = " + Clm[j].PostalCode)
		}
		if Clm[j].Multiple > 1 {
			Log.Println("Error with claim.Multiple > 1")
		}
		if strconv.Itoa(Clm[j].Claim_id) == "" {
			Log.Println("Error with claim.Claim_id = " + strconv.Itoa(Clm[j].Claim_id))
		}
		if strconv.Itoa(Clm[j].Claim_id) == "nil" {
			Log.Println("Error with claim.Claim_id nil = " + strconv.Itoa(Clm[j].Claim_id))
		}
		if Clm[j].Bill_number == "" {
			Log.Println("Error with claim.Bill_Number = " + Clm[j].Bill_number)
		}
		if Clm[j].Bill_number == "nil" {
			Log.Println("Error with claim.Bill_Number nil =  " + Clm[j].Bill_number)
		}
		if Clm[j].Contact_name == "" {
			Log.Println("Error with claim.Contact_name = " + Clm[j].Contact_name)
		}
		if Clm[j].Contact_name == "nil" {
			Log.Println("Error with claim.Contact_name = nil" + Clm[j].Contact_name)
		}
		if Clm[j].Assigned_to == "" {
			Log.Println("Error with claim.Company = " + Clm[j].Company)
		}
		if Clm[j].Assigned_to == "nil" {
			Log.Println("Error with claim.Company nil = " + Clm[j].Company)
		}
		if Clm[j].Note == "" {
			Log.Println("Error with claim.Note = " + Clm[j].Note)
		}
		if Clm[j].Note == "nil" {
			Log.Println("Error with claim.Note = 'nil' " + Clm[j].Note)
		}
		if Clm[j].Analyst == "" {
			Log.Println("Error with claim.Analyst = " + Clm[j].Analyst)
		}
		if Clm[j].Analyst == "nil" {
			Log.Println("Error with claim.Analyst nil = " + Clm[j].Analyst)
		}
		if Clm[j].DaysDiff <= 5 {
			Log.Println("Days dif was less than 5")
		}

		Clm[j].IsValid = false
		LogEmail(strconv.Itoa(Mc[i].Id), Mc[i].EmailFrom, Mc[i].EmailTo, strconv.Itoa(Clm[j].Detail_line_id),
			strconv.Itoa(Clm[j].Claim_id), Clm[j].Bill_number, Clm[j].Contact_name, Clm[j].Assigned_to, Clm[j].Note,
			Clm[j].Multiple, "Error")
		Log.Println("Error with formating.  Sending to error email box.")
	} else {
		LogEmail(strconv.Itoa(Mc[i].Id), Mc[i].EmailFrom, Mc[i].EmailTo, strconv.Itoa(Clm[j].Detail_line_id),
			strconv.Itoa(Clm[j].Claim_id), Clm[j].Bill_number, Clm[j].Contact_name, Clm[j].Assigned_to, Clm[j].Note,
			Clm[j].Multiple, "Success")
		Log.Println("Formatting is correct.  Sending out email.")
		Clm[j].IsValid = true
	}
}

func FormatNote(i, id int) string {
	//
	var filename = ""
	var s = ""
	if id == 101 {
		filename = "101.txt"
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			Log.Print(err)
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
		Log.Println("Replaced strings - 101")
	}
	if id == 146 {
		filename = "146.txt"
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			Log.Print(err)
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
		Log.Println("Replaced strings - 146")
	}
	if id == 99 {
		filename = "99.txt"
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			Log.Print(err)
		}
		s = strings.Replace(string(b), "\n", "<br>", -1)
		//s = string(b)
	}
	return s
}
