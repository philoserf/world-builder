package stars

// StarTypeRow is one row of the WBH p. 15 Star Type Determination table.
//
// Default column is Type; rolls of 2 redirect to Special, rolls of 12
// redirect to Hot. Class III+ rolls go to Giants. The Unusual and
// Peculiar columns are entered only when the procedure says so.
type StarTypeRow struct {
	Type, Hot, Special, Unusual, Giants, Peculiar string
}

// StarTypeDetermination is the WBH p. 15 Star Type Determination table.
var StarTypeDetermination = map[int]StarTypeRow{
	2:  {Type: "Special", Hot: "A", Special: "Class VI", Unusual: "Peculiar", Giants: "Class III", Peculiar: "Black Hole"},
	3:  {Type: "M", Hot: "A", Special: "Class VI", Unusual: "Class VI", Giants: "Class III", Peculiar: "Pulsar"},
	4:  {Type: "M", Hot: "A", Special: "Class VI", Unusual: "Class IV", Giants: "Class III", Peculiar: "Neutron Star"},
	5:  {Type: "M", Hot: "A", Special: "Class VI", Unusual: "BD", Giants: "Class III", Peculiar: "Nebula"},
	6:  {Type: "M", Hot: "A", Special: "Class IV", Unusual: "BD", Giants: "Class III", Peculiar: "Nebula"},
	7:  {Type: "K", Hot: "A", Special: "Class IV", Unusual: "BD", Giants: "Class III", Peculiar: "Protostar"},
	8:  {Type: "K", Hot: "A", Special: "Class IV", Unusual: "D", Giants: "Class III", Peculiar: "Protostar"},
	9:  {Type: "G", Hot: "A", Special: "Class III", Unusual: "D", Giants: "Class II", Peculiar: "Protostar"},
	10: {Type: "G", Hot: "B", Special: "Class III", Unusual: "D", Giants: "Class II", Peculiar: "Star Cluster"},
	11: {Type: "F", Hot: "B", Special: "Giants", Unusual: "Class III", Giants: "Class Ib", Peculiar: "Anomaly"},
	12: {Type: "Hot", Hot: "O", Special: "Giants", Unusual: "Giants", Giants: "Class Ia", Peculiar: "Anomaly"},
}
