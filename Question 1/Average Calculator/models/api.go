package models

type NumbersResponse struct {
	Numbers []int `json:"numbers"`
}

type APIResponse struct {
	WindowPrevState []int   `json:"windowPrevState"`
	WindowCurrState []int   `json:"windowCurrState"`
	Numbers         []int   `json:"numbers"`
	Avg             float64 `json:"avg"`
}
