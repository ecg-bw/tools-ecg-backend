package service

import (
	"fmt"
	"testing"
	"time"
	"tools-ecg-backend/internal/mocks"
	"tools-ecg-backend/internal/models"

	"github.com/go-openapi/testify/v2/require"
	"go.uber.org/mock/gomock"
)

func TestGetStandplan(t *testing.T) {
	exampleStandplan := []models.Day{
		{
			Id:			1,
			Weekday: 	"",
			Date: 		time.Now(),
			Shifts:		[]models.Shift{
				{
					Id:				1,
					DayId:			1,
					Title:			"test",
					StartTime:		time.Now(),
					EndTime:		time.Now(),
					RequiredSlots:	3,
					Description:	nil,
					Assignments:	&[]models.ShiftAssignment{
						{
							Id: 			1,
							ShiftId:		1,
							Name: 			"test",
							PhoneNumber:	"001",
						},
					},
				},
			},	
		},
	}

	failedError := fmt.Errorf("error occured")

	tests := []struct{
		name 	string
		want	[]models.Day
		before	func (mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository
		wantErr	bool
	}{
		{
			name:	"Sucess/get empty Standplan",
			want:	[]models.Day{},
			before:	func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().GetAllDays().Return([]models.Day{}, nil)
				return mock
			},
			wantErr:	false,
		},
		{
			name:	"Sucess/get full Standplan",
			want:	exampleStandplan,
			before: func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().GetAllDays().Return(exampleStandplan, nil)
				mock.EXPECT().GetAllShiftsByDayId(exampleStandplan[len(exampleStandplan)-1].Id).Return(exampleStandplan[len(exampleStandplan)-1].Shifts, nil)
				mock.EXPECT().GetAllAssignmentsByShiftId(exampleStandplan[len(exampleStandplan)-1].Shifts[len(exampleStandplan[len(exampleStandplan)-1].Shifts)-1].Id).Return(*exampleStandplan[len(exampleStandplan)-1].Shifts[len(exampleStandplan[len(exampleStandplan)-1].Shifts)-1].Assignments, nil)
				return mock
			},
			wantErr: false,
		},
		{
			name:	"Failed/error for getting days",
			want:	nil,
			before: func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().GetAllDays().Return(nil, failedError)
				return mock
			},
			wantErr: true,
		},
		{
			name:	"Failed/error for getting shifts",
			want:	nil,
			before: func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().GetAllDays().Return(exampleStandplan, nil)
				mock.EXPECT().GetAllShiftsByDayId(exampleStandplan[len(exampleStandplan)-1].Id).Return(nil, failedError)
				return mock
			},
			wantErr: true,
		},
		{
			name:	"Failed/error for getting assignments",
			want:	nil,
			before: func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().GetAllDays().Return(exampleStandplan, nil)
				mock.EXPECT().GetAllShiftsByDayId(exampleStandplan[len(exampleStandplan)-1].Id).Return(exampleStandplan[len(exampleStandplan)-1].Shifts, nil)
				mock.EXPECT().GetAllAssignmentsByShiftId(exampleStandplan[len(exampleStandplan)-1].Shifts[len(exampleStandplan[len(exampleStandplan)-1].Shifts)-1].Id).Return(nil, failedError)
				return mock
			},
			wantErr: true,
		},
		
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockWaffelnRepo := mocks.NewMockWaffelnRepository(ctrl)
			mockStandplanRepo := mocks.NewMockStandplanRepository(ctrl)

			mockStandplanRepo = tt.before(mockStandplanRepo)

			standplanSvc := NewStandplanService(mockStandplanRepo, mockWaffelnRepo)

			result, err := standplanSvc.GetStandplan()

			if err != nil && !tt.wantErr {
				t.Errorf("GetWaffelnTotalCount() error = %v, wantErr = %t", err, tt.wantErr)
			}

			require.Equal(t, tt.want, result, "GetWaffelnTotalCount() result = %v, want = %v", result, tt.want) 
		})
	}
}

func TestCreateShiftAssignment(t *testing.T) {
	exampleAssignment := models.ShiftAssignment{
		ShiftId: 		1,
		Name: 			"test",
		PhoneNumber: 	"test",
	}

	exampleId := 1
	failedErr := fmt.Errorf("error occured")

	tests := []struct{
		name 	string
		fields	models.ShiftAssignment
		want	int
		before	func (mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository
		wantErr	bool
	}{
		{
			name: "Sucess/create sucessfull",
			fields: exampleAssignment,
			want: exampleId,
			before: func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().GetShiftById(exampleAssignment.ShiftId).Return(&models.Shift{RequiredSlots: 3}, nil)
				mock.EXPECT().GetShiftAssignmentsByShiftId(exampleAssignment.ShiftId).Return([]models.ShiftAssignment{}, nil)
				mock.EXPECT().CreateShiftAssignment(exampleAssignment).Return(exampleId, nil)
				return mock
			},
			wantErr: false,
		},
		{
			name: "Failed/error on getting shift",
			fields: exampleAssignment,
			want: 0,
			before: func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().GetShiftById(exampleAssignment.ShiftId).Return(nil, failedErr)
				return mock
			},
			wantErr: true,
		},
		{
			name: "Failed/shift is not existing",
			fields: exampleAssignment,
			want: 0,
			before: func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().GetShiftById(exampleAssignment.ShiftId).Return(nil, nil)
				return mock
			},
			wantErr: true,
		},
		{
			name: "Failed/error for getting assignments",
			fields: exampleAssignment,
			want: 0,
			before: func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().GetShiftById(exampleAssignment.ShiftId).Return(&models.Shift{RequiredSlots: 3}, nil)
				mock.EXPECT().GetShiftAssignmentsByShiftId(exampleAssignment.ShiftId).Return(nil, failedErr)
				return mock
			},
			wantErr: true,
		},
		{
			name: "Failed/shift is already full",
			fields: exampleAssignment,
			want: 0,
			before: func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().GetShiftById(exampleAssignment.ShiftId).Return(&models.Shift{RequiredSlots: 1}, nil)
				mock.EXPECT().GetShiftAssignmentsByShiftId(exampleAssignment.ShiftId).Return([]models.ShiftAssignment{{}}, nil)
				return mock
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockWaffelnRepo := mocks.NewMockWaffelnRepository(ctrl)
			mockStandplanRepo := mocks.NewMockStandplanRepository(ctrl)

			mockStandplanRepo = tt.before(mockStandplanRepo)

			standplanSvc := NewStandplanService(mockStandplanRepo, mockWaffelnRepo)

			result, err := standplanSvc.CreateShiftAssignment(tt.fields)

			if err != nil && !tt.wantErr {
				t.Errorf("GetWaffelnTotalCount() error = %v, wantErr = %t", err, tt.wantErr)
			}

			require.Equal(t, tt.want, result, "GetWaffelnTotalCount() result = %v, want = %v", result, tt.want) 
		})
	}
}

func TestUpdateShiftAssignment(t *testing.T) {
	exampleAssignment := models.ShiftAssignment{
		ShiftId: 		1,
		Name: 			"test",
		PhoneNumber: 	"test",
	}

	failedErr := fmt.Errorf("error occured")

	tests := []struct{
		name 	string
		fields	models.ShiftAssignment
		before	func (mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository
		wantErr	bool
	}{
		{
			name:	"sucessfully update assignment",
			fields: exampleAssignment,
			before: func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().UpdateShiftAssignment(exampleAssignment).Return(nil)
				return mock
			},
			wantErr: false,
		},
		{
			name:	"error on update shift assignment",
			fields: exampleAssignment,
			before: func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().UpdateShiftAssignment(exampleAssignment).Return(failedErr)
				return mock
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockWaffelnRepo := mocks.NewMockWaffelnRepository(ctrl)
			mockStandplanRepo := mocks.NewMockStandplanRepository(ctrl)

			mockStandplanRepo = tt.before(mockStandplanRepo)

			standplanSvc := NewStandplanService(mockStandplanRepo, mockWaffelnRepo)

			err := standplanSvc.UpdateShiftAssignment(tt.fields)

			if err != nil && !tt.wantErr {
				t.Errorf("GetWaffelnTotalCount() error = %v, wantErr = %t", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteShiftAssignment(t *testing.T) {
	exampleId := 1

	failedErr := fmt.Errorf("error occured")

	tests := []struct{
		name 	string
		fields	int
		before	func (mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository
		wantErr	bool
	}{
		{
			name:	"sucessfully delete assignment",
			fields: exampleId,
			before: func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().DeleteShiftAssignment(exampleId).Return(nil)
				return mock
			},
			wantErr: false,
		},
		{
			name:	"error on delete shift assignment",
			fields: exampleId,
			before: func(mock *mocks.MockStandplanRepository) *mocks.MockStandplanRepository {
				mock.EXPECT().DeleteShiftAssignment(exampleId).Return(failedErr)
				return mock
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockWaffelnRepo := mocks.NewMockWaffelnRepository(ctrl)
			mockStandplanRepo := mocks.NewMockStandplanRepository(ctrl)

			mockStandplanRepo = tt.before(mockStandplanRepo)

			standplanSvc := NewStandplanService(mockStandplanRepo, mockWaffelnRepo)

			err := standplanSvc.DeleteShiftAssignment(tt.fields)

			if err != nil && !tt.wantErr {
				t.Errorf("GetWaffelnTotalCount() error = %v, wantErr = %t", err, tt.wantErr)
			}
		})
	}
}

func TestGetWaffelnCount(t *testing.T) {
	totalCount := 13
	failedError := fmt.Errorf("error occured")

	tests := []struct {
		name	string
		want	int
		before	func (mock *mocks.MockWaffelnRepository) *mocks.MockWaffelnRepository
		wantErr	bool
	}{
		{
			name: "Should return correct total count",
			want: totalCount,
			before:	func(mock *mocks.MockWaffelnRepository) *mocks.MockWaffelnRepository {
				mock.EXPECT().GetWaffelnTotalCount().Return(totalCount, nil)
				return mock
			},
			wantErr: false,
		},
		{
			name: "Should return error",
			want: 0,
			before: func(mock *mocks.MockWaffelnRepository) *mocks.MockWaffelnRepository {
				mock.EXPECT().GetWaffelnTotalCount().Return(0, failedError)
				return mock
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockWaffelnRepo := mocks.NewMockWaffelnRepository(ctrl)
			mockStandplanRepo := mocks.NewMockStandplanRepository(ctrl)

			mockWaffelnRepo = tt.before(mockWaffelnRepo)

			standplanSvc := NewStandplanService(mockStandplanRepo, mockWaffelnRepo)

			result, err := standplanSvc.GetWaffelnTotalCount()

			if err != nil && !tt.wantErr {
				t.Errorf("GetWaffelnTotalCount() error = %v, wantErr = %t", err, tt.wantErr)
			}

			if result != tt.want {
				t.Errorf("GetWaffelnTotalCount() result = %v, want = %v", result, tt.want)
			}
		})
	}
}

func TestCreateWaffel(t *testing.T) {
	testId := 1
	createdAt := time.Now()
	testWaffel := models.Waffel {
				CreatedAt: createdAt,
			}
	testErr := fmt.Errorf("error occured")

	tests := []struct {
		name	string
		want	int
		before	func(mockWaffelnRepo *mocks.MockWaffelnRepository, mockStandplanRepo *mocks.MockStandplanRepository) (*mocks.MockWaffelnRepository, *mocks.MockStandplanRepository)
		wantErr	bool
	}{
		{
			name: "sucessfully",
			want: testId,
			before: func(mockWaffelnRepo *mocks.MockWaffelnRepository, mockStandplanRepo *mocks.MockStandplanRepository) (*mocks.MockWaffelnRepository, *mocks.MockStandplanRepository) {
				mockStandplanRepo.EXPECT().GetShiftByTime(createdAt).Return(testId, nil)
				
				updatedWaffel := testWaffel
				updatedWaffel.ShiftId = testId
				mockWaffelnRepo.EXPECT().CreateWaffel(updatedWaffel).Return(testId, nil)
				
				return mockWaffelnRepo, mockStandplanRepo
			},
			wantErr: false,
		},
		{
			name: "failed to get shiftId",
			want: 0,
			before: func(mockWaffelnRepo *mocks.MockWaffelnRepository, mockStandplanRepo *mocks.MockStandplanRepository) (*mocks.MockWaffelnRepository, *mocks.MockStandplanRepository) {
				mockStandplanRepo.EXPECT().GetShiftByTime(createdAt).Return(0, testErr)
				
				return mockWaffelnRepo, mockStandplanRepo
			},
			wantErr: true,
		},
		{
			name: "failed to create Waffel",
			want: 0,
			before: func(mockWaffelnRepo *mocks.MockWaffelnRepository, mockStandplanRepo *mocks.MockStandplanRepository) (*mocks.MockWaffelnRepository, *mocks.MockStandplanRepository) {
				mockStandplanRepo.EXPECT().GetShiftByTime(createdAt).Return(testId, nil)
				
				updatedWaffel := testWaffel
				updatedWaffel.ShiftId = testId
				mockWaffelnRepo.EXPECT().CreateWaffel(updatedWaffel).Return(0, testErr)
				
				return mockWaffelnRepo, mockStandplanRepo
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockWaffelnRepo := mocks.NewMockWaffelnRepository(ctrl)
			mockStandplanRepo := mocks.NewMockStandplanRepository(ctrl)

			mockWaffelnRepo, mockStandplanRepo = tt.before(mockWaffelnRepo, mockStandplanRepo)

			standplanSvc := NewStandplanService(mockStandplanRepo, mockWaffelnRepo)

			result, err := standplanSvc.CreateWaffel(testWaffel)

			if err != nil && !tt.wantErr {
				t.Errorf("GetWaffelnTotalCount() error = %v, wantErr = %t", err, tt.wantErr)
			}

			if result != tt.want {
				t.Errorf("CreateWaffel() result = %v, want = %v", result, tt.want)
			}
		})
	}
}