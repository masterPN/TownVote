package service

type LoginService interface {
	LoginUser(id_no string, id_name string) bool
}

type loginInfomation struct {
	id_no   string
	id_name string
}

func StaticLoginService() LoginService {
	return &loginInfomation{
		id_no:   "9",
		id_name: "goline",
	}
}

func (info *loginInfomation) LoginUser(id_no string, id_name string) bool {
	return info.id_no == id_no && info.id_name == id_name
}
