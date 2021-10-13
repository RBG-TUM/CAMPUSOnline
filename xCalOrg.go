package campusonline

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const InOrgId = 14189
const MaOrgID = 14178
const PhOrgID = 14179
const xCalOrgDN = "xcal/organization/courses/xml?token=%s&timeMode=absolute&orgUnitID=%d&fromDate=%s&untilDate=%s"

//GetXCalOrgIN returns all events in the specified time stamp for the computer science organisation
func (c *CampusOnline) GetXCalOrgIN(from time.Time, until time.Time) (ICalendar, error) {
	url := baseURL + fmt.Sprintf(xCalOrgDN, c.token, PhOrgID, from.Format("20060102"), until.Format("20060102"))
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
	re, _ := regexp.Compile("^[0-9]+")
	for _, event := range c.Vcalendar.Events {
		if strings.Contains(event.Summary, "Praktikum Systemadministration") {
			event.Categories.Item = "Vorlesung"
			event.Location.Text = "5620.01.102 (102, Hörsaal 2, \"Interims I\"), Boltzmannstr. 5(5620), 85748 Garching b. München"
		}
		if strings.Contains(strings.ToLower(event.Categories.Item), "vorlesung") &&
			(event.Status == "fix" || event.Status == "geplant") &&
			inRoomList(event.Location.Text) &&
			!strings.Contains(strings.ToLower(event.Comment), "videoübertragung aus") {
			// remove prepending digits
			event.Summary = strings.TrimSpace(re.ReplaceAllString(event.Summary, ""))

			// Replace with readable locations
			newLocation := event.Location.Text
			for k, e := range roomList {
				if strings.Contains(newLocation, k) {
					event.Location.Text = e
					break
				}
			}
			newEvents = append(newEvents, event)
		}
	}
	c.Vcalendar.Events = newEvents
}

var roomList = map[string]string{
	"5602.EG.001":  "MI HS1",
	"5604.EG.011":  "MI HS2",
	"5606.EG.011":  "MI HS3",
	"5608.EG.038":  "00.08.038",
	"5613.EG.009A": "00.13.009A",
	"5620.01.101":  "Interims I 101",
	"5620.01.102":  "Interims I 102",
	"5510.02.001":  "MW 2001",
	"5510.EG.001":  "MW 0001",
}

func inRoomList(roomText string) bool {
	for s := range roomList {
		if strings.Contains(roomText, s) {
			return true
		}
	}
	return false
}

func (c *ICalendar) GroupByCourse() []Course {
	slugCount := map[string]int{} // keep track of slug dupes
	courses := map[string]*Course{}
	for _, event := range c.Vcalendar.Events {
		TOUrl := event.Description.Altrep
		splitUrl := strings.Split(TOUrl, "pStpSpNr=")
		if len(splitUrl) != 2 {
			continue
		}
		foundCourse, found := courses[splitUrl[1]]
		start, parseErr := time.ParseInLocation("20060102T150405", event.Dtstart, time.Local)
		if parseErr != nil {
			continue
		}
		end, parseErr := time.ParseInLocation("20060102T150405", event.Dtend, time.Local)
		if parseErr != nil {
			continue
		}
		if !found {
			cID, err := strconv.Atoi(splitUrl[1])
			if err != nil {
				continue
			}

			// take care of slug dupes
			slug := generateCourseSlug(event.Summary)
			count, found := slugCount[slug]
			if found {
				slugCount[slug]++
				slug += fmt.Sprintf("%d", count)
			} else {
				slugCount[slug] = 1
			}

			// create course with event
			courses[splitUrl[1]] = &Course{
				Title:    event.Summary,
				Slug:     slug,
				CourseID: cID,
				Import:   false,
				Events: []Event{{
					Start:    start,
					End:      end,
					RoomName: event.Location.Text,
					Comment:  event.Comment,
					Import:   true,
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
				Import:   true,
			})
		}
	}
	var res []Course
	for _, course := range courses {
		res = append(res, *course)
	}
	return res
}

func generateCourseSlug(title string) string {
	courseSlug := ""
	for _, l := range strings.Split(title, " ") {
		runes := []rune(l)
		if len(runes) != 0 && (unicode.IsNumber(runes[0]) || unicode.IsLetter(runes[0])) {
			courseSlug += string(runes[0])
		}
	}
	return courseSlug
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
		hasMainContact := false
		for _, person := range res.Course.Contacts.Person {
			isMainContact := false
			pRole := ""
			for _, r := range person.Role {
				if pRole != "" {
					pRole += ", "
				}
				pRole += r.Text
			}
			if !hasMainContact && (strings.Contains(strings.ToLower(pRole), "leiter") || strings.Contains(strings.ToLower(pRole), "prüfer")) {
				isMainContact = true
				hasMainContact = true
			}
			courses[i].Contacts = append(courses[i].Contacts, ContactPerson{
				FirstName:   person.Name.Given,
				LastName:    person.Name.Family,
				Email:       person.ContactData.Email,
				Role:        pRole,
				MainContact: isMainContact,
			})
		}
		if !hasMainContact && len(courses[i].Contacts) != 0 {
			courses[i].Contacts[0].MainContact = true
		}
	}
	return courses, nil
}
