package main

import (
	"fmt"
	campusonline "github.com/RBG-TUM/CAMPUSOnline"
)

func main()  {
	co, _ := campusonline.New("","")
	roomstuff, _ := co.GetScheduleForRoom(12325, "21W")
	fmt.Println(len(roomstuff.Courses))
}
