package repository

import (
	"regexp"
	"fmt"
	"testing"
	"tools-ecg-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-openapi/testify/v2/require"
)

func TestCreateShiftAssignment(t *testing.T) {
	exampleAssignment := models.ShiftAssignment{
		ShiftId: 		1,
		Name: 			"test",
		PhoneNumber: 	"test",
	}

	exampleId := 1
	testErr := fmt.Errorf("error occured")

	tests := []struct{
		name 	string
		want	int
		fields	models.ShiftAssignment
		before	func (mock sqlmock.Sqlmock) sqlmock.Sqlmock
		wantErr	bool
	}{
		{
			name: 	"sucessfully create assignment",
			want: 	exampleId,
			fields: exampleAssignment,
			before: func(mock sqlmock.Sqlmock) sqlmock.Sqlmock {
				mock.ExpectQuery(regexp.QuoteMeta(createShiftAssignmentQuery)).
					WithArgs(exampleAssignment.ShiftId, exampleAssignment.Name, exampleAssignment.PhoneNumber).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(exampleId))
				return mock
			},
			wantErr: false,
		},
		{
			name: 	"sucessfully create assignment",
			want: 	0,
			fields: exampleAssignment,
			before: func(mock sqlmock.Sqlmock) sqlmock.Sqlmock {
				mock.ExpectQuery(regexp.QuoteMeta(createShiftAssignmentQuery)).
					WithArgs(exampleAssignment.ShiftId, exampleAssignment.Name, exampleAssignment.PhoneNumber).
					WillReturnError(testErr)
				return mock
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock = tt.before(mock)

		standplanRepo := NewStandplanRepository(db)

		result, err := standplanRepo.CreateShiftAssignment(tt.fields)

		if err != nil && !tt.wantErr {
			t.Errorf("CreateShiftAssignment() error = %v, wantErr = %t", err, tt.wantErr)
		}

		if mock.ExpectationsWereMet() != nil {
			t.Errorf("CreateShiftAssignment() error = %v", err)
		}

		require.Equal(t, tt.want, result, "GetWaffelnTotalCount() result = %v, want = %v", result, tt.want) 
	}

}