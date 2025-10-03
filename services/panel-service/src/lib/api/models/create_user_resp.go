package models

type CreateUserResponse struct {
	Response struct {
		Username        string `json:"username"`
		Status          string `json:"status"`
		ExpireAt        string `json:"expireAt"`
		HwidDeviceLimit int    `json:"hwidDeviceLimit"`
		SubscriptionUrl string `json:"subscriptionUrl"`
	} `json:"response"`
}
