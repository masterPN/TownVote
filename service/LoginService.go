package service

type LoginService interface {
	LoginUser(id_no string, id_laserCode string) bool
}

type loginInfomation struct {
	id_no        string
	id_laserCode string
}

func StaticLoginService() LoginService {
	return &loginInfomation{
		id_no:        "1234567890123",
		id_laserCode: "JT123",
	}
}

func (info *loginInfomation) LoginUser(id_no string, id_laserCode string) bool {
	return info.id_no == id_no && info.id_laserCode == id_laserCode
}
