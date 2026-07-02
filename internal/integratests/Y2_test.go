package integratests

import (
	"emobile/internal/models"
	"encoding/json"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func (suite *TS) Test_02() {
	tests := []struct {
		name   string
		hand   string
		sub    models.Subscription
		status int
		noErr  bool
		summ   int64
	}{
		{
			name:   "Drop table",
			hand:   "/delete",
			sub:    models.Subscription{},
			noErr:  true,
			status: http.StatusOK,
		},
		{
			name: "List empty table",
			hand: "/list",
			// sub:    models.Subscription{},
			noErr:  true,
			status: http.StatusNoContent,
		},
		{
			name: "Add OK record",
			hand: "/add",
			sub: models.Subscription{
				Service_name: "Yandex Plus",
				Price:        400,
				User_id:      suite.uids[0],
				Start_date:   "02-2025",
				End_date:     "11-2025",
			},
			noErr:  true,
			status: http.StatusOK,
		},
		{
			name:   "List table with 1 record",
			hand:   "/list",
			noErr:  true,
			status: http.StatusOK,
		},
		{
			name: "Get summ for 2 months",
			hand: "/summa",
			sub: models.Subscription{
				Service_name: "Yandex Plus",
				Price:        400,
				User_id:      suite.uids[0],
				Start_date:   "10-2025",
				End_date:     "11-2025",
			},
			noErr:  true,
			status: http.StatusOK,
			summ:   800,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var resp *resty.Response
			var err error
			httpc := resty.New().SetBaseURL(suite.host)
			req := httpc.R().SetHeader("Content-Type", "application/json").SetDoNotParseResponse(false).
				SetBody(tt.sub)

			switch tt.hand {

			case "/delete":
				resp, err = req.Delete(tt.hand)

			case "/list":
				resp, err = req.Get(tt.hand)

			case "/add":
				resp, err = req.Post(tt.hand)

			case "/summa":
				resp, err = req.Post(tt.hand)
				suite.Require().NoError(err)

				ret := models.RetStruct{}
				err = json.Unmarshal([]byte(resp.String()), &ret)

				suite.Require().NoError(err, "bad unmarshal )")
				suite.Require().Equal(tt.summ, ret.Cunt)

			default:
				return
			}

			if tt.noErr {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
			suite.Require().Equal(tt.status, resp.StatusCode())

		})
	}
}
