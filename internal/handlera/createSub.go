package handlera

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"emobile/internal/dbase"
	"emobile/internal/models"

	"github.com/google/uuid"
)

// DBPinger godoc
// @Summary Database health check
// @Description Checks if database connection is alive
// @Produce json
// @Success 200 {object} map[string]string "Database is reachable"
// @Failure 500 {object} map[string]string "Database connection error"
// @Router / [get]
func (db *InterStruct) DBPinger(rwr http.ResponseWriter, req *http.Request) {

	err := dbase.Ping(req.Context())
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	fmt.Fprintf(rwr, `{"status":"StatusOK"}`)
}

// CreateHandler создает новую подписку
// @Summary Создание подписки
// @Description Создает новую подписку с валидацией данных
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.Subscription true "Данные подписки"
// @Success 200 {object} models.RetStruct "Подписка успешно создана"
// @Failure 400 {string} string "Неверные данные (отсутствуют обязательные поля, неверный формат UUID, даты)"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /add [post]
func (db *InterStruct) CreateHandler(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	sub := models.Subscription{}
	err := json.NewDecoder(req.Body).Decode(&sub)
	if err != nil {
		models.Logger.Error("json", "NewDecoder", err)
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(err)
		return
	}

	if sub.Service_name == "" {
		models.Logger.Error("no service name")
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(errors.New("no service name"))
		return
	}
	if sub.Price == 0 {
		models.Logger.Error("no price")
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(errors.New("no price"))
		return
	}
	if sub.User_id == "" {
		models.Logger.Error("no user_id")
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(errors.New("no user_id"))
		return
	}
	_, err = uuid.Parse(sub.User_id)
	if err != nil {
		models.Logger.Error("bad user_id, not UUID format")
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(errors.New("bad user_id, not UUID format"))
		return
	}

	if sub.Sdt.IsZero() {
		models.Logger.Error("no start date")
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(errors.New("no start date"))
		return
	}
	// если при непустой конечной дате она раньше начальной
	if !sub.Edt.IsZero() && sub.Edt.Before(sub.Sdt) {
		models.Logger.Error("end date before start")
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(errors.New("end date before start"))
		return
	}
	// если конечная дата подписки не задана - устанавливаем её в бесконечность
	// таким образом, все поля подписки при внесении в базу будут заполнены ненулевыми значениями
	if sub.Edt.IsZero() {
		sub.Edt = time.Date(9999, time.December, 31, 23, 59, 59, 999999999, time.UTC)
	}

	cTag, err := db.Inter.AddSub(req.Context(), sub)
	if err != nil {
		models.Logger.Error("AddSub table method", "", err)
		rwr.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rwr).Encode(err)
		return
	}

	ret := models.RetStruct{
		Name: "Подписка создана",
		Cunt: cTag.RowsAffected(),
	}

	models.Logger.Info("Подписка", "создана", sub)

	rwr.WriteHeader(http.StatusOK)
	json.NewEncoder(rwr).Encode(ret)

}
