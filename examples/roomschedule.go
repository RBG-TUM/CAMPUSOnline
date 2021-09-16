package main

import campusonline "github.com/RBG-TUM/CAMPUSOnline"

func main()  {
	co, _ := campusonline.New("xxx")
	println("# HS1")
	_, _ = co.GetScheduleForRoom(12325)
}