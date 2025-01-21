package QCheckWE

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"github.com/go-resty/resty/v2"
)

type WeQuotaChecker struct {
	LNDNumber string
	LNDPass   string
	ACCTID    string
	Session   *resty.Client
}

func NewWeQuotaChecker(landlineNumber, password string) (*WeQuotaChecker, error) {
	if landlineNumber == "" || password == "" {
		return nil, fmt.Errorf("landline number and password are required")
	}

	if !strings.HasPrefix(landlineNumber, "02") || len(landlineNumber) != 10 {
		return nil, fmt.Errorf("invalid landline number format. Must start with 02 and be 10 digits")
	}

	acctID := "FBB" + landlineNumber[1:]

	client := resty.New()
	client.SetBaseURL("https://api-my.te.eg")
	client.SetHeaders(map[string]string{
		"Accept":          "application/json, text/plain, */*",
		"Accept-Language": "en-US,en;q=0.9,ar;q=0.8",
		"Content-Type":    "application/json",
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"channelId":       "702",
		"isCoporate":      "false",
		"isMobile":        "false",
		"isSelfcare":      "true",
		"languageCode":    "en-US",
	})

	return &WeQuotaChecker{
		LNDNumber: landlineNumber,
		LNDPass:   password,
		ACCTID:    acctID,
		Session:   client,
	}, nil
}

func (w *WeQuotaChecker) CheckQuota() (map[string]interface{}, error) {
	authData, err := w.authenticate()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	customer := authData["customer"].(map[string]interface{})
	subscriber := authData["subscriber"].(map[string]interface{})
	token := authData["token"].(string)

	offerId, err := w.getSubscribedOfferings(token)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscribed offerings: %v", err)
	}

	quota, err := w.getQuotaDetails(token, subscriber["subscriberId"].(string), offerId)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota details: %v", err)
	}

	used, ok := quota["used"].(float64)
	if !ok {
		return nil, fmt.Errorf("failed to assert used as float64")
	}

	total, ok := quota["total"].(float64)
	if !ok {
		return nil, fmt.Errorf("failed to assert total as float64")
	}

	usagePrc := (used / total) * 100

	effectiveTimeFloat, ok := quota["effectiveTime"].(float64)
	if !ok {
		effectiveTimeInt, ok := quota["effectiveTime"].(int64)
		if !ok {
			return nil, fmt.Errorf("failed to assert effectiveTime as float64 or int64")
		}
		effectiveTimeFloat = float64(effectiveTimeInt)
	}

	renewedDate := w.tsConv(int64(effectiveTimeFloat), false)[0]

	expireTimeFloat, ok := quota["expireTime"].(float64)
	if !ok {
		expireTimeInt, ok := quota["expireTime"].(int64)
		if !ok {
			return nil, fmt.Errorf("failed to assert expireTime as float64 or int64")
		}
		expireTimeFloat = float64(expireTimeInt)
	}

	expiryDate := w.tsConv(int64(expireTimeFloat), true)

	return map[string]interface{}{
		"name":            customer["custName"].(string),
		"offerName":       quota["offerName"].(string),
		"remaining":       quota["remain"].(float64),
		"total":           total,
		"usagePercentage": fmt.Sprintf("%.2f", usagePrc),
		"renewalDate":     renewedDate,
		"expiryDate":      expiryDate[0],
		"expiryIn":        expiryDate[1],
	}, nil
}

func (w *WeQuotaChecker) tsConv(unixTimestamp int64, returnUntil bool) []string {
	t := time.Unix(unixTimestamp/1000, 0)
	formattedDate := t.Format("02/01/2006 at 03:04 PM")

	dates := []string{formattedDate}

	if returnUntil {
		now := time.Now()
		diff := t.Sub(now)

		if diff.Hours() < 24 {
			hoursLeft := int(diff.Hours())
			dates = append(dates, fmt.Sprintf("in %d hours", hoursLeft))
		} else {
			daysLeft := int(diff.Hours() / 24)
			dates = append(dates, fmt.Sprintf("in %d days", daysLeft))
		}
	}

	return dates
}

func (w *WeQuotaChecker) authenticate() (map[string]interface{}, error) {
	_, err := w.Session.R().
		SetBody(map[string]interface{}{}).
		Post("/echannel/service/besapp/base/rest/busiservice/v1/common/querySysParams")
	if err != nil {
		return nil, fmt.Errorf("failed to query sys params: %v", err)
	}

	authResponse, err := w.Session.R().
		SetBody(map[string]interface{}{
			"acctId":    w.ACCTID,
			"appLocale": "en-US",
			"password":  w.LNDPass,
		}).
		Post("/echannel/service/besapp/base/rest/busiservice/v1/auth/userAuthenticate")
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	var authData map[string]interface{}
	if err := json.Unmarshal(authResponse.Body(), &authData); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %v", err)
	}

	if authData["header"].(map[string]interface{})["retCode"].(string) != "0" {
		return nil, fmt.Errorf("authentication failed: %v", authData["header"].(map[string]interface{})["retMsg"])
	}

	return authData["body"].(map[string]interface{}), nil
}

func (w *WeQuotaChecker) getSubscribedOfferings(token string) (string, error) {
	offersResponse, err := w.Session.R().
		SetBody(map[string]interface{}{
			"msisdn":           w.ACCTID,
			"numberServiceType": "FBB",
			"groupId":          "",
		}).
		SetHeader("csrftoken", token).
		Post("/echannel/service/besapp/base/rest/busiservice/cz/v1/auth/getSubscribedOfferings")
	if err != nil {
		return "", fmt.Errorf("failed to get subscribed offerings: %v", err)
	}

	var offersData map[string]interface{}
	if err := json.Unmarshal(offersResponse.Body(), &offersData); err != nil {
		return "", fmt.Errorf("failed to decode offers response: %v", err)
	}

	if offersData["header"].(map[string]interface{})["retCode"].(string) != "0" {
		return "", fmt.Errorf("failed to get subscribed offerings: %v", offersData["header"].(map[string]interface{})["retMsg"])
	}

	offeringList := offersData["body"].(map[string]interface{})["offeringList"].([]interface{})
	if len(offeringList) == 0 {
		return "", fmt.Errorf("no offerings found")
	}

	return offeringList[0].(map[string]interface{})["mainOfferingId"].(string), nil
}

func (w *WeQuotaChecker) getQuotaDetails(token, subscriberId, offerId string) (map[string]interface{}, error) {
	quotaResponse, err := w.Session.R().
		SetBody(map[string]interface{}{
			"subscriberId": subscriberId,
			"mainOfferId":  offerId,
		}).
		SetHeader("csrftoken", token).
		Post("/echannel/service/besapp/base/rest/busiservice/cz/cbs/bb/queryFreeUnit")
	if err != nil {
		return nil, fmt.Errorf("failed to get quota details: %v", err)
	}

	var quotaData map[string]interface{}
	if err := json.Unmarshal(quotaResponse.Body(), &quotaData); err != nil {
		return nil, fmt.Errorf("failed to decode quota response: %v", err)
	}

	if quotaData["header"].(map[string]interface{})["retCode"].(string) != "0" {
		return nil, fmt.Errorf("failed to get quota details: %v", quotaData["header"].(map[string]interface{})["retMsg"])
	}

	quotaList := quotaData["body"].([]interface{})
	if len(quotaList) == 0 {
		return nil, fmt.Errorf("no quota details found")
	}

	return quotaList[0].(map[string]interface{}), nil
}
