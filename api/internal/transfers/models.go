package transfers

type Transfer struct {
	SourceID        int64  `db:"source_id" json:"source_id"`
	SourceName      string `json:"source"`
	DestinationID   int64  `db:"destination_id" json:"destination_id"`
	DestinationName string `json:"destination"`
	Amount          int64  `db:"amount" json:"amount"`
	ExecutionDate   string `db:"execution_date" json:"execution_date"`
}
