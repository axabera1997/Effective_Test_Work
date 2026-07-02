package dbase

import (
	"emobile/internal/models"
	"time"

	"github.com/google/uuid"
)

func (suite *TstHand) Test_01AddSubFunc() {

	// Gлобальный sub, template
	subG := models.Subscription{
		Service_name: "Yandex Plus",
		Price:        400,
		User_id:      uuid.NewString(),
		// User_id:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		Start_date: "02-2025",
		End_date:   "11-2025",
	}
	err := MakeTT(&subG)
	suite.Require().NoError(err)

	// трём таблицу передав пустую запись
	// впрочем, там и так нет ничего, на случай,
	// если в дальнейшем добавятся операции перед этим тестом
	_, err = suite.dataBase.DeleteSub(suite.ctx, models.Subscription{})
	suite.Require().NoError(err)

	// число подписок должно стать 0
	subs, err := suite.dataBase.ListSub(suite.ctx, 40, 0)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(subs))

	cTag, err := suite.dataBase.AddSub(suite.ctx, subG)
	suite.Require().NoError(err)
	// EqualValues для унификации, т.к. RowsAffected() int64, а 1 - int
	suite.Require().EqualValues(1, cTag.RowsAffected())

	// структура "плоская", без указателей, поэтому копия независимая
	sub1 := subG
	sub1.Edt = time.Time{}
	sub1.Price = 0
	MakeTT(&sub1)

	// генерируем user_id
	sub1.User_id = uuid.NewString()
	// добавление записи, с иным user_id
	cTag, err = suite.dataBase.AddSub(suite.ctx, sub1)
	suite.Require().NoError(err)
	// 1 - запись добавилась
	suite.Require().EqualValues(1, cTag.RowsAffected())

	// число подписок должно стать 2
	subs, err = suite.dataBase.ListSub(suite.ctx, 40, 0)
	suite.Require().NoError(err)
	suite.Require().EqualValues(2, len(subs))

	sub1.Service_name = "Мурзилка"
	// повторяем добавление записи, с иным Service_name
	cTag, err = suite.dataBase.AddSub(suite.ctx, sub1)
	suite.Require().NoError(err)
	// 1 - запись добавилась
	suite.Require().EqualValues(1, cTag.RowsAffected())

	// число подписок должно стать 3
	subs, err = suite.dataBase.ListSub(suite.ctx, 40, 0)
	suite.Require().NoError(err)
	suite.Require().EqualValues(3, len(subs))

	sub2 := sub1
	sub2.Service_name = "Völkischer Beobachter"
	// удаляем несуществующее ныне
	cTag, err = suite.dataBase.DeleteSub(suite.ctx, sub2)
	suite.Require().NoError(err)
	// 0 - нету такого, вот и не удалилось
	suite.Require().EqualValues(0, cTag.RowsAffected())

	// удаляем самую первую запись - subG
	cTag, err = suite.dataBase.DeleteSub(suite.ctx, subG)
	suite.Require().NoError(err)
	// 1 - норм
	suite.Require().EqualValues(1, cTag.RowsAffected())

	// 3-1 = 2
	subs, err = suite.dataBase.ListSub(suite.ctx, 40, 0)
	suite.Require().NoError(err)
	suite.Require().EqualValues(2, len(subs))

	// подымем Мурзиле цену
	sub1.Price = 777
	cTag, err = suite.dataBase.UpdateSub(suite.ctx, sub1)
	suite.Require().NoError(err)
	// 1 - норм, апгрейд
	suite.Require().EqualValues(1, cTag.RowsAffected())

	// количество не изменилось
	subs, err = suite.dataBase.ListSub(suite.ctx, 40, 0)
	suite.Require().NoError(err)
	suite.Require().EqualValues(2, len(subs))

	// pagesize 1, result 1 must be
	subs, err = suite.dataBase.ListSub(suite.ctx, 1, 0)
	suite.Require().NoError(err)
	suite.Require().EqualValues(1, len(subs))

	// offset 1, result 1 must be - 2-1
	subs, err = suite.dataBase.ListSub(suite.ctx, 40, 1)
	suite.Require().NoError(err)
	suite.Require().EqualValues(1, len(subs))

	// drop table
	_, err = suite.dataBase.DeleteSub(suite.ctx, models.Subscription{})
	suite.Require().NoError(err)

	subP := models.Subscription{
		Service_name: "Чаян",
		Price:        700,
		User_id:      uuid.NewString(),
		// User_id:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		Start_date: "02-2025",
		End_date:   "11-2025",
	}

	err = MakeTT(&subP)
	suite.Require().NoError(err)

	cTag, err = suite.dataBase.AddSub(suite.ctx, subP)
	suite.Require().NoError(err)
	// 1 - запись добавилась
	suite.Require().EqualValues(1, cTag.RowsAffected())

	checkP := models.Subscription{
		Service_name: "Чаян",
		//		User_id:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		Start_date: "02-2025",
		End_date:   "04-2025",
	}

	err = MakeTT(&checkP)
	suite.Require().NoError(err)

	summa, err := suite.dataBase.SumSub(suite.ctx, checkP)
	suite.Require().NoError(err)
	//  02, 03, 04 - 3 months
	suite.Require().EqualValues(int64(700*3), summa)

	// c апреля по июль ещё один подписчик
	subA := models.Subscription{
		Service_name: "Чаян",
		Price:        700,
		User_id:      uuid.NewString(),
		Start_date:   "04-25",
		End_date:     "07-2025",
	}

	err = MakeTT(&subA)
	suite.Require().NoError(err)

	cTag, err = suite.dataBase.AddSub(suite.ctx, subA)
	suite.Require().NoError(err)
	// 1 - запись добавилась
	suite.Require().EqualValues(1, cTag.RowsAffected())

	summa, err = suite.dataBase.SumSub(suite.ctx, checkP)
	suite.Require().NoError(err)
	//  02, 03, 04 - 3 months плюс один месяц по subA
	suite.Require().EqualValues(int64(700*4), summa)

}
