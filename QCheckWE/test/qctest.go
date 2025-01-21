package test

import (
	"fmt"
	"QCheckWE"
)

func main() {
	landlineNumber := "0244761737"
	password := "khaled1970"

	checker, err := QCheckWE.NewWeQuotaChecker(landlineNumber, password)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	quotaInfo, err := checker.CheckQuota()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf(`
		Customer: %s
		Plan: %s
		Remaining: %.2f / %.2f (%s%% Used)
		Renewed On: %s
		Expires On: %s (%s)
	`,
		quotaInfo["name"],
		quotaInfo["offerName"],
		quotaInfo["remaining"],
		quotaInfo["total"],
		quotaInfo["usagePercentage"],
		quotaInfo["renewalDate"],
		quotaInfo["expiryDate"],
		quotaInfo["expiryIn"],
	)
}