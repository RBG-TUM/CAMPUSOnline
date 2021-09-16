package campusonline

import (
	"encoding/xml"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"
)

const baseURL = "https://campus.tum.de/tumonlinej/ws/webservice_v1.0/"
const roomDN = "/rdm/room/schedule/xml?token=%s&timeMode=absolute&roomID=%d&buildingCode=&fromDate=%s&untilDate=%s"

type CampusOnline struct {
	token string
	cache *ristretto.Cache
}

func New(token string) (*CampusOnline, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}
	return &CampusOnline{token: token, cache: cache}, nil
}

type Room struct {
	Courses []Course `json:"courses"`
}

type Course struct {
	Title       string  `json:"title"`
	CourseID    int     `json:"course_id"`
	ContactName string  `json:"contact_name"`
	ContactMail string  `json:"contact_mail"`
	Events      []Event `json:"events"`
}

type Event struct {
	Title  string    `json:"title"`
	RoomID int       `json:"room_id"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

func (c *CampusOnline) GetScheduleForRoom(roomID int) (*Room, error) {
	query := baseURL + fmt.Sprintf(roomDN, c.token, roomID, time.Now().Format("20060102"), time.Now().Add(time.Hour*24*7*30*5).Format("20060102"))
	var httpRes []byte
	if cached, found := c.cache.Get(query); found {
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
		c.cache.SetWithTTL(query, resBody, 1, time.Minute)
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
