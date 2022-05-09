package dto

type LoginCredentials struct {
	Id_no        string `form:"id_no"`
	Id_laserCode string `form:"id_laserCode"`
}
