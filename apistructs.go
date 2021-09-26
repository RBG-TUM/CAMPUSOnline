package campusonline

import (
	"encoding/xml"
	"strings"
)

// RDM is a struct that can unmarshal tum online rdm replies
type RDM struct {
	XMLName        xml.Name `xml:"RDM"`
	Text           string   `xml:",chardata"`
	Cor            string   `xml:"cor,attr"`
	Xsi            string   `xml:"xsi,attr"`
	SchemaLocation string   `xml:"schemaLocation,attr"`
	Resource       struct {
		Text    string `xml:",chardata"`
		TypeID  string `xml:"typeID,attr"`
		Content struct {
			Text      string `xml:",chardata"`
			Attribute []struct {
				Text          string `xml:",chardata"`
				AttrID        string `xml:"attrID,attr"`
				AttrDataType  string `xml:"attrDataType,attr"`
				CharacterData string `xml:"characterData"`
			} `xml:"attribute"`
			ResourceGroup struct {
				Text    string `xml:",chardata"`
				TypeID  string `xml:"typeID,attr"`
				Content struct {
					Text   string          `xml:",chardata"`
					Events []CalendarEvent `xml:"resource"` // this is the actual calendar entry list
				} `xml:"description"`
			} `xml:"resourceGroup"`
		} `xml:"description"`
	} `xml:"resource"`
}

type CalendarEvent struct {
	Text        string `xml:",chardata"`
	TypeID      string `xml:"typeID,attr"`
	Description struct {
		Text       string `xml:",chardata"`
		Attributes []struct {
			Text         string `xml:",chardata"`
			AttrID       string `xml:"attrID,attr"`
			AttrDataType string `xml:"attrDataType,attr"`
		} `xml:"attribute"`
	} `xml:"description"`
}

func getResourceAttrVal(c CalendarEvent, key string) (val string, found bool) {
	for _, attribute := range c.Description.Attributes {
		if key == attribute.AttrID {
			return attribute.Text, true
		}
	}
	return "", false
}

func filterAttrRDM(rdm RDM, key string, typeID string) RDM {
	rdmRet := rdm
	rdmRet.Resource.Content.ResourceGroup.Content.Events = []CalendarEvent{}
	for _, event := range rdm.Resource.Content.ResourceGroup.Content.Events {
		if val, found := getResourceAttrVal(event, key); found && val == typeID { // A -> Abhaltung
			rdmRet.Resource.Content.ResourceGroup.Content.Events = append(rdmRet.Resource.Content.ResourceGroup.Content.Events, event)
		}
	}
	return rdmRet
}

//Rowset Is a struct representing a response from the "veranstaltungenSuche" api
type Rowset struct {
	XMLName xml.Name `xml:"rowset"`
	Text    string   `xml:",chardata"`
	Row     []struct {
		Text                   string `xml:",chardata"`
		StpSpNr                string `xml:"stp_sp_nr"`
		StpLvNr                string `xml:"stp_lv_nr"`
		StpSpTitel             string `xml:"stp_sp_titel"`
		DauerInfo              string `xml:"dauer_info"`
		StpSpSst               string `xml:"stp_sp_sst"`
		StpLvArtName           string `xml:"stp_lv_art_name"`
		StpLvArtKurz           string `xml:"stp_lv_art_kurz"`
		SjName                 string `xml:"sj_name"`
		Semester               string `xml:"semester"`
		SemesterName           string `xml:"semester_name"`
		SemesterID             string `xml:"semester_id"`
		OrgNrBetreut           string `xml:"org_nr_betreut"`
		OrgNameBetreut         string `xml:"org_name_betreut"`
		OrgKennungBetreut      string `xml:"org_kennung_betreut"`
		VortragendeMitwirkende struct {
			Text   string `xml:",chardata"`
			Isnull string `xml:"isnull,attr"`
		} `xml:"vortragende_mitwirkende"`
	} `xml:"row"`
}

//CDM Is a struct representing a response from the course export api
type CDM struct {
	XMLName                   xml.Name `xml:"CDM"`
	Text                      string   `xml:",chardata"`
	Xsi                       string   `xml:"xsi,attr"`
	NoNamespaceSchemaLocation string   `xml:"noNamespaceSchemaLocation,attr"`
	Language                  string   `xml:"language,attr"`
	Properties                struct {
		Text       string `xml:",chardata"`
		Datasource string `xml:"datasource"`
		Datetime   struct {
			Text string `xml:",chardata"`
			Date string `xml:"date,attr"`
			Time string `xml:"time,attr"`
		} `xml:"datetime"`
	} `xml:"properties"`
	Course struct {
		Text       string `xml:",chardata"`
		Language   string `xml:"language,attr"`
		TypeID     string `xml:"typeID,attr"`
		TypeName   string `xml:"typeName,attr"`
		CourseID   string `xml:"courseID"`
		CourseName struct {
			Chardata string `xml:",chardata"`
			Text     string `xml:"text"`
		} `xml:"courseName"`
		CourseCode        string `xml:"courseCode"`
		CourseDescription string `xml:"courseDescription"`
		Level             struct {
			Text    string `xml:",chardata"`
			WebLink struct {
				Text        string `xml:",chardata"`
				UserDefined string `xml:"userDefined,attr"`
				Href        string `xml:"href"`
			} `xml:"webLink"`
		} `xml:"level"`
		TeachingTerm string `xml:"teachingTerm"`
		Credits      struct {
			Text         string `xml:",chardata"`
			HoursPerWeek string `xml:"hoursPerWeek,attr"`
		} `xml:"credits"`
		LearningObjectives string `xml:"learningObjectives"`
		AdmissionInfo      struct {
			Text                 string `xml:",chardata"`
			AdmissionDescription struct {
				Text        string `xml:",chardata"`
				UserDefined string `xml:"userDefined,attr"`
				WebLink     struct {
					Text        string `xml:",chardata"`
					UserDefined string `xml:"userDefined,attr"`
					Href        string `xml:"href"`
				} `xml:"webLink"`
			} `xml:"admissionDescription"`
		} `xml:"admissionInfo"`
		InstructionLanguage struct {
			Text         string `xml:",chardata"`
			TeachingLang string `xml:"teachingLang,attr"`
		} `xml:"instructionLanguage"`
		Syllabus struct {
			Text     string `xml:",chardata"`
			SubBlock struct {
				Text        string `xml:",chardata"`
				UserDefined string `xml:"userDefined,attr"`
				WebLink     []struct {
					Text     string `xml:",chardata"`
					Href     string `xml:"href"`
					LinkName string `xml:"linkName"`
				} `xml:"webLink"`
			} `xml:"subBlock"`
		} `xml:"syllabus"`
		Exam struct {
			Text      string `xml:",chardata"`
			InfoBlock struct {
				Text    string `xml:",chardata"`
				WebLink struct {
					Text        string `xml:",chardata"`
					UserDefined string `xml:"userDefined,attr"`
					Href        string `xml:"href"`
				} `xml:"webLink"`
			} `xml:"infoBlock"`
		} `xml:"exam"`
		TeachingActivity struct {
			Text                 string `xml:",chardata"`
			TeachingActivityID   string `xml:"teachingActivityID"`
			TeachingActivityName struct {
				Chardata string `xml:",chardata"`
				Text     string `xml:"text"`
			} `xml:"teachingActivityName"`
			InfoBlock struct {
				Text    string `xml:",chardata"`
				WebLink struct {
					Text        string `xml:",chardata"`
					UserDefined string `xml:"userDefined,attr"`
					Href        string `xml:"href"`
				} `xml:"webLink"`
			} `xml:"infoBlock"`
		} `xml:"teachingActivity"`
		Contacts struct {
			Text   string `xml:",chardata"`
			Person []struct {
				Text     string `xml:",chardata"`
				PersonID string `xml:"personID"`
				Name     struct {
					Text   string `xml:",chardata"`
					Given  string `xml:"given"`
					Family string `xml:"family"`
				} `xml:"name"`
				Role []struct {
					Chardata string `xml:",chardata"`
					RoleID   string `xml:"roleID,attr"`
					Text     string `xml:"text"`
				} `xml:"role"`
				ContactData struct {
					Text        string `xml:",chardata"`
					ContactName struct {
						Chardata string `xml:",chardata"`
						Text     string `xml:"text"`
					} `xml:"contactName"`
					Adr struct {
						Text     string `xml:",chardata"`
						Extadr   string `xml:"extadr"`
						Street   string `xml:"street"`
						Locality string `xml:"locality"`
						Pcode    string `xml:"pcode"`
						Country  string `xml:"country"`
					} `xml:"adr"`
					VisitHour struct {
						Text   string `xml:",chardata"`
						Header string `xml:"header"`
					} `xml:"visitHour"`
					Telephone struct {
						Text    string `xml:",chardata"`
						Teltype string `xml:"teltype,attr"`
					} `xml:"telephone"`
					Fax     string `xml:"fax"`
					Email   string `xml:"email"`
					WebLink struct {
						Text string `xml:",chardata"`
						Href string `xml:"href"`
					} `xml:"webLink"`
				} `xml:"contactData"`
				InfoBlock struct {
					Text    string `xml:",chardata"`
					WebLink struct {
						Text        string `xml:",chardata"`
						UserDefined string `xml:"userDefined,attr"`
						Href        string `xml:"href"`
					} `xml:"webLink"`
					Picture struct {
						Text    string `xml:",chardata"`
						WebLink struct {
							Text        string `xml:",chardata"`
							UserDefined string `xml:"userDefined,attr"`
							Href        string `xml:"href"`
						} `xml:"webLink"`
					} `xml:"picture"`
					SubBlock []struct {
						Text        string   `xml:",chardata"`
						UserDefined string   `xml:"userDefined,attr"`
						SubBlock    []string `xml:"subBlock"`
					} `xml:"subBlock"`
				} `xml:"infoBlock"`
			} `xml:"person"`
		} `xml:"contacts"`
		InfoBlock struct {
			Text     string `xml:",chardata"`
			SubBlock struct {
				Text        string `xml:",chardata"`
				UserDefined string `xml:"userDefined,attr"`
				SubBlock    struct {
					Text        string `xml:",chardata"`
					UserDefined string `xml:"userDefined,attr"`
				} `xml:"subBlock"`
			} `xml:"subBlock"`
		} `xml:"infoBlock"`
	} `xml:"course"`
}

type ICalendar struct {
	XMLName        xml.Name `xml:"iCalendar"`
	Text           string   `xml:",chardata"`
	XCal           string   `xml:"xCal,attr"`
	Xsi            string   `xml:"xsi,attr"`
	SchemaLocation string   `xml:"schemaLocation,attr"`
	Vcalendar      struct {
		Text     string `xml:",chardata"`
		Calscale string `xml:"calscale,attr"`
		Method   string `xml:"method,attr"`
		Version  string `xml:"version,attr"`
		Prodid   string `xml:"prodid,attr"`
		Events   Events `xml:"vevent"`
	} `xml:"vcalendar"`
}

type VEvent struct {
	Text        string `xml:",chardata"`
	Uid         string `xml:"uid"`
	Dtstamp     string `xml:"dtstamp"`
	Dtstart     string `xml:"dtstart"`
	Dtend       string `xml:"dtend"`
	Duration    string `xml:"duration"`
	Summary     string `xml:"summary"`
	Description struct {
		Text   string `xml:",chardata"`
		Altrep string `xml:"altrep,attr"`
	} `xml:"description"`
	Location struct {
		Text   string `xml:",chardata"`
		Altrep string `xml:"altrep,attr"`
	} `xml:"location"`
	Status    string `xml:"status"`
	Organizer struct {
		Text string `xml:",chardata"`
		Cn   string `xml:"cn,attr"`
	} `xml:"organizer"`
	Attendee []struct {
		Text string `xml:",chardata"`
		Cn   string `xml:"cn,attr"`
	} `xml:"attendee"`
	Categories struct {
		Text string `xml:",chardata"`
		Item string `xml:"item"`
	} `xml:"categories"`
	Comment string `xml:"comment"`
}

type Events []VEvent

func (evt Events) Len() int { return len(evt) }

func (v Events) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

func (v Events) Less(i, j int) bool { return strings.Compare(v[i].Dtstart, v[j].Dtstart) < 0 }
