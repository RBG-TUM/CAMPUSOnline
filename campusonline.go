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
	Title    string          `json:"title"`
	CourseID int             `json:"course_id"`
	Events   []Event         `json:"events"`
	Contacts []ContactPerson `json:"contacts"`
}

type Event struct {
	Title    string    `json:"title"`
	RoomID   int       `json:"room_id"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	RoomName string    `json:"room_name"`
	Comment  string    `json:"comment"`
}

func (c *CampusOnline) GetCourseIdAndContacts(courseName string, semester string) (courseID int, contacts []ContactPerson, err error) {
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
			var c []ContactPerson
			for _, person := range course.Course.Contacts.Person {
				r := ""
				for _, role := range person.Role {
					if r != "" {
						r += ", "
					}
					r += role.Text
				}
				c = append(c, ContactPerson{FirstName: person.Name.Given, LastName: person.Name.Family, Email: person.ContactData.Email, Role: r})
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

type ContactPerson struct {
	FirstName string
	LastName  string
	Email     string
	Role      string
}
