package handlera

import (
	"emobile/internal/dbase"
	"emobile/internal/models"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

func (suite *TstHand) Test_03UpdateSub() {
	// update запись по запросу subForUpdate
	subForUpdate := models.Subscription{
		Service_name: "Yandex Plus", //
		Price:        666,
		User_id:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		Start_date:   "08-08",
		End_date:     "02-22",
	}

	subM, err := json.Marshal(subForUpdate)
	suite.Require().NoError(err)

	requestBody := strings.NewReader(string(subM))

	request := httptest.NewRequest(http.MethodPut, "/update", requestBody)

	// Создание ResponseRecorder
	response := httptest.NewRecorder()
	// вызов хандлера
	suite.db.UpdateHandler(response, request)

	res := response.Result()
	defer res.Body.Close()

	// HTTP put UPDATE должен вернуть http.StatusOK
	suite.Require().Equal(http.StatusOK, res.StatusCode)

	// Составляем запрос READ для чтения только что UPDATEd записи
	subReadUpdated := models.Subscription{
		Service_name: "Yandex Plus",
		User_id:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
	}

	subM, err = json.Marshal(subReadUpdated)
	suite.Require().NoError(err)

	requestBody = strings.NewReader(string(subM))

	// Создание HTTP POST-запроса на чтение записи
	request = httptest.NewRequest(http.MethodPost, "/read", requestBody)

	// Установка заголовков
	request.Header.Set("Content-Type", "application/json")

	// Создание ResponseRecorder
	response = httptest.NewRecorder()
	// вызов хандлера
	suite.db.ReadHandler(response, request)

	res = response.Result()
	defer res.Body.Close()

	// http.StatusOK should be
	suite.Require().Equal(http.StatusOK, res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	suite.Require().NoError(err)

	// размаршалливаем список
	subs := []models.Subscription{}
	err = json.Unmarshal(resBody, &subs)
	suite.Require().NoError(err)
	// должна быть всего одна запись
	suite.Require().Equal(1, len(subs))

	// убеждаемся, что запись обновилась
	suite.Require().Equal(subForUpdate.Service_name, subs[0].Service_name)
	suite.Require().Equal(subForUpdate.User_id, subs[0].User_id)
	suite.Require().Equal(subForUpdate.Price, subs[0].Price)
	// suite.Require().EqualValues(subForUpdate.Start_date.(time.Time), subs[0].Start_date.(time.Time))
	// suite.Require().EqualValues(subForUpdate.End_date.(time.Time), subs[0].End_date.(time.Time))
}

func (suite *TstHand) Test_04SetSumma() {

	subForUpdate := models.Subscription{
		Service_name: "Yandex Plus", //
		Price:        666,
		User_id:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		Start_date:   "08-08",
		End_date:     "02-22",
	}

	subM, err := json.Marshal(subForUpdate)
	suite.Require().NoError(err)

	requestBody := strings.NewReader(string(subM))

	request := httptest.NewRequest(http.MethodPut, "/update", requestBody)

	// Создание ResponseRecorder
	response := httptest.NewRecorder()
	// вызов хандлера
	suite.db.UpdateHandler(response, request)

	res := response.Result()
	defer res.Body.Close()

	// HTTP put UPDATE должен вернуть http.StatusOK
	suite.Require().Equal(http.StatusOK, res.StatusCode)

}

func (suite *TstHand) Test_05DeleteAllSubs() {
	// пустая структура.
	subForDelete := models.Subscription{}

	subM, err := json.Marshal(subForDelete)
	suite.Require().NoError(err)

	requestBody := strings.NewReader(string(subM))

	request := httptest.NewRequest(http.MethodPut, "/delete", requestBody)

	// Создание ResponseRecorder
	response := httptest.NewRecorder()
	// вызов хандлера
	suite.db.DeleteHandler(response, request)

	res := response.Result()
	defer res.Body.Close()

	// HTTP put UPDATE должен вернуть http.StatusOK
	suite.Require().Equal(http.StatusOK, res.StatusCode)

	// проверяем на обнуление после DELETE с пустой структурой
	db, err := dbase.NewPostgresPool(suite.ctx, models.DSN)
	suite.Require().NoError(err)
	defer db.DB.Close()

	// запрос в БД на получения списка всех подписок
	subs, err := db.ListSub(suite.ctx, 40, 0)
	suite.Require().NoError(err)
	// список должен быть пустой
	suite.Require().Equal(0, len(subs))

}
