package handlera

import (
	"emobile/internal/models"
	"encoding/json"
	"errors"
	"net/http"
)

// ReadHandler выполняет поиск подписок по критериям
// @Summary Поиск подписок
// @Description Возвращает подписки, соответствующие заданным критериям поиска
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param criteria body models.Subscription true "Критерии поиска подписок"
// @Success 200 {array} models.Subscription "Найденные подписки"
// @Success 204 {object} models.RetStruct "Подписки не найдены"
// @Failure 400 {string} string "Неверный формат данных"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /read [post]
func (db *InterStruct) ReadHandler(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	readSub := models.Subscription{}
	err := json.NewDecoder(req.Body).Decode(&readSub)
	if err != nil {
		http.Error(rwr, err.Error(), http.StatusBadRequest)
		return
	}

	subs, err := db.Inter.ReadSub(req.Context(), readSub)
	if err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(subs) != 0 {
		rwr.WriteHeader(http.StatusOK)
		models.Logger.Info("Найдено", "подписок", len(subs))
		json.NewEncoder(rwr).Encode(subs)
		return
	}

	rwr.WriteHeader(http.StatusNoContent)

	models.Logger.Info("Read - Не найдено записей")
	ret := models.RetStruct{
		Name: "Не найдено записей, удовлетворяющих запросу",
		Cunt: 0,
	}
	json.NewEncoder(rwr).Encode(ret)

}

// UpdateHandler обновляет данные подписки
// @Summary Обновление подписки
// @Description Обновляет данные подписки по заданным критериям (обязательные поля: service_name и user_id)
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.Subscription true "Данные для обновления подписки"
// @Success 200 {object} models.RetStruct "Подписка успешно обновлена"
// @Success 204 {object} models.RetStruct "Подписка не найдена"
// @Failure 400 {string} string "Неверные данные (отсутствуют обязательные поля)"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /update [put]
func (db *InterStruct) UpdateHandler(rwr http.ResponseWriter, req *http.Request) {

	rwr.Header().Set("Content-Type", "application/json")

	readSub := models.Subscription{}
	err := json.NewDecoder(req.Body).Decode(&readSub)
	if err != nil {
		http.Error(rwr, err.Error(), http.StatusBadRequest)
		return
	}

	// в запросе Update обязятельныц поля Service_name и User_id
	if readSub.Service_name == "" {
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(errors.New("no service name"))
		return
	}
	if readSub.User_id == "" {
		rwr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rwr).Encode(errors.New("no user_id"))
		return
	}

	cTag, err := db.Inter.UpdateSub(req.Context(), readSub)
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rwr).Encode(err)
		return
	}

	ret := models.RetStruct{
		Name: "Обновлено записей",
		Cunt: cTag.RowsAffected(),
	}
	if cTag.RowsAffected() == 0 {
		ret.Name = "Не найдено записей, удовлетворяющих запросу"
		rwr.WriteHeader(http.StatusNoContent)
		models.Logger.Info(ret.Name)
	} else {
		rwr.WriteHeader(http.StatusOK)
		models.Logger.Info("UPDATE", "", ret)
	}

	json.NewEncoder(rwr).Encode(ret)
}
