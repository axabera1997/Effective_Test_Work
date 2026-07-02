package handlera

import (
	"emobile/internal/models"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

// type errMessage struct {
// 	Message   string `json:"Message"`
// 	Detail    string `json:"Detail"`
// 	TableName string `json:"TableName"`
// }

func (suite *TstHand) Test_02ReadSub() {

	// Völkischer Beobachter   Avanti

	sub := models.Subscription{
		Service_name: "Yandex Plus",
		Price:        400,
		User_id:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		//	Start_date:   "01-02-2025",
		//	End_date:     "11-2025",
	}

	tests := []struct {
		name string
		//	dbEndPoint string
		sub     models.Subscription
		status  int
		records int
	}{
		{
			name: "start date less than recorded",
			sub: func() models.Subscription {
				s := sub
				s.Start_date = "01-25"
				return s
			}(),
			status:  http.StatusNoContent,
			records: 0,
		},
		{
			name:    "Normaldu",
			sub:     sub,
			status:  http.StatusOK,
			records: 1,
		},
		{
			name: "Normaldu ZERO price",
			sub: func() models.Subscription {
				s := sub
				s.Price = 0
				return s
			}(),
			status:  http.StatusOK,
			records: 1,
		},
		{
			name: "Normaldu with start date",
			sub: func() models.Subscription {
				s := sub
				s.Start_date = "02-25"
				return s
			}(),
			status:  http.StatusOK,
			records: 1,
		},
		{
			name: "end date more than recorded",
			sub: func() models.Subscription {
				s := sub
				s.End_date = "01-35"
				return s
			}(),
			status:  http.StatusNoContent,
			records: 0,
		},
	}

	for _, tt := range tests {

		suite.Run(tt.name, func() {

			subM, err := json.Marshal(tt.sub)
			suite.Require().NoError(err)

			requestBody := strings.NewReader(string(subM))

			// Создание POST-запроса
			request := httptest.NewRequest(http.MethodPost, "/read", requestBody)

			// Установка заголовков
			request.Header.Set("Content-Type", "application/json")

			// Создание ResponseRecorder
			response := httptest.NewRecorder()
			// вызов хандлера
			suite.db.ReadHandler(response, request)

			res := response.Result()
			defer res.Body.Close()

			// Assert чтобы выполнилось сравнение tt.reply, string(resBody)
			suite.Require().Equal(tt.status, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			suite.Require().NoError(err)

			if tt.records != 0 {
				// размаршалливаем список подписок
				subs := []models.Subscription{}
				err = json.Unmarshal(resBody, &subs)
				suite.Require().NoError(err)
				// должно быть 2 записи
				suite.Require().Equal(tt.records, len(subs))
			}

		})
	}

}
