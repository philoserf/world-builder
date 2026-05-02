package stars

import "testing"

func TestStarTypeDetermination_Complete(t *testing.T) {
	for r := 2; r <= 12; r++ {
		row, ok := StarTypeDetermination[r]
		if !ok {
			t.Fatalf("missing row %d", r)
		}
		if row.Type == "" || row.Hot == "" || row.Special == "" ||
			row.Unusual == "" || row.Giants == "" || row.Peculiar == "" {
			t.Fatalf("row %d has empty cell: %+v", r, row)
		}
	}
}

func TestStarTypeDetermination_KnownCells(t *testing.T) {
	// WBH p. 15 spot checks.
	checks := []struct {
		row  int
		col  string
		want string
	}{
		{7, "Type", "K"},
		{7, "Hot", "A"},
		{2, "Type", "Special"},
		{2, "Peculiar", "Black Hole"},
		{12, "Type", "Hot"},
		{12, "Hot", "O"},
		{12, "Giants", "Class Ia"},
		{11, "Type", "F"},
		{11, "Peculiar", "Anomaly"},
	}
	for _, c := range checks {
		t.Run(c.col, func(t *testing.T) {
			row := StarTypeDetermination[c.row]
			var got string
			switch c.col {
			case "Type":
				got = row.Type
			case "Hot":
				got = row.Hot
			case "Special":
				got = row.Special
			case "Unusual":
				got = row.Unusual
			case "Giants":
				got = row.Giants
			case "Peculiar":
				got = row.Peculiar
			}
			if got != c.want {
				t.Fatalf("row %d %s = %q, want %q", c.row, c.col, got, c.want)
			}
		})
	}
}
