package client

import (
	"fmt"
	"time"
)

type MealTransactionRequest struct {
	IIN       string `json:"iin"`
	Date      string `json:"date"`
	SchoolBin string `json:"school_bin"`
}

type MealTransactionResponse struct {
	Success   bool   `json:"success"`
	ErrorCode int    `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

func SendMealTransaction(iin string, eventTime time.Time, schoolBin string) error {
	fmt.Printf("[MOCK] Отправка в Social Wallet: IIN=%s, BIN=%s, Date=%s\n",
		iin, schoolBin, eventTime.UTC().Format("2006-01-02 15:04:05"))

	return nil
}

// package client

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"time"
// )

// type MealTransactionRequest struct {
// 	IIN       string `json:"iin"`
// 	Date      string `json:"date"`
// 	SchoolBin string `json:"school_bin"`
// }

// type MealTransactionResponse struct {
// 	Success   bool   `json:"success"`
// 	ErrorCode int    `json:"error_code,omitempty"`
// 	ErrorMsg  string `json:"error_msg,omitempty"`
// }

// func SendMealTransaction(iin string, eventTime time.Time, schoolBin string) error {
// 	url := "https://api.socialwallet.kz/api/v1/sdu/meal/transaction"

// 	reqBody := MealTransactionRequest{
// 		IIN:       iin,
// 		Date:      eventTime.UTC().Format("2006-01-02 15:04:05"),
// 		SchoolBin: schoolBin,
// 	}

// 	data, err := json.Marshal(reqBody)
// 	if err != nil {
// 		return err
// 	}

// 	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(data))
// 	if err != nil {
// 		return err
// 	}

// 	httpReq.SetBasicAuth("login", "password") 
// 	httpReq.Header.Set("Content-Type", "application/json")

// 	client := http.Client{Timeout: 5 * time.Second}
// 	resp, err := client.Do(httpReq)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("HTTP %d", resp.StatusCode)
// 	}

// 	var response MealTransactionResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
// 		return err
// 	}

// 	if !response.Success {
// 		return fmt.Errorf("SocialWallet error %d: %s", response.ErrorCode, response.ErrorMsg)
// 	}

// 	return nil
// }
