package models

import "time"

type ShiftAssignment struct {
	Id          int
	ShiftId     int
	Name        string
	PhoneNumber string
}

type Shift struct {
	Id 		int
	DayId 	int
	Title 	string
	StartTime 	time.Time
	EndTime 	time.Time
	RequiredSlots int
	Description *string
	Assignments *[]ShiftAssignment
}

type Day struct {
	Id int
	Weekday string
	Date time.Time
	Shifts []Shift
}
