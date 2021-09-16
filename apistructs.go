package campusonline

import "encoding/xml"

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
