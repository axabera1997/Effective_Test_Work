package handlera

import (
	"encoding/json"
	"net/http"
	"strconv"

	"emobile/internal/config"
	"emobile/internal/models"
)

// ListHandler возвращает список подписок с пагинацией
// @Summary Получение списка подписок
// @Description Возвращает список всех подписок с поддержкой пагинации
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param pageSize query int false "Размер страницы (по умолчанию из конфига)"
// @Success 200 {array} models.Subscription "Список подписок"
// @Success 204 {object} models.RetStruct "Подписки не найдены"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /list [get]
func (db *InterStruct) ListHandler(rwr http.ResponseWriter, req *http.Request) {

	// Получение параметров страницы
	pageStr := req.URL.Query().Get("page")
	pageSizeStr := req.URL.Query().Get("pageSize")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// записей на страницу вывода задаётся в .env  PAGE_SIZE
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = config.Configuration.PageSize
	}
	if pageSize == 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// запрос в БД на получения списка всех подписок
	subs, err := db.Inter.ListSub(req.Context(), pageSize, offset)
	if err != nil {
		http.Error(rwr, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(subs) != 0 {
		rwr.WriteHeader(http.StatusOK)
		models.Logger.Info("Cписок", "подписки", subs)
		json.NewEncoder(rwr).Encode(subs)
		return
	}

	rwr.WriteHeader(http.StatusNoContent)
	models.Logger.Info("Нет записей в подписках")

	ret := models.RetStruct{
		Name: "Нет записей в подписках",
		Cunt: 0,
	}
	json.NewEncoder(rwr).Encode(ret)

}
