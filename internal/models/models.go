package models

type (
	User struct {
		Login     string
		Password  string
		SecretKey string
	}

	Registration struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	Order struct {
		Order   string `json:"order"`
		Status  string `json:"status"`
		Accrual int    `json:"accrual"`
	}

	Orders struct {
		Token    string
		NumOrder int
	}

	UserOrders struct {
		NumOrders int     `json:"number"`
		Accrual   float64 `json:"accrual"`
		Status    string  `json:"status"`
		Upload    string  `json:"uploaded_at"`
	}

	Balance struct {
		Accrual   float64 `json:"accrual"`
		Withdrawn float64 `json:"current"`
	}

	Withdrawn struct {
		Sum   float64 `json:"sum"`
		Order int     `json:"order"`
	}

	AllWithdrawn struct {
		Order     int     `json:"order"`
		Sum       float64 `json:"sum"`
		Processed string  `json:"processed_at"`
	}
)
