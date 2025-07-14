package apiserver

type response struct {
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

type weatherRecordDay struct {
	DayID       string               `json:"day_id"`
	Country     string               `json:"country"`
	Timezone    string               `json:"timezone"`
	Date        string               `json:"date"`
	Conditions  string               `json:"conditions"`
	Description string               `json:"description"`
	Hours       []weatherRecordHours `json:"hours"`
}

type weatherRecordHours struct {
	DayID      string   `json:"day_id"`
	Time       string   `json:"time"`
	Conditions string   `json:"conditions"`
	Stations   []string `json:"stations"`
}
