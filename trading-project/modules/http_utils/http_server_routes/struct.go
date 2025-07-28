package httpserverroutes

type Response_UserAuth struct {
	AccessToken  string `json:"Access_token"`
	RefreshToken string `json:"Refresh_token"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
}

type Jwt_Claims struct {
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	Role       string `json:"role"`
	AccessType string `json:"access_type"`
}

type Item struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}
