package store

import (
	"slices"
	"testing"
)

func TestApplyClubChanges(t *testing.T) {
	t.Parallel()
	schedule := []DayClub{
		{ChildID: 1, LastName: "Иванов", FirstName: "Петя", Start: "14:00", Name: "Шахматы"},
		{ChildID: 2, LastName: "Козлова", FirstName: "Аня", Start: "12:30", Name: "Танцы"},
		{ChildID: 2, LastName: "Козлова", FirstName: "Аня", Start: "14:00", Name: "Шахматы"},
		{ChildID: 1, LastName: "Иванов", FirstName: "Петя", Start: "16:00", Name: "Шахматы"},
	}

	tests := []struct {
		name    string
		changes []ClubChange
		want    []DayClub
	}{
		{
			name: "no changes keeps the schedule sorted",
			want: []DayClub{schedule[1], schedule[0], schedule[2], schedule[3]},
		},
		{
			name:    "cancel removes every time of the child's club only",
			changes: []ClubChange{{ChildID: 1, Name: "Шахматы"}},
			want:    []DayClub{schedule[1], schedule[2]},
		},
		{
			name:    "new time replaces every scheduled time",
			changes: []ClubChange{{ChildID: 1, LastName: "Иванов", FirstName: "Петя", Start: "11:00", Name: "Шахматы"}},
			want: []DayClub{
				{ChildID: 1, LastName: "Иванов", FirstName: "Петя", Start: "11:00", Name: "Шахматы"},
				schedule[1], schedule[2],
			},
		},
		{
			name:    "club without a pair is added",
			changes: []ClubChange{{ChildID: 3, LastName: "Сидоров", FirstName: "Коля", Start: "15:00", Name: "Футбол"}},
			want: []DayClub{
				schedule[1], schedule[0], schedule[2],
				{ChildID: 3, LastName: "Сидоров", FirstName: "Коля", Start: "15:00", Name: "Футбол"},
				schedule[3],
			},
		},
		{
			name:    "cancel without a pair changes nothing",
			changes: []ClubChange{{ChildID: 1, Name: "Футбол"}, {ChildID: 3, Name: "Шахматы"}},
			want:    []DayClub{schedule[1], schedule[0], schedule[2], schedule[3]},
		},
		{
			name:    "names are compared exactly",
			changes: []ClubChange{{ChildID: 1, Name: "шахматы"}},
			want:    []DayClub{schedule[1], schedule[0], schedule[2], schedule[3]},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := applyClubChanges(schedule, tt.changes); !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
