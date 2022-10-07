package campusonline

import (
	"encoding/xml"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	baseURL        = "https://campus.tum.de/tumonlinej/ws/webservice_v1.0/"
	basicBaseURL   = "https://campus.tum.de/tumonline/wbservicesbasic."
	roomDN         = "/rdm/room/schedule/xml?token=%s&timeMode=absolute&roomID=%d&buildingCode=&fromDate=%s&untilDate=%s"
	courseSearchDN = "veranstaltungenSuche?pToken=%s&pSuche=%s&pSemester=%s"
	courseExportDN = "/cdm/course/xml?token=%s&courseID=%d"
)

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
	Slug     string          `json:"slug"`
	CourseID int             `json:"course_id"`
	Events   []Event         `json:"events"`
	Contacts []ContactPerson `json:"contacts"`
	Import   bool            `json:"import"`
}

type Event struct {
	Title    string    `json:"title"`
	RoomID   int       `json:"room_id"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	RoomName string    `json:"room_name"`
	Comment  string    `json:"comment"`
	Import   bool      `json:"import"`
	EventID  string    `json:"event_id"`
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
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	MainContact bool   `json:"main_contact"`
}
