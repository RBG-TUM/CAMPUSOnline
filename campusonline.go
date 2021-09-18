package campusonline

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://campus.tum.de/tumonlinej/ws/webservice_v1.0/"
const basicBaseURL = "https://campus.tum.de/tumonline/wbservicesbasic."
const roomDN = "/rdm/room/schedule/xml?token=%s&timeMode=absolute&roomID=%d&buildingCode=&fromDate=%s&untilDate=%s"
const courseSearchDN = "veranstaltungenSuche?pToken=%s&pSuche=%s&pSemester=%s"
const courseExportDN = "/cdm/course/xml?token=%s&courseID=%d"

type CampusOnline struct {
	token      string
	basicToken string
	cache      *ristretto.Cache
}

func New(token string, basicToken string) (*CampusOnline, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}
	return &CampusOnline{token: token, basicToken: basicToken, cache: cache}, nil
}

type Room struct {
	Courses []Course `json:"courses"`
}

type Course struct {
	Title       string    `json:"title"`
	CourseID    int       `json:"course_id"`
	ContactName string    `json:"contact_name"`
	ContactMail string    `json:"contact_mail"`
	Events      []Event   `json:"events"`
	Contacts    []Contact `json:"contacts"`
}

type Event struct {
	Title  string    `json:"title"`
	RoomID int       `json:"room_id"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

func (c *CampusOnline) GetScheduleForRoom(roomID int, semester string) (*Room, error) {
	query := baseURL + fmt.Sprintf(roomDN, c.token, roomID, time.Now().Format("20060102"), time.Now().Add(time.Hour*24*7*30*5).Format("20060102"))
	var httpRes []byte
	cacheKey := fmt.Sprintf("%s%d", "roomschedule", roomID)
	if cached, found := c.cache.Get(cacheKey); found {
		httpRes = cached.([]byte)
	} else {
		res, err := http.Get(query)
		if err != nil {
			return nil, err
		}
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		c.cache.SetWithTTL(cacheKey, resBody, 1, time.Minute)
		httpRes = resBody
	}
	var res RDM
	err := xml.Unmarshal(httpRes, &res)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	filtered := filterAttrRDM(res, "eventTypeID", "LV")      // only lehrveranstaltungen
	filtered2 := filterAttrRDM(filtered, "courseType", "VO") // only vorlesungen
	filtered3 := filterAttrRDM(filtered2, "status", "fix")   // nothing deleted or moved
	grouped := groupEventsByName(filtered3.Resource.Content.ResourceGroup.Content.Events)
	roomResult := Room{}
	for courseName, events := range grouped {
		sort.Slice(events, func(i, j int) bool {
			start1, _ := getResourceAttrVal(events[i], "dtstart")
			parsedTime1, err := time.Parse("20060102T150405", start1)
			if err != nil {
				return false
			}
			start2, _ := getResourceAttrVal(events[j], "dtstart")
			parsedTime2, err := time.Parse("20060102T150405", start2)
			if err != nil {
				return false
			}
			return parsedTime1.Before(parsedTime2)
		})
		course := Course{Title: courseName}
		courseID, contacts, err := c.GetCourseIdAndContacts(courseName, semester)
		if err != nil {
			log.Println(err)
		} else {
			course.CourseID = courseID
			course.Contacts = contacts
		}

		for _, event := range events {
			startStr, found := getResourceAttrVal(event, "dtstart")
			if !found {
				continue
			}
			start, err := time.Parse("20060102T150405", startStr)
			if err != nil {
				continue
			}
			endStr, found := getResourceAttrVal(event, "dtend")
			if !found {
				continue
			}
			end, err := time.Parse("20060102T150405", endStr)
			if err != nil {
				continue
			}
			course.Events = append(course.Events, Event{
				Title:  "",
				Start:  start,
				End:    end,
				RoomID: roomID,
			})
		}
		roomResult.Courses = append(roomResult.Courses, course)
	}
	return &roomResult, nil
}

func groupEventsByName(events []CalendarEvent) map[string][]CalendarEvent {
	res := make(map[string][]CalendarEvent)
	for _, event := range events {
		title, found := getResourceAttrVal(event, "eventTitle")
		if !found {
			continue
		}
		res[title] = append(res[title], event)
	}
	return res
}

func (c *CampusOnline) GetCourseIdAndContacts(courseName string, semester string) (courseID int, contacts []Contact, err error) {
	//Get courseID through course search:
	re, err := regexp.Compile("(\\(.*\\))|(\\[.*])") // Einführung in die Informatik 1 [IN0001] -> Einführung in die Informatik 1
	if err != nil {
		return 0, contacts, err
	}
	courseNameOrig := courseName // save original course name for later verification
	courseName = string(re.ReplaceAll([]byte(courseName), []byte("")))
	courseName = strings.TrimSpace(courseName)
	searchUrl := basicBaseURL + fmt.Sprintf(courseSearchDN, c.basicToken, url.PathEscape(courseName), semester)
	log.Println(searchUrl)
	httpResponse, err := http.Get(searchUrl)
	if err != nil {
		return 0, contacts, err
	}
	respBody, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return 0, contacts, err
	}
	var response Rowset
	err = xml.Unmarshal(respBody, &response)
	if err != nil {
		log.Println(err)
		return 0, contacts, err
	}
	for _, s := range response.Row {
		if (s.StpLvArtKurz == "PR" || s.StpLvArtKurz == "VO" || s.StpLvArtKurz == "VI") && s.VortragendeMitwirkende.Text != "" { // only Lecture that contains persons
			courseID, err := strconv.Atoi(s.StpSpNr)
			if err != nil {
				return 0, contacts, err
			}
			course, err := c.exportCourseByID(courseID)
			if err != nil {
				log.Println(err)
				continue
			}
			if course.Course.CourseName.Text != courseNameOrig {
				continue
			}

			//Get course contacts via courseID
			var c []Contact
			for _, person := range course.Course.Contacts.Person {
				r := ""
				for _, role := range person.Role {
					if r != "" {
						r += ", "
					}
					r += role.Text
				}
				c = append(c, Contact{FirstName: person.Name.Given, LastName: person.Name.Family, Email: person.ContactData.Email, Role: r})
			}
			return courseID, c, nil
		}
	}
	return 0, contacts, errors.New("can't find courseID for course")
}

func (c *CampusOnline) exportCourseByID(id int) (CDM, error) {
	url := baseURL + fmt.Sprintf(courseExportDN, c.token, id)
	response, err := http.Get(url)
	if err != nil {
		return CDM{}, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return CDM{}, err
	}
	var result CDM
	err = xml.Unmarshal(body, &result)
	if err != nil {
		return CDM{}, err
	}
	return result, nil
}

type Contact struct {
	FirstName string
	LastName  string
	Email     string
	Role      string
}
