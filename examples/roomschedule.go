package main

import (
	"fmt"
	campusonline "github.com/RBG-TUM/CAMPUSOnline"
	"time"
)

func main() {
	co, _ := campusonline.New("xxx", "xxx")
	roomstuff, _ := co.GetXCalOrgIN(
		time.Date(2021, 10, 1, 0, 0, 0, 0, time.Local),
		time.Date(2022, 3, 31, 23, 59, 59, 0, time.Local),
	)
	ical := &roomstuff
	fmt.Println(len(roomstuff.Vcalendar.Events))
	ical.Filter()
	ical.Sort()
	fmt.Println(len(roomstuff.Vcalendar.Events))
	courses := ical.GroupByCourse()
	println(len(courses))
	for _, course := range courses {
		println(course.Title, ":", course.CourseID)
		for _, event := range course.Events {
			println(event.Start.Format("2006-01-02 15:04"), "\t ", event.RoomName)
		}
	}
}
