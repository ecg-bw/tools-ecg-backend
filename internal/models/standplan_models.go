package models

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
	StartTime 	string
	EndTime 	string
	RequiredSlots int
	Description *string
}