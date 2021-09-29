package campusonline

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const InOrgId = 14189
const xCalOrgDN = "xcal/organization/courses/xml?token=%s&timeMode=absolute&orgUnitID=%d&fromDate=%s&untilDate=%s"

//GetXCalOrgIN returns all events in the specified time stamp for the computer science organisation
func (c *CampusOnline) GetXCalOrgIN(from time.Time, until time.Time) (ICalendar, error) {
	url := baseURL + fmt.Sprintf(xCalOrgDN, c.token, InOrgId, from.Format("20060102"), until.Format("20060102"))
	println(url)
	resp, err := http.Get(url)
	if err != nil {
		return ICalendar{}, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ICalendar{}, err
	}
	var res ICalendar
	err = xml.Unmarshal(body, &res)
	if err != nil {
		return ICalendar{}, err
	}
	return res, nil
}

func (c *ICalendar) Sort() {
	sort.Sort(c.Vcalendar.Events)
}

func (c *ICalendar) Filter() {
	var newEvents []VEvent
	for _, event := range c.Vcalendar.Events {
		if event.Categories.Item == "Vorlesung" &&
			(event.Status == "fix" || event.Status == "geplant" )&&
			inRoomList(event.Location.Text) &&
			!strings.Contains(strings.ToLower(event.Comment), "videoübertragung aus") {
			newEvents = append(newEvents, event)
		}
	}
	c.Vcalendar.Events = newEvents
}

var roomList = []string{
	"5602.EG.001",  // HS1
	"5604.EG.011",  // HS2
	"5606.EG.011",  // HS3
	"5608.EG.038",  // 00.08.038 seminar room
	"5613.EG.009A", // 00.13.009A seminar room
	"5620.01.101",  // Interims I 101
	"5620.01.102",  // Interims I 102
	"5510.02.001",  // MW 2001 (Rudolf-Diesel-HS)
	"5510.EG.001",  // MW 0001 (Gustav-Niemann-HS)
}

func inRoomList(roomText string) bool {
	for _, s := range roomList {
		if strings.Contains(roomText, s) {
			return true
		}
	}
	return false
}

func (c *ICalendar) GroupByCourse() []Course {
	courses := map[string]*Course{}
	for _, event := range c.Vcalendar.Events {
		TOUrl := event.Description.Altrep
		splitUrl := strings.Split(TOUrl, "pStpSpNr=")
		if len(splitUrl) != 2 {
			continue
		}
		foundCourse, found := courses[splitUrl[1]]
		start, parseErr := time.Parse("20060102T150405", event.Dtstart)
		if parseErr != nil {
			continue
		}
		end, parseErr := time.Parse("20060102T150405", event.Dtend)
		if parseErr != nil {
			continue
		}
		if !found {
			cID, err := strconv.Atoi(splitUrl[1])
			if err != nil {
				continue
			}
			courses[splitUrl[1]] = &Course{
				Title:    event.Summary,
				CourseID: cID,
				Events: []Event{{
					Start:    start,
					End:      end,
					RoomName: event.Location.Text,
					Comment:  event.Comment,
				}},
				Contacts: nil,
			}
		} else {
			foundCourse.Events = append(foundCourse.Events, Event{
				Title:    "",
				Start:    start,
				End:      end,
				RoomName: event.Location.Text,
				Comment:  event.Comment,
			})
		}
	}
	var res []Course
	for _, course := range courses {
		res = append(res, *course)
	}
	return res
}

func (c CampusOnline) LoadCourseContacts(courses []Course) ([]Course, error) {
	for i := range courses {
		url := baseURL + fmt.Sprintf(courseExportDN, c.token, courses[i].CourseID)
		got, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		body, err := ioutil.ReadAll(got.Body)
		if err != nil {
			return nil, err
		}
		var res CDM
		err = xml.Unmarshal(body, &res)
		if err != nil {
			return nil, err
		}
		for _, person := range res.Course.Contacts.Person {
			pRole := ""
			for _, r := range person.Role {
				if pRole != "" {
					pRole += ", "
				}
				pRole += r.Text
			}
			courses[i].Contacts = append(courses[i].Contacts, ContactPerson{
				FirstName: person.Name.Given,
				LastName:  person.Name.Family,
				Email:     person.ContactData.Email,
				Role:      pRole,
			})
		}
	}
	return courses, nil
}
