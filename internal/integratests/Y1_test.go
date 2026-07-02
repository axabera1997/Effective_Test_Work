package integratests

import (
	"emobile/internal/models"
	"encoding/json"
	"net/http"

	"github.com/go-resty/resty/v2"
)

var sub = models.Subscription{
	Service_name: "Жейминь Жибао",
	Price:        400,
	//	User_id:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
	Start_date: "02-2020",
	End_date:   "11-2029",
}

func (suite *TS) Test_01() {

	// создаём случайны User_id чтобы не было конфликтов
	ui := suite.uids[0]
	sub1 := sub
	sub1.User_id = ui

	httpc := resty.New().SetBaseURL("http://localhost:8080")

	req := httpc.R().SetHeader("Content-Type", "application/json").SetDoNotParseResponse(false).
		SetBody(sub1)

	resp, err := req.Post("/add")
	suite.Require().NoError(err, "req.Post add)")
	suite.Require().Equal(http.StatusOK, resp.StatusCode())
	suite.Require().JSONEq(`{"Cunt":1, "Name":"Внесено записей"}`, resp.String())

	resp, err = req.Post("/read")
	suite.Require().NoError(err, "req.Post read)")
	suite.Require().Equal(http.StatusOK, resp.StatusCode())

	// размаршаллим полученное. это слайс подписок с одной записью
	subs := []models.Subscription{}
	err = json.Unmarshal([]byte(resp.String()), &subs)
	suite.Require().NoError(err, "bad unmarshal )")

	// запись должна быть одна
	suite.Require().Equal(1, len(subs))

	// проверим пару полей на соответствие
	suite.Require().Equal(subs[0].Service_name, sub1.Service_name)
	suite.Require().Equal(subs[0].User_id, sub1.User_id)

	resp, err = req.Get("/list")
	suite.Require().NoError(err, "req.Post list)")
	suite.Require().Equal(http.StatusOK, resp.StatusCode())

	// размаршаллим полученное.
	// это опять же слайс подписок с одной записью
	// ЕСЛИ обнуляли таблицу в SetupTest !!!
	subs = []models.Subscription{}
	err = json.Unmarshal([]byte(resp.String()), &subs)
	suite.Require().NoError(err, "bad unmarshal )")
	// запись должна быть одна
	suite.Require().Equal(1, len(subs))

	// проверим пару полей на соответствие
	suite.Require().Equal(subs[0].Service_name, sub1.Service_name)
	suite.Require().Equal(subs[0].User_id, sub1.User_id)

	sub2 := sub
	sub2.Start_date = "02-2010"
	sub2.End_date = "02-2019"
	req = httpc.R().SetHeader("Content-Type", "application/json").SetDoNotParseResponse(false).
		SetBody(sub2)
	resp, err = req.Post("/summa")
	suite.Require().NoError(err, "summa")
	suite.Require().Equal(http.StatusOK, resp.StatusCode())

	suite.Require().JSONEq(`{"Cunt":0, "Name":"Нет таких подписок"}`, resp.String())

	sub2.End_date = "02-2020"
	req = httpc.R().SetHeader("Content-Type", "application/json").SetDoNotParseResponse(false).
		SetBody(sub2)
	resp, err = req.Post("/summa")
	suite.Require().NoError(err, "summa")
	suite.Require().Equal(http.StatusOK, resp.StatusCode())

	suite.Require().JSONEq(`{"Cunt":400, "Name":"Сумма подписок"}`, resp.String())

	// внесём подписку в щастливом будущем
	sub3 := sub2
	sub3.Start_date = "02-2040"
	sub3.End_date = "02-2050"
	sub3.Service_name = "Партизан Приморья"
	sub3.User_id = ui
	req = httpc.R().SetHeader("Content-Type", "application/json").SetDoNotParseResponse(false).
		SetBody(sub3)
	resp, err = req.Post("/add")
	suite.Require().NoError(err, "req.Post add)")
	suite.Require().Equal(http.StatusOK, resp.StatusCode())
	suite.Require().JSONEq(`{"Cunt":1, "Name":"Внесено записей"}`, resp.String())

	req = httpc.R().SetHeader("Content-Type", "application/json").SetDoNotParseResponse(false).
		SetBody(sub2)
	resp, err = req.Post("/summa")
	suite.Require().NoError(err, "summa")
	suite.Require().Equal(http.StatusOK, resp.StatusCode())
	// должно быть 400, будущее не в счёт
	suite.Require().JSONEq(`{"Cunt":400, "Name":"Сумма подписок"}`, resp.String())

}
